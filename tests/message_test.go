package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	torrent "github.com/umair-hassan2/torrent-client/cmd"
)

func TestBitFieldHasPiece(t *testing.T) {
	// Create test cases
	testCases := []struct {
		name       string
		bitfield   torrent.BitField
		pieceIndex int
		expected   bool
	}{
		{
			name:       "First bit set",
			bitfield:   torrent.BitField{0b10000000}, // Binary: 10000000
			pieceIndex: 0,
			expected:   true,
		},
		{
			name:       "First bit not set",
			bitfield:   torrent.BitField{0b01000000}, // Binary: 01000000
			pieceIndex: 0,
			expected:   false,
		},
		{
			name:       "Middle bit set",
			bitfield:   torrent.BitField{0b00100000}, // Binary: 00100000
			pieceIndex: 2,
			expected:   true,
		},
		{
			name:       "Last bit of first byte set",
			bitfield:   torrent.BitField{0b00000001}, // Binary: 00000001
			pieceIndex: 7,
			expected:   true,
		},
		{
			name:       "First bit of second byte set",
			bitfield:   torrent.BitField{0b00000000, 0b10000000}, // Binary: 00000000 10000000
			pieceIndex: 8,
			expected:   true,
		},
		{
			name:       "Piece index out of range",
			bitfield:   torrent.BitField{0b11111111}, // Binary: 11111111
			pieceIndex: 8,                            // Only have bits 0-7
			expected:   false,
		},
		{
			name:       "All bits set",
			bitfield:   torrent.BitField{0b11111111}, // Binary: 11111111
			pieceIndex: 5,
			expected:   true,
		},
		{
			name:       "No bits set",
			bitfield:   torrent.BitField{0b00000000}, // Binary: 00000000
			pieceIndex: 3,
			expected:   false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.bitfield.HasPiece(tc.pieceIndex)
			assert.Equal(t, tc.expected, result,
				"BitField.HasPiece(%d) = %v; want %v",
				tc.pieceIndex, result, tc.expected)
		})
	}
}

func TestBitFieldSetPiece(t *testing.T) {
	// Create test cases
	testCases := []struct {
		name             string
		initialBitfield  torrent.BitField
		pieceIndex       int
		expectedBitfield torrent.BitField
		description      string
	}{
		{
			name:             "Set first bit in empty field",
			initialBitfield:  torrent.BitField{0b00000000},
			pieceIndex:       0,
			expectedBitfield: torrent.BitField{0b10000000},
			description:      "Setting bit 0 in an empty byte should result in 10000000",
		},
		{
			name:             "Set last bit in empty field",
			initialBitfield:  torrent.BitField{0b00000000},
			pieceIndex:       7,
			expectedBitfield: torrent.BitField{0b00000001},
			description:      "Setting bit 7 in an empty byte should result in 00000001",
		},
		{
			name:             "Set middle bit in empty field",
			initialBitfield:  torrent.BitField{0b00000000},
			pieceIndex:       3,
			expectedBitfield: torrent.BitField{0b00010000},
			description:      "Setting bit 3 in an empty byte should result in 00010000",
		},
		{
			name:             "Set bit that's already set",
			initialBitfield:  torrent.BitField{0b10000000},
			pieceIndex:       0,
			expectedBitfield: torrent.BitField{0b10000000},
			description:      "Setting a bit that's already set should not change the value",
		},
		{
			name:             "Set bit in second byte",
			initialBitfield:  torrent.BitField{0b00000000, 0b00000000},
			pieceIndex:       8,
			expectedBitfield: torrent.BitField{0b00000000, 0b10000000},
			description:      "Setting bit 8 should change the first bit in the second byte",
		},
		{
			name:             "Set bit in partially filled field",
			initialBitfield:  torrent.BitField{0b10100000},
			pieceIndex:       3, // Changed from 2 to 3 (fourth bit from left)
			expectedBitfield: torrent.BitField{0b10110000},
			description:      "Setting bit 3 in a partially filled byte should result in 10110000",
		},
		{
			name:             "Set bit out of range",
			initialBitfield:  torrent.BitField{0b00000000},
			pieceIndex:       8,
			expectedBitfield: torrent.BitField{0b00000000},
			description:      "Setting a bit beyond the range should leave the bitfield unchanged",
		},
		{
			name:             "Set bit in multi-byte field",
			initialBitfield:  torrent.BitField{0b11111111, 0b00000000, 0b00000000},
			pieceIndex:       15,
			expectedBitfield: torrent.BitField{0b11111111, 0b00000001, 0b00000000},
			description:      "Setting bit 15 should set the last bit in the second byte",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Make a copy of the initial bitfield for testing to avoid modifying the test data
			bitfield := make(torrent.BitField, len(tc.initialBitfield))
			copy(bitfield, tc.initialBitfield)

			// Call SetPiece
			bitfield.SetPiece(tc.pieceIndex)

			// Verify the result
			assert.Equal(t, tc.expectedBitfield, bitfield,
				"BitField.SetPiece(%d) did not set the expected bit. %s",
				tc.pieceIndex, tc.description)

			// Verify that HasPiece now returns true for this piece (if within range)
			if tc.pieceIndex < len(tc.initialBitfield)*8 {
				assert.True(t, bitfield.HasPiece(tc.pieceIndex),
					"After SetPiece(%d), HasPiece(%d) should return true",
					tc.pieceIndex, tc.pieceIndex)
			}
		})
	}
}
