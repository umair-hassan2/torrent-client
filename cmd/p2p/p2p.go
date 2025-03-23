package p2p

import (
	"net"
	"os"

	"github.com/umair-hassan2/torrent-client/cmd/common"
	"github.com/umair-hassan2/torrent-client/cmd/torrent"
	"github.com/umair-hassan2/torrent-client/cmd/torrent_file"
)

const ClientId = "zero-net"

func Begin(fileName string) {
	// TODO: UI interface to upload file
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	fileContent, err := torrent_file.DecodeFile(file)
	if err != nil {
		panic(err)
	}

	torrentFile := torrent_file.FromBencodeToTorrentFile(fileContent)
	currentPeer := common.NewPeer("", net.IP("127.0.0.1"), 3000)
	torrent := torrent.New(*currentPeer, torrentFile)
	go torrent.Start()
}
