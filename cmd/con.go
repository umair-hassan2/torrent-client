package torrent

import (
	"net"
	"time"
)

type Con struct {
	timeOut int
	tp      string
}

type TcpCon struct {
	Con
}
type UdpCon struct {
	Con
}

type HandShake struct {
	infoHash [20]byte
	peerId   [20]byte
	pstr     string // BitTorrent protocol
}

func (con *TcpCon) Dial(remoteAddr string) (net.Conn, error) {
	return net.DialTimeout(con.tp, remoteAddr, time.Duration(con.timeOut)*time.Second)
}

func CreateTcpCon(timeOut int) *TcpCon {
	return &TcpCon{Con{timeOut, "tcp"}}
}

func CreateUdpCon(timeOut int) *UdpCon {
	return &UdpCon{Con{timeOut, "udp"}}
}

// build a handshake buffer
func (h *HandShake) Serialize() []byte {
	var reservedBytes [8]byte
	buf := make([]byte, len(h.pstr)+8+1+20+20)
	// length of protocol identifier
	buf[0] = byte(len(h.pstr))
	//protocol identifier string
	idx := 1
	idx += copy(buf[idx:], h.pstr)
	// reserved bytes - currently keeping it 0
	idx += copy(buf[idx:], reservedBytes[:])
	// info hash
	idx += copy(buf[idx:], h.infoHash[:])
	// peer id
	copy(buf[idx:], h.peerId[:])
	return buf
}

func DeSerializeHandshake(buf []byte) HandShake {
	pstrLen := int(buf[0])
	pstr := string(buf[1 : pstrLen+1])
	infoHash := [20]byte{}
	copy(infoHash[:], buf[pstrLen+1+8:pstrLen+1+8+20])
	peerId := [20]byte{}
	copy(peerId[:], buf[pstrLen+1+8+20:])
	return HandShake{infoHash, peerId, pstr}
}
