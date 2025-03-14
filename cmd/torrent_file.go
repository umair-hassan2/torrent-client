package torrent

import (
	"net/url"
	"strconv"
)

type TorrentFile struct {
	Announce    string
	Comment     string
	Length      int
	Name        string
	PieceLength int
	InfoHash    [20]byte // SHA-1 hash of bencoded torrent file - fixed length of 20 bytes
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

func (t *TorrentFile) BuildTrackerUrl(peerId [20]byte, portNumber int) (string, error) {
	trackerUrl, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"peer_id":    []string{string(peerId[:])},
		"info_hash":  []string{string(t.InfoHash[:])},
		"port":       []string{strconv.Itoa(portNumber)},
		"left":       []string{"100"},
		"downloaded": []string{"0"},
		"uploaded":   []string{"0"},
		"compact":    []string{"1"}, // tracket returns packages string instead of bencoded hash - https://www.bittorrent.org/beps/bep_0023.html
	}
	trackerUrl.RawQuery = params.Encode()
	return trackerUrl.String(), nil
}
