package main

import (
	"fmt"
	"os"

	torrent "github.com/umair-hassan2/torrent-client/cmd"
)

func main() {
	torrentFilePath := "tests/test_files/sample_2.torrent"
	file, err := os.Open(torrentFilePath)
	if err != nil {
		fmt.Println("ERROR OCCURED")
		fmt.Println(err)
	}

	dt, err := torrent.DecodeFile(file)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(dt.Announce)
}
