package torrent

import (
	"encoding/binary"
	"io"
)

const (
	MsgChoke         uint8 = 0
	MsgUnChoke       uint8 = 1
	MsgInterested    uint8 = 2
	MsgNotInterested uint8 = 3
	MsgHave          uint8 = 4
	MsgBitfield      uint8 = 5
	MsgRequest       uint8 = 6
	MsgPiece         uint8 = 7
	MsgCancel        uint8 = 8
)

// bit torrent message has three main parts
// 1. length of message - 4 bytes
// 2. message id - 1 byte
// 3. payload - variable length
type Message struct {
	Payload []byte
	Id      uint8
	Length  uint32
}

// bit field messages is an array of bytes
type BitField []byte

func (m *Message) Serialize() []byte {
	msgLen := uint32(len(m.Payload) + 1)
	buf := make([]byte, msgLen+4)
	binary.BigEndian.PutUint32(buf[0:], msgLen)
	buf[4] = byte(m.Id)
	copy(buf[5:], m.Payload)
	return buf
}

// read message from stream
func Read(stream *io.Reader) (*Message, error) {
	var length uint32
	err := binary.Read(*stream, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(*stream, buf)
	if err != nil {
		return &Message{}, err
	}

	return &Message{
		Length:  length,
		Id:      uint8(buf[0]),
		Payload: buf[1:],
	}, nil
}

// check if bit field has certain piece
func (bf *BitField) HasPiece(pieceIndex int) bool {
	byteIndex := pieceIndex / 8
	bitIndex := pieceIndex % 8
	if byteIndex >= len(*bf) || bitIndex > 7 {
		return false
	}
	ok := (1 & ((*bf)[byteIndex] >> (7 - bitIndex))) == 1
	return ok
}

func (bf *BitField) SetPiece(pieceIndex int) {
	byteIndex := pieceIndex / 8
	bitIndex := pieceIndex % 8
	if byteIndex >= len(*bf) || bitIndex > 7 {
		return
	}

	(*bf)[byteIndex] |= 1 << (7 - bitIndex)
}
