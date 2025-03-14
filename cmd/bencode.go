package torrent

import (
	"io"

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
