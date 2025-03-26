package torrent

import (
	"crypto/sha1"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/umair-hassan2/torrent-client/cmd/client"
	"github.com/umair-hassan2/torrent-client/cmd/common"
	"github.com/umair-hassan2/torrent-client/cmd/message"
	"github.com/umair-hassan2/torrent-client/cmd/torrent_file"
	"github.com/umair-hassan2/torrent-client/cmd/tracker"
	"github.com/umair-hassan2/torrent-client/pkg/types"
)

const (
	MAX_BLOCK_SIZE                   = 100 * 1024 * 8
	MAX_ALLOWED_RETRIES              = 5
	MAX_ALLOWED_DOWNLOAD_CONNECTIONS = 40
	MAX_ALLOWED_UPLOAD_CONNECTIONS   = 40
)

var open_download_con int
var open_upload_con int

// Torrent represents one torrent file
// It is responsible to perform every step to download it's specific file
type Torrent struct {
	Url         string
	InfoHash    [20]byte
	PieceLength int
	Length      int
	PieceHashes [][20]byte
	currentPeer *types.Peer
	remotePeers []*types.Peer
}

// Torrent is created from a torrent file data
func New(peer types.Peer, torrentFile *torrent_file.TorrentFile) *Torrent {
	return &Torrent{
		Url:         torrentFile.Announce,
		InfoHash:    torrentFile.InfoHash,
		PieceLength: torrentFile.PieceLength,
		Length:      torrentFile.Length,
		PieceHashes: torrentFile.PieceHashes,
		currentPeer: &peer,
	}
}

func (t *Torrent) ReadRemotePeerMessage(c *client.Client, peer *types.Peer, state *types.DownloadingState) error {
	msg, err := message.Read(c.Con)
	if err != nil {
		return err
	}
	switch msg.Id {
	case message.MsgUnChoke:
		c.Choked = false
	case message.MsgChoke:
		c.Choked = true
	case message.MsgHave:
		index := message.ParseHaveMessage(msg)
		c.BitField.SetPiece(index)
	case message.MsgPiece:
		pieceMessage, err := message.ParsePieceMessage(msg)
		if err != nil {
			return err
		}
		// TODO: check if length of block data is equal to what we requested earlier in our request
		state.Result = append(state.Result, pieceMessage.BlockData...)
	default:
		doNothing()
	}

	return nil
}

func doNothing() {

}

// TODO: Implement request pipelining to have at most 5-10 unfulfilled request in the piepline
// TODO: Can we make it adaptable to improve performance ?
func (t *Torrent) downloadAPiece(c *client.Client, piece *types.PieceWork, peer types.Peer) (*types.PieceResult, error) {
	// set timer for connection
	// reset the timer when this function returns
	// always try to download 100KB block from remote peer
	c.Con.SetDeadline(time.Now().Add(time.Second * 30))
	defer c.Con.SetDeadline(time.Time{})

	state := types.DownloadingState{}

	// keep iterating until entire piece is downloaded
	for state.Downloaded < piece.Length {
		// check if peer is unchoked
		if !c.Choked {
			block := min(MAX_BLOCK_SIZE, piece.Length-state.Downloaded)
			err := c.SendRequest(piece.Index, state.Downloaded+1, block)
			if err != nil {
				return nil, err
			}
			state.Downloaded += block
		}

		err := t.ReadRemotePeerMessage(c, &peer, &state)
		if err != nil {
			return nil, err
		}
	}

	return &types.PieceResult{
		Index: piece.Index,
		Data:  state.Result,
	}, nil
}

