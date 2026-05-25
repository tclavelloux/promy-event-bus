package eventbus

import "time"

// Config configures the event bus.
type Config struct {
	// Type specifies the implementation ("redis" or "pubsub").
	Type string `yaml:"type"`

	// Redis configuration (when Type == "redis").
	Redis RedisConfig `yaml:"redis"`
	// Consumer configuration.
	Consumer ConsumerConfig `yaml:"consumer"`
}

// RedisConfig configures Redis Streams.
type RedisConfig struct {
	// DSN is the Redis connection string.
	// Example: "redis://localhost:6379/0" or "redis://user:password@host:port/db"
	DSN string `yaml:"dsn"`

	// PoolSize is the connection pool size.
	// Default: 10
	PoolSize int `yaml:"pool_size"`

	// MaxRetries for Redis operations.
	// Default: 3
	MaxRetries int `yaml:"max_retries"`

	// MinRetryBackoff is the minimum retry delay.
	// Default: 8ms
	MinRetryBackoff time.Duration `yaml:"min_retry_backoff"`

	// MaxRetryBackoff is the maximum retry delay.
	// Default: 512ms
	MaxRetryBackoff time.Duration `yaml:"max_retry_backoff"`

	// DialTimeout for establishing connections.
	// Default: 5s
	DialTimeout time.Duration `yaml:"dial_timeout"`

	// ReadTimeout for read operations.
	// Default: 3s
	ReadTimeout time.Duration `yaml:"read_timeout"`

	// WriteTimeout for write operations.
	// Default: 3s
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// ConsumerStreamConfig holds the tuning parameters for consuming a single stream.
type ConsumerStreamConfig struct {
	// BlockDuration is how long to block waiting for new messages before polling again.
	BlockDuration time.Duration `yaml:"block_duration"`

	// BatchSize is the number of messages to fetch per read operation.
	BatchSize int `yaml:"batch_size"`

	// MaxConcurrency is the maximum number of messages to process in parallel.
	MaxConcurrency int `yaml:"max_concurrency"`
}

// ConsumerConfig configures event consumption.
type ConsumerConfig struct {
	// Group is the consumer group name shared by all workers of this service.
	Group string `yaml:"group"`

	// ConsumerID uniquely identifies this consumer instance within the group.
	ConsumerID string `yaml:"consumer_id"`

	// Defaults provides the baseline tuning for all streams.
	Defaults ConsumerStreamConfig `yaml:"defaults"`

	// Streams holds per-stream overrides keyed by stream name (e.g., "events:users").
	// Only non-zero fields override the defaults.
	Streams map[string]ConsumerStreamConfig `yaml:"streams"`
}

// StreamConfig returns the resolved ConsumerStreamConfig for the given stream name.
// It starts from Defaults and applies any non-zero fields from the per-stream override.
func (c ConsumerConfig) StreamConfig(stream string) ConsumerStreamConfig {
	cfg := c.Defaults
	if override, ok := c.Streams[stream]; ok {
		if override.BatchSize > 0 {
			cfg.BatchSize = override.BatchSize
		}
		if override.BlockDuration > 0 {
			cfg.BlockDuration = override.BlockDuration
		}
		if override.MaxConcurrency > 0 {
			cfg.MaxConcurrency = override.MaxConcurrency
		}
	}

	return cfg
}
