package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/umair-hassan2/torrent-client/cmd/common"
	"github.com/umair-hassan2/torrent-client/cmd/message"
	"github.com/umair-hassan2/torrent-client/pkg/types"
)

type Client struct {
	Con      net.Conn
	peerId   [20]byte
	infoHash [20]byte
	peer     types.Peer
	bitField message.BitField
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

func readBitFieldMessage(conn net.Conn) (message.BitField, error) {
	// set time outs
	conn.SetDeadline(time.Now().Add(time.Second * 3))
	defer conn.SetDeadline(time.Time{})

	// Read message from the connection
	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}

	if msg == nil {
		return nil, fmt.Errorf("expected a bit field message but received null message")
	}

	// verify that this message is a bitfield message
	if msg.Id == message.MsgBitfield {
		return nil, fmt.Errorf("expected a bit field message but received %s", message.FindMessagebyId(msg.Id))
	}

	return msg.Payload, nil
}

func New(peer types.Peer, peerId, infoHash [20]byte) (*Client, error) {
	// open a tcp connection
	con, err := net.DialTimeout("tcp", common.PeerAdress(peer), 3*time.Millisecond)
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
	message := message.FormatHaveMessage(pieceIndex)
	_, err := c.Con.Write(message.Serialize())
	return err
}

func (c *Client) SendChoke() error {
	message := message.Message{
		Id: message.MsgChoke,
	}
	_, err := c.Con.Write(message.Serialize())
	return err
}

func (c *Client) SendUnChoke() error {
	message := message.Message{
		Id: message.MsgUnChoke,
	}
	_, err := c.Con.Write(message.Serialize())
	return err
}

func (c *Client) SendInterested() error {
	message := message.Message{
		Id: message.MsgInterested,
	}
	_, err := c.Con.Write(message.Serialize())
	return err
}

func (c *Client) SendNotInterested() error {
	message := message.Message{
		Id: message.MsgNotInterested,
	}
	_, err := c.Con.Write(message.Serialize())
	return err
}

func (c *Client) SendRequest(pieceIndex, begin, length int) error {
	message := message.FormatRequestMessage(pieceIndex, begin, length)
	_, err := c.Con.Write(message.Serialize())
	return err
}
