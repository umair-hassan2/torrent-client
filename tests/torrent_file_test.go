package tests

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	torrent "github.com/umair-hassan2/torrent-client/cmd"
)

func TestDecodeFile(t *testing.T) {
	testFile := "test_files/sample_2.torrent"

	// Call the DecodeFile function
	file, err := os.Open(testFile)
	assert.NoError(t, err, "Failed to open test torrent file")
	defer file.Close()

	torrentFile, err := torrent.DecodeFile(file)

	// Verify the results - Check if parsing succeeded
	assert.NoError(t, err, "DecodeFile returned an error")
	assert.NotNil(t, torrentFile, "Torrent should not be nil")

	assert.Equal(t, "udp://tracker.openbittorrent.com:80/announce", torrentFile.Announce)
	assert.Equal(t, "bbb_sunflower_1080p_60fps_normal.mp4", torrentFile.Info.Name)
	assert.Equal(t, int(355856562), torrentFile.Info.Length)   // Updated to cast to int
	assert.Equal(t, int(524288), torrentFile.Info.PieceLength) // Updated to cast to int

}

func TestFromBencodeToTorrentFile(t *testing.T) {
	testFile := "test_files/sample_2.torrent"
	file, err := os.Open(testFile)
	assert.NoError(t, err, "Failed to open test torrent file")
	defer file.Close()

	btf, err := torrent.DecodeFile(file)
	assert.NoError(t, err, "DecodeFile returned an error")
	assert.NotNil(t, btf, "Torrent should not be nil")

	torrentFile := torrent.FromBencodeToTorrentFile(btf)
	assert.Equal(t, "udp://tracker.openbittorrent.com:80/announce", torrentFile.Announce)
	assert.Equal(t, "bbb_sunflower_1080p_60fps_normal.mp4", torrentFile.Name)
	assert.Equal(t, int(355856562), torrentFile.Length)   // Updated to cast to int
	assert.Equal(t, int(524288), torrentFile.PieceLength) // Updated to cast to int
	assert.Equal(t, int(679), len(torrentFile.PieceHashes))
}

func TestBuildTrackerUrl(t *testing.T) {
	infoHash := "aaaaaaaaaaaaaaaaaaaa"
	infoHashBytes := []byte(infoHash)
	sampleTorrentFile := &torrent.TorrentFile{
		Announce:    "localhost/announce",
		Length:      120,
		Name:        "sample_1.txt",
		Comment:     "file sent from MacOs",
		PieceLength: 20,
		InfoHash:    [20]byte(infoHashBytes),
	}

	t.Run("should create tracker url", func(t *testing.T) {
		peerId := [20]byte([]byte("aaaaaaaaaaaaaaaaaaaa"))
		port := 6881
		actualTrackerUrl, err := sampleTorrentFile.BuildTrackerUrl([20]byte(peerId), port)

		assert.NoError(t, err, "should not raise erorr in tracker url creation")

		expectTrackerUrl := "localhost/announce?compact=1&downloaded=0&info_hash=aaaaaaaaaaaaaaaaaaaa&left=100&peer_id=aaaaaaaaaaaaaaaaaaaa&port=6881&uploaded=0"
		assert.Equal(t, expectTrackerUrl, actualTrackerUrl, "tracker url mismatch")
	})
}
