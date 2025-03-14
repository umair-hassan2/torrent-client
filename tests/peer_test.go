package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	torrent "github.com/umair-hassan2/torrent-client/cmd"
)

func TestGeneraePeerId(t *testing.T) {
	t.Run("size of peer id should be 20 bytes", func(t *testing.T) {
		peerId, err := torrent.RandomPeerId()
		assert.NoError(t, err, "should not raise error in peer id generation")
		assert.NotNil(t, peerId, "Peer id should not be nil")
		assert.Equal(t, len([]byte(peerId)), 20, "peer id should have size 20 bytes")
	})

	t.Run("peer id should have client id as a prefix", func(t *testing.T) {
		peerId, err := torrent.RandomPeerId()
		clientId := torrent.ClientId
		assert.NoError(t, err, "should not raise error in peer id generation")
		assert.NotNil(t, peerId, "Peer id should not be nil")
		assert.True(t, len(peerId) >= len(clientId), "peer id should be at least as long as the prefix")
		assert.Equal(t, string(peerId[:len(clientId)]), clientId, fmt.Sprintf("peer id should have %s as a prefix", clientId))
	})
}
