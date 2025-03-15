package torrent

type NotImplementedError struct{}

func (n *NotImplementedError) Error() string {
	return "Not Implemented"
}
