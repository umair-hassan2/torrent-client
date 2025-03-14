package torrent

import (
	"crypto/rand"
	"encoding/base64"
)

type Peer struct {
	peerId string
	ip     string
	port   int
}

func NewPeer(peerId string, ip string, port int) *Peer {
	return &Peer{
		peerId: peerId,
		ip:     ip,
		port:   port,
	}
}

func RandomPeerId() ([]byte, error) {
	peerId := make([]byte, 12)
	_, err := rand.Read(peerId)
	if err != nil {
		return []byte{}, err
	}
	encoded := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(peerId)
	if len(encoded) > 12 {
		encoded = encoded[:12]
	}
	encoded = ClientId + encoded
	return []byte(encoded), nil
}
