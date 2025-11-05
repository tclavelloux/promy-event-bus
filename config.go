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

// ConsumerConfig configures event consumption.
type ConsumerConfig struct {
	// Group is the consumer group name.
	// All consumers with the same group share the workload.
	Group string `yaml:"group"`

	// ConsumerID identifies this consumer within the group.
	// Should be unique per consumer instance.
	ConsumerID string `yaml:"consumer_id"`

	// BlockDuration for blocking reads.
	// How long to wait for new messages before checking again.
	// Default: 1s
	BlockDuration time.Duration `yaml:"block_duration"`

	// BatchSize for batch reads.
	// Number of messages to fetch per read operation.
	// Default: 10
	BatchSize int `yaml:"batch_size"`

	// MaxConcurrency for parallel processing.
	// Maximum number of messages to process concurrently.
	// Default: 5
	MaxConcurrency int `yaml:"max_concurrency"`
}
