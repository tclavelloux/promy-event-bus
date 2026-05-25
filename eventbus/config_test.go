package eventbus_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/streams"
)

const (
	testGroup      = "test-group"
	testConsumerID = "test-consumer"
)

func TestConsumerConfig_StreamConfig(t *testing.T) {
	defaults := eventbus.ConsumerStreamConfig{
		BatchSize:      50,
		BlockDuration:  2 * time.Second,
		MaxConcurrency: 10,
	}

	t.Run("returns defaults when no stream entry exists", func(t *testing.T) {
		cfg := eventbus.ConsumerConfig{
			Group:      testGroup,
			ConsumerID: testConsumerID,
			Defaults:   defaults,
			Streams: map[string]eventbus.ConsumerStreamConfig{
				"events:other": {MaxConcurrency: 1},
			},
		}

		result := cfg.StreamConfig(streams.StreamUsers)

		assert.Equal(t, defaults.BatchSize, result.BatchSize)
		assert.Equal(t, defaults.BlockDuration, result.BlockDuration)
		assert.Equal(t, defaults.MaxConcurrency, result.MaxConcurrency)
	})

	t.Run("applies partial override", func(t *testing.T) {
		cfg := eventbus.ConsumerConfig{
			Group:      testGroup,
			ConsumerID: testConsumerID,
			Defaults:   defaults,
			Streams: map[string]eventbus.ConsumerStreamConfig{
				streams.StreamUsers: {MaxConcurrency: 1},
			},
		}

		result := cfg.StreamConfig(streams.StreamUsers)

		assert.Equal(t, defaults.BatchSize, result.BatchSize)
		assert.Equal(t, defaults.BlockDuration, result.BlockDuration)
		assert.Equal(t, 1, result.MaxConcurrency)
	})

	t.Run("applies full override", func(t *testing.T) {
		cfg := eventbus.ConsumerConfig{
			Group:      testGroup,
			ConsumerID: testConsumerID,
			Defaults:   defaults,
			Streams: map[string]eventbus.ConsumerStreamConfig{
				streams.StreamUsers: {
					BatchSize:      100,
					BlockDuration:  5 * time.Second,
					MaxConcurrency: 1,
				},
			},
		}

		result := cfg.StreamConfig(streams.StreamUsers)

		assert.Equal(t, 100, result.BatchSize)
		assert.Equal(t, 5*time.Second, result.BlockDuration)
		assert.Equal(t, 1, result.MaxConcurrency)
	})

	t.Run("handles nil streams map without panic", func(t *testing.T) {
		cfg := eventbus.ConsumerConfig{
			Group:      testGroup,
			ConsumerID: testConsumerID,
			Defaults:   defaults,
			Streams:    nil,
		}

		result := cfg.StreamConfig(streams.StreamUsers)

		assert.Equal(t, defaults.BatchSize, result.BatchSize)
		assert.Equal(t, defaults.BlockDuration, result.BlockDuration)
		assert.Equal(t, defaults.MaxConcurrency, result.MaxConcurrency)
	})

	t.Run("ignores zero-value override fields", func(t *testing.T) {
		cfg := eventbus.ConsumerConfig{
			Group:      testGroup,
			ConsumerID: testConsumerID,
			Defaults:   defaults,
			Streams: map[string]eventbus.ConsumerStreamConfig{
				streams.StreamUsers: {},
			},
		}

		result := cfg.StreamConfig(streams.StreamUsers)

		assert.Equal(t, defaults.BatchSize, result.BatchSize)
		assert.Equal(t, defaults.BlockDuration, result.BlockDuration)
		assert.Equal(t, defaults.MaxConcurrency, result.MaxConcurrency)
	})
}
