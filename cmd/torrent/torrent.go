package torrent

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/umair-hassan2/torrent-client/cmd/common"
	"github.com/umair-hassan2/torrent-client/cmd/torrent_file"
	"github.com/umair-hassan2/torrent-client/cmd/tracker"
	"github.com/umair-hassan2/torrent-client/pkg/types"
)

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

func downloadAPiece(piece types.PieceWork, workerChan chan types.PieceWork, resultChan chan types.PieceResult) {
	panic((&common.NotImplementedError{}).Error())
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

	for piece := range workerChan {
		go downloadAPiece(piece, workerChan, resultChan)
	}

	//gather downloaded piece and convert into a single file
	file := make([]byte, t.Length)
	for downloadedPiece := range resultChan {
		start, end := common.CalculatePieceBounds(downloadedPiece.Index, t.PieceLength, t.Length)
		copy(file[start:end], downloadedPiece.Data[start:end])
	}
	fmt.Println("FILE DOWNLOADED")
	fmt.Println(string(file))
}

// entry point for a torrent communication
func (t *Torrent) Start() {
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
