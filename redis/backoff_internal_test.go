//nolint:all // Test file
package redis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 0},
		{1, 0},
		{2, 100 * time.Millisecond},
		{3, 500 * time.Millisecond},
		{4, 2500 * time.Millisecond},
		{5, 10 * time.Second},
		{20, 10 * time.Second},
	}

	for _, tt := range tests {
		got := calculateBackoff(tt.attempt)
		assert.Equal(t, tt.want, got)
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name  string
		input string
		zero  bool
	}{
		{"valid RFC3339", "2024-01-15T10:30:00Z", false},
		{"empty string", "", true},
		{"garbage", "garbage", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTime(tt.input)

			if tt.zero {
				assert.True(t, got.IsZero())
			} else {
				assert.False(t, got.IsZero())
				assert.Equal(t, tt.input, got.Format(time.RFC3339))
			}
		})
	}
}
