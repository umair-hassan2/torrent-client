package client

import (
	"fmt"
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umair-hassan2/torrent-client/pkg/types"
)

// MockTCPServer represents a mock TCP server for testing
type MockTCPServer struct {
	listener net.Listener
	done     chan bool
}

func NewMockServer(t *testing.T) *MockTCPServer {
	l, err := net.Listen("tcp", "127.0.0.1:0") // 0 means random available port
	require.NoError(t, err)

	mock := &MockTCPServer{
		listener: l,
		done:     make(chan bool),
	}

	// Start mock server
	go mock.start()
	return mock
}

func (m *MockTCPServer) start() {
	for {
		conn, err := m.listener.Accept()
		if err != nil {
			select {
			case <-m.done:
				return
			default:
				continue
			}
		}
		go m.handleConnection(conn)
	}
}

func (m *MockTCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Mock handshake response
	response := make([]byte, 68)
	copy(response[0:], []byte{19})                      // pstrlen
	copy(response[1:20], []byte("BitTorrent protocol")) // pstr
	// Leave reserved bytes as zeros
	// Copy the rest of the handshake as zeros (info_hash and peer_id)

	conn.Write(response)

	// Mock bitfield message
	bitfieldMsg := []byte{0, 0, 0, 2, 5, 0xFF} // Length prefix + ID + payload
	conn.Write(bitfieldMsg)
}

func (m *MockTCPServer) Close() {
	close(m.done)
	m.listener.Close()
}

func (m *MockTCPServer) Addr() string {
	return m.listener.Addr().String()
}

func TestConnectionWithMockServer(t *testing.T) {
	mockServer := NewMockServer(t)
	defer mockServer.Close()

	host, portStr, err := net.SplitHostPort(mockServer.Addr())
	require.NoError(t, err)

	peer := types.Peer{
		IP:   net.ParseIP(host),
		Port: atoi(portStr),
		ID:   "test-peer-id",
	}

	var peerId, infoHash [20]byte
	copy(peerId[:], []byte("test-peer-id-123456"))
	copy(infoHash[:], []byte("test-info-hash-1234"))

	client, err := New(peer, peerId, infoHash)
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.True(t, client.Choked)
	assert.NotNil(t, client.BitField)
}

func TestInvalidConnection(t *testing.T) {
	tests := []struct {
		name    string
		peer    types.Peer
		wantErr bool
	}{
		{
			name: "invalid IP",
			peer: types.Peer{
				IP:   net.ParseIP("256.256.256.256"),
				Port: 6881,
				ID:   "test-peer-id",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			peer: types.Peer{
				IP:   net.ParseIP("127.0.0.1"),
				Port: -1,
				ID:   "test-peer-id",
			},
			wantErr: true,
		},
		{
			name: "non-existent host",
			peer: types.Peer{
				IP:   net.ParseIP("10.0.0.1"),
				Port: 6881,
				ID:   "test-peer-id",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var peerId, infoHash [20]byte
			client, err := New(tt.peer, peerId, infoHash)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestMessageExchange(t *testing.T) {
	mockServer := NewMockServer(t)
	defer mockServer.Close()

	host, portStr, err := net.SplitHostPort(mockServer.Addr())
	require.NoError(t, err)

	peer := types.Peer{
		IP:   net.ParseIP(host),
		Port: atoi(portStr),
		ID:   "test-peer-id",
	}

	var peerId, infoHash [20]byte
	client, err := New(peer, peerId, infoHash)
	require.NoError(t, err)
	require.NotNil(t, client)

	tests := []struct {
		name    string
		fn      func() error
		wantErr bool
	}{
		{"SendInterested", client.SendInterested, false},
		{"SendNotInterested", client.SendNotInterested, false},
		{"SendChoke", client.SendChoke, false},
		{"SendUnChoke", client.SendUnChoke, false},
		{"SendHave", func() error { return client.SendHave(1) }, false},
		{"SendRequest", func() error { return client.SendRequest(1, 0, 16384) }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to convert string port to int
func atoi(s string) int {
	port, _ := net.LookupPort("tcp", s)
	return port
}

func TestConnectionSetup(t *testing.T) {
	peer := types.Peer{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 6881,
		ID:   "test-peer-id",
	}

	conn, err := New(peer, [20]byte{}, [20]byte{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if conn == nil {
		t.Fatal("Expected a valid connection, got nil")
	}

	// Test edge case: invalid IP
	invalidPeer := types.Peer{
		IP:   net.ParseIP("256.256.256.256"), // Invalid IP
		Port: 6881,
		ID:   "invalid-peer-id",
	}

	_, err = New(invalidPeer, [20]byte{}, [20]byte{})
	if err == nil {
		t.Error("Expected error for invalid IP, got none")
	}
}

func TestConnectionClose(t *testing.T) {
	peer := types.Peer{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 6881,
		ID:   "test-peer-id",
	}

	conn, err := New(peer, [20]byte{}, [20]byte{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	fmt.Println(conn)
}

func StartMockTcpServer(host string, port int) {
	net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)))

}
