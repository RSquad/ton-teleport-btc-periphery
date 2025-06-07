package utils

import (
	"bytes"
	"testing"
)

func TestBytesPadTo(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		size     int
		expected []byte
	}{
		{
			name:     "pad short input with zeros",
			input:    []byte{1, 2, 3},
			size:     5,
			expected: []byte{0, 0, 1, 2, 3},
		},
		{
			name:     "input same size as target",
			input:    []byte{1, 2, 3, 4, 5},
			size:     5,
			expected: []byte{1, 2, 3, 4, 5},
		},
		{
			name:     "truncate input longer than target",
			input:    []byte{1, 2, 3, 4, 5, 6, 7},
			size:     4,
			expected: []byte{1, 2, 3, 4},
		},
		{
			name:     "empty input",
			input:    []byte{},
			size:     3,
			expected: []byte{0, 0, 0},
		},
		{
			name:     "size zero with non-empty input",
			input:    []byte{1, 2, 3},
			size:     0,
			expected: []byte{},
		},
		{
			name:     "size zero with empty input",
			input:    []byte{},
			size:     0,
			expected: []byte{},
		},
		{
			name:     "single byte input padded",
			input:    []byte{42},
			size:     4,
			expected: []byte{0, 0, 0, 42},
		},
		{
			name:     "single byte input truncated",
			input:    []byte{42},
			size:     0,
			expected: []byte{},
		},
		{
			name:     "pad large input",
			input:    []byte{1, 2},
			size:     10,
			expected: []byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BytesPadTo(tt.input, tt.size)

			if !bytes.Equal(result, tt.expected) {
				t.Errorf("BytesPadTo(%v, %d) = %v, want %v",
					tt.input, tt.size, result, tt.expected)
			}

			if len(result) != tt.size {
				t.Errorf("BytesPadTo(%v, %d) returned slice of length %d, want %d",
					tt.input, tt.size, len(result), tt.size)
			}
		})
	}
}

func TestBytesPadToEdgeCases(t *testing.T) {
	t.Run("negative size", func(t *testing.T) {
		input := []byte{1, 2, 3}
		size := -1
		result := BytesPadTo(input, size)

		if len(result) != 0 {
			t.Errorf("Expected empty slice for negative size, got %v", result)
		}
	})
}

func TestBytesPadToDoesNotModifyOriginal(t *testing.T) {
	original := []byte{1, 2, 3, 4, 5}
	originalCopy := make([]byte, len(original))
	copy(originalCopy, original)

	// Test padding case
	BytesPadTo(original, 8)
	if !bytes.Equal(original, originalCopy) {
		t.Error("bytesPadTo modified the original slice when padding")
	}

	// Test truncating case
	result := BytesPadTo(original, 3)
	if !bytes.Equal(original, originalCopy) {
		t.Error("bytesPadTo modified the original slice when truncating")
	}

	// Verify results are independent
	if len(result) >= 1 {
		result[0] = 99
		if original[0] == 99 {
			t.Error("Modifying result affected original slice")
		}
	}
}
