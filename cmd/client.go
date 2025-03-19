package torrent

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

type Client struct {
	Con      net.Conn
	peerId   [20]byte
	infoHash [20]byte
	peer     Peer
	bitField BitField
	Choked   bool
}

func StartHandShake(con net.Conn, infoHash, peerId [20]byte) error {
	con.SetDeadline(time.Now().Add(3 * time.Second))
	defer con.SetDeadline(time.Time{})
	// we send hand shake request with payload having info hash, peer id, pstr, pstr length, reserved bytes
	handShake := NewHandShake(infoHash, peerId)

	// send handshake request
	_, err := con.Write(handShake.Serialize())
	if err != nil {
		return err
	}

	// read from connection
	handShakeResponse, err := ReadHandShake(con)

	if err != nil {
		return err
	}

	// verify integrity of info hash
	if !bytes.Equal(handShake.infoHash[:], handShakeResponse.infoHash[:]) {
		return fmt.Errorf("info hash mismatch")
	}

	return nil
}

func readBitFieldMessage(conn net.Conn) (BitField, error) {
	// set time outs
	conn.SetDeadline(time.Now().Add(time.Second * 3))
	defer conn.SetDeadline(time.Time{})

	// Read message from the connection
	message, err := Read(conn)
	if err != nil {
		return nil, err
	}

	if message == nil {
		return nil, fmt.Errorf("Expected a bit field message but received null message")
	}

	// verify that this message is a bitfield message
	if message.Id == MsgBitfield {
		return nil, fmt.Errorf("Expected a bit field message but received %s", findMessagebyId(message.Id))
	}

	return message.Payload, nil
}

func New(peer Peer, peerId, infoHash [20]byte) (*Client, error) {
	// open a tcp connection
	con, err := net.DialTimeout("tcp", peer.String(), 3*time.Millisecond)
	if err != nil {
		return nil, err
	}

	err = StartHandShake(con, infoHash, peerId)
	if err != nil {
		con.Close()
		return nil, err
	}

	// read bitfield message
	bitFieldMessage, err := readBitFieldMessage(con)
	if err != nil {
		con.Close()
		return nil, err
	}

	newClient := Client{
		peerId:   peerId,
		peer:     peer,
		infoHash: infoHash,
		bitField: bitFieldMessage,
		Con:      con,
		Choked:   true, // peer is choked by default
	}
	return &newClient, nil
}

// there are basic 9 types of messages
func (c *Client) SendHave(pieceIndex int) error {
	message := FormatHaveMessage(pieceIndex)
	_, err := c.Con.Write(message.Serialize())
	return err
}

func (c *Client) SendChoke() error {
	message := Message{
		Id: MsgChoke,
	}
	_, err := c.Con.Write(message.Serialize())
	return err
}

func (c *Client) SendUnChoke() error {
	message := Message{
		Id: MsgUnChoke,
	}
	_, err := c.Con.Write(message.Serialize())
	return err
}

func (c *Client) SendInterested() error {
	message := Message{
		Id: MsgInterested,
	}
	_, err := c.Con.Write(message.Serialize())
	return err
}

func (c *Client) SendNotInterested() error {
	message := Message{
		Id: MsgNotInterested,
	}
	_, err := c.Con.Write(message.Serialize())
	return err
}

func (c *Client) SendRequest(pieceIndex, begin, length int) error {
	message := FormatRequestMessage(pieceIndex, begin, length)
	_, err := c.Con.Write(message.Serialize())
	return err
}
