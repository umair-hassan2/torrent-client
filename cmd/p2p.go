package torrent

const ClientId = "zero-net"

// download pieces
type Torrent TorrentFile

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

func (t *Torrent) DownloadSinglePiece(pieceWork PieceWork, pieceChan chan *PieceWork, resultChan chan *PieceResult) {
	// download piece and add it to result channel
}

func (t *Torrent) Download() ([]byte, error) {
	pieceChan := make(chan *PieceWork, len(t.PieceHashes))
	resultChan := make(chan *PieceResult, len(t.PieceHashes))
	// download all pieces one by one concurrently
	for index, piece := range t.PieceHashes {
		len := t.calculatePieceSize(index)
		pieceWork := PieceWork{
			index:  index,
			length: len,
			hash:   piece[:],
		}
		pieceChan <- &pieceWork
	}

	for piece := range pieceChan {
		// concurrently download all pieces and collate them into result chan
		go t.DownloadSinglePiece(*piece, pieceChan, resultChan)
	}

	finalResult := make([]byte, t.Length)
	// collate downloaded piece
	for piece := range resultChan {
		start, end := t.getPieceRange(piece.index)
		copy(finalResult[start:end], piece.result)
	}

	close(pieceChan)

	return finalResult, nil
}
