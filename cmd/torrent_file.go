package torrent

type TorrentFile struct {
	Announce    string
	Comment     string
	Length      int
	Name        string
	PieceLength int
	InfoHash    [20]byte
	PieceHashes [][20]byte
}

func FromBencodeToTorrentFile(bencodeTorrentFile *bencodeTorrentFile) *TorrentFile {
	torrentFile := &TorrentFile{
		Announce:    bencodeTorrentFile.Announce,
		Length:      bencodeTorrentFile.Info.Length,
		Name:        bencodeTorrentFile.Info.Name,
		Comment:     bencodeTorrentFile.Comment,
		PieceLength: bencodeTorrentFile.Info.PieceLength,
		PieceHashes: bencodeTorrentFile.GetHashPieces(),
	}

	return torrentFile
}
