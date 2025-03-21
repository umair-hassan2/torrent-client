package torrent

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
)

type Peer struct {
	peerId string
	ip     net.IP
	port   int
}

func NewPeer(peerId string, ip net.IP, port int) *Peer {
	return &Peer{
		peerId: peerId,
		ip:     ip,
		port:   port,
	}
}

func (p Peer) String() string {
	return fmt.Sprintf("%s:%d", p.ip.String(), p.port)
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
