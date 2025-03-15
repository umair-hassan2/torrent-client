package torrent

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"

	"github.com/jackpal/bencode-go"
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

type bencodeCompactTrackerResponse struct {
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
func ParseTrackerResponse(response string) (bencodeCompactTrackerResponse, error) {
	trackerResponse := bencodeCompactTrackerResponse{}
	err := bencode.Unmarshal(strings.NewReader(response), &trackerResponse)
	if err != nil {
		return bencodeCompactTrackerResponse{}, err
	}

	return trackerResponse, nil
}

// load list of remote peers from compact tracker response
func (btr *bencodeCompactTrackerResponse) GetRemotePeers() ([]*Peer, error) {
	peerEntrySize := 6
	totalPeers := len(btr.Peers) / peerEntrySize
	if len(btr.Peers)%totalPeers != 0 {
		return nil, errors.New("Invalid Peer bin size")
	}

	remotePeers := make([]*Peer, totalPeers)

	// first 4 bytes = ip address
	// last 2 bytes = port number
	for i := 0; i < totalPeers; i++ {
		offset := i * peerEntrySize
		ipBytes := btr.Peers[offset : offset+4]
		portBytes := btr.Peers[offset+4 : offset+6]
		remotePeers[i] = &Peer{
			peerId: "",
			ip:     net.IP(ipBytes),
			port:   int(binary.BigEndian.Uint16(portBytes)),
		}
	}
	return remotePeers, nil
}
