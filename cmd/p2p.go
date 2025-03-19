package torrent

const ClientId = "zero-net"

// download pieces
type Torrent TorrentFile

type PieceWork struct {
	index  int
	hash   []byte
	length int
}

type PieceResult struct {
	index  int
	result []byte
}

func (t *Torrent) calculatePieceSize(index int) int {
	start, end := t.getPieceRange(index)
	return end - start
}

func (t *Torrent) getPieceRange(index int) (int, int) {
	start := index * t.Length
	end := (index + 1) * t.Length
	if end > t.Length {
		end = t.Length
	}
	return start, end
}

func (t *Torrent) DownloadSinglePiece(pieceWork *PieceWork, workerChan chan *PieceWork, resultChan chan *PieceResult) {
	// download piece and add it to result channel
	// Grab a piece from the worker channel
	// create a client with a connection to a peer
	// but connection is choked by default

	// Yet to be Implemented
}

func (t *Torrent) Download() ([]byte, error) {
	// create two channel - 1 channel to grab the piece which are yet to be downloaded - 1 channel to store downloaded pieces
	// calculate index and length of all file pieces and store as PieceWork
	// Attemp to download pieces one by one in go routines
	// iterate result channel and collate these pieces into a single file
	workerChan := make(chan *PieceWork, len(t.PieceHashes))
	resultChan := make(chan *PieceResult, len(t.PieceHashes))
	for index, piece := range t.PieceHashes {
		len := t.calculatePieceSize(index)
		pw := PieceWork{
			index:  index,
			hash:   piece[:],
			length: len,
		}
		workerChan <- &pw
	}

	// download piece
	for piece := range workerChan {
		go t.DownloadSinglePiece(piece, workerChan, resultChan)
	}

	result := make([]byte, t.Length)

	// collate download pieces
	for downloaded := range resultChan {
		start, end := t.getPieceRange(downloaded.index)
		copy(result[start:end], downloaded.result)
	}

	return result, nil
}
