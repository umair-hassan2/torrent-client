package common

import (
	"net"
	"strconv"

	"github.com/umair-hassan2/torrent-client/pkg/types"
)

type NotImplementedError struct{}

func (n *NotImplementedError) Error() string {
	return "Not Implemented"
}

func NewPeer(peerId string, ip net.IP, port int) *types.Peer {
	return &types.Peer{
		ID:   peerId,
		Port: port,
		IP:   ip,
	}
}

func PeerAdress(peer types.Peer) string {
	return peer.IP.String() + ":" + strconv.Itoa(peer.Port)
}

func CalculatePieceBounds(pieceIndex, pieceLength, fileLength int) (int, int) {
	startIndex := pieceIndex * pieceLength
	lastIndex := min((pieceIndex+1)*pieceLength, fileLength)
	return startIndex, lastIndex
}