func (t *Torrent) downloadFromPeer(peer types.Peer, workerChan *chan types.PieceWork, resultChan *chan types.PieceResult) error {
	// check if this new connection is allowed as per settings
	if open_download_con >= MAX_ALLOWED_DOWNLOAD_CONNECTIONS {
		// return and don't perform furthur processing
		return fmt.Errorf("attempt to open more than %v connections", MAX_ALLOWED_DOWNLOAD_CONNECTIONS)
	}

	var peerId [20]byte
	copy(peerId[:], peer.ID[:20])
	c, err := client.New(peer, peerId, t.InfoHash)
	if err != nil {
		return err
	}
	common.AddTo(&open_download_con, 1)
	defer c.Con.Close()
	defer common.DecFrom(&open_download_con, 1)

	// client maps current peer to one remote peer
	c.SendUnChoke()
	c.SendInterested()

	for piece := range *workerChan {
		// client does not have this piece so put it back to worker chan and we will try to download again in future (from this peer or some other peer)
		if !c.BitField.HasPiece(piece.Index) {
			(*workerChan) <- piece
			continue
		}

		// attempt to donwload this piece
		downloaded, err := t.downloadAPiece(c, &piece, peer)
		if err != nil {
			log.Default().Printf("Failed to download piece %q, from peer %q", piece, peer)
			// TODO: Add retries and after max retries throw an error for this piece but keep other pieces
			(*workerChan) <- piece
			continue
		}

		// perform integrity check of downloaded piece
		if sha1.Sum(downloaded.Data) != t.PieceHashes[piece.Index] {
			return fmt.Errorf("downloaded piece %q failed integriy check from remote peer %q", downloaded, peer)
		}

		(*resultChan) <- *downloaded
	}

	return nil
}

func (t *Torrent) Download() {
	workerChan := make(chan types.PieceWork, len(t.PieceHashes))
	resultChan := make(chan types.PieceResult, len(t.PieceHashes))

	for index, piece := range t.PieceHashes {
		work := types.PieceWork{
			Index:  index,
			Hash:   piece,
			Length: t.PieceLength,
		}

		workerChan <- work
	}

	for _, peer := range t.remotePeers {
		go t.downloadFromPeer(*peer, &workerChan, &resultChan)
	}

	// bytes which are dowonloaded so far
	downloadedBytes := 0
	percentage := 0
	//gather downloaded piece and convert into a single file
	file := make([]byte, t.Length)
	for downloadedPiece := range resultChan {
		downloadedBytes += len(downloadedPiece.Data)
		percentage = (t.Length / downloadedBytes) * 100
		fmt.Printf("%v percent downloaded, bytes = %v", percentage, downloadedBytes)

		start, end := common.CalculatePieceBounds(downloadedPiece.Index, t.PieceLength, t.Length)
		copy(file[start:end], downloadedPiece.Data[start:end])
	}
	fmt.Println("FILE DOWNLOADED")
	fmt.Println(string(file))
}

// entry point for a torrent communication
func (t *Torrent) Start() {
	open_download_con = 0
	// get peer list from tracker server
	trackerUrl, err := t.BuildTrackerUrl()
	if err != nil {
		panic(err)
	}

	trackerResponse, err := tracker.GetTrackerResponse(trackerUrl)
	if err != nil {
		panic(err)
	}

	// after every response.Interval seconds ... get fresh list of remote peers from tracker server
	// YET TO IMPLEMENT ^^^
	copy(t.remotePeers[:], trackerResponse.Peers[:])

	// Once you get remote peers start downloading from these peers
	t.Download()
}

func (t *Torrent) BuildTrackerUrl() (string, error) {
	trackerUrl, err := url.Parse(t.Url)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"peer_id":    []string{string(t.currentPeer.ID[:])},
		"info_hash":  []string{string(t.InfoHash[:])},
		"port":       []string{strconv.Itoa(t.currentPeer.Port)},
		"left":       []string{"100"},
		"downloaded": []string{"0"},
		"uploaded":   []string{"0"},
		"compact":    []string{"1"}, // tracket returns packages string instead of bencoded hash - https://www.bittorrent.org/beps/bep_0023.html
	}
	trackerUrl.RawQuery = params.Encode()
	return trackerUrl.String(), nil
}
