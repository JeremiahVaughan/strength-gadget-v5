package model

import (
	"testing"
)

func TestSerializeUniqueMember(t *testing.T) {
	tests := []struct {
		name     string
		score    int
		index    uint16
		expected string
	}{
		{
			name:     "base case",
			score:    100,
			index:    1,
			expected: "100:1",
		},
		{
			name:     "negative score",
			score:    -100,
			index:    1,
			expected: "-100:1",
		},
		{
			name:     "zero score",
			score:    0,
			index:    3,
			expected: "0:3",
		},
		{
			name:     "large index",
			score:    200,
			index:    65535,
			expected: "200:65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := serializeUniqueMember(tt.score, tt.index)
			if result != tt.expected {
				t.Fatalf("Expected %s but got %s", tt.expected, result)
			}
		})
	}
}
