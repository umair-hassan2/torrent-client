package types

import "net"

type Peer struct {
	IP   net.IP
	Port int
	ID   string
}

type TorrentFile struct {
	Announce    string
	InfoHash    []byte
	PieceHashes [][]byte
	PieceLength int
	Length      int
	Name        string
}

type PieceWork struct {
	Index  int
	Hash   [20]byte
	Length int
}

type PieceResult struct {
	Index int
	Data  []byte
}

type DownloadingState struct {
	Downloaded int
	Retries    int
	Result     []byte
}
