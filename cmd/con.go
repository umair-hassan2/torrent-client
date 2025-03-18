package torrent

import "io"

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

func NewHandShake(infoHash, peerId [20]byte) *HandShake {
	return &HandShake{
		infoHash: infoHash,
		peerId:   peerId,
		pstr:     "BitTorrent protocol",
	}
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

func ReadHandShake(conn io.Reader) (*HandShake, error) {
	// read length of protocol identifier
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(conn, lengthBuf)
	if err != nil {
		return nil, err
	}

	// read protocol identifier string
	pstrBuf := make([]byte, int(lengthBuf[0]))
	_, err = io.ReadFull(conn, pstrBuf)
	if err != nil {
		return nil, err
	}

	// read remaining data from connection
	handShakeBuf := make([]byte, 48)
	_, err = io.ReadFull(conn, handShakeBuf)
	if err != nil {
		return nil, err
	}

	var infoHash, peerId [20]byte
	copy(infoHash[:], handShakeBuf[8:8+20])
	copy(peerId[:], handShakeBuf[8+20:8+20+20])
	handShake := HandShake{
		infoHash: infoHash,
		peerId:   peerId,
		pstr:     string(pstrBuf),
	}

	return &handShake, nil
}
