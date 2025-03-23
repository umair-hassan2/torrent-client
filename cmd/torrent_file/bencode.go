package torrent_file

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"

	bencode "github.com/jackpal/bencode-go"
	"github.com/umair-hassan2/torrent-client/cmd/common"
	"github.com/umair-hassan2/torrent-client/pkg/types"
)

type bencodeInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

type bencodeTorrentFile struct {
	Announce string      `bencode:"announce"`
	Comment  string      `bencode:"comment"`
	Info     bencodeInfo `bencode:"info"`
}

type BencodeCompactTrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    []byte `bencode:"peers"`
}

// decode .torrent file
func DecodeFile(reader io.Reader) (*bencodeTorrentFile, error) {
	torrentFile := bencodeTorrentFile{}
	err := bencode.Unmarshal(reader, &torrentFile)
	if err != nil {
		return nil, err
	}

	return &torrentFile, nil
}

// returns list of hashes of pieces in give .torrent file
func (btf *bencodeTorrentFile) GetHashPieces() (hashPieces [][20]byte) {
	pieces := []byte(btf.Info.Pieces)
	for i := 0; i < len(btf.Info.Pieces); i += 20 {
		hashPiece := pieces[i : i+20]
		hashPieces = append(hashPieces, [20]byte(hashPiece))
	}
	return hashPieces
}

// parse tracker response
func ParseTrackerResponse(response string) (BencodeCompactTrackerResponse, error) {
	trackerResponse := BencodeCompactTrackerResponse{}
	err := bencode.Unmarshal(strings.NewReader(response), &trackerResponse)
	if err != nil {
		return BencodeCompactTrackerResponse{}, err
	}

	return trackerResponse, nil
}

// load list of remote peers from compact tracker response
func (btr *BencodeCompactTrackerResponse) GetRemotePeers() ([]*types.Peer, error) {
	peerEntrySize := 6
	totalPeers := len(btr.Peers) / peerEntrySize
	if len(btr.Peers)%totalPeers != 0 {
		return nil, fmt.Errorf("peer size is invalid")
	}

	remotePeers := make([]*types.Peer, totalPeers)

	// first 4 bytes = ip address
	// last 2 bytes = port number
	for i := 0; i < totalPeers; i++ {
		offset := i * peerEntrySize
		ipBytes := btr.Peers[offset : offset+4]
		portBytes := btr.Peers[offset+4 : offset+6]
		remotePeers[i] = common.NewPeer("", net.IP(ipBytes), int(binary.BigEndian.Uint16(portBytes)))
	}
	return remotePeers, nil
}
