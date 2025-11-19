

# promy-event-bus

A lightweight, production-ready event bus library for Go microservices using Redis Streams. Designed for asynchronous event-driven communication between services with at-least-once delivery, consumer groups, automatic retry, and **schema governance**.

## Features

- **🚀 Redis Streams**: Fast, scalable event streaming with Redis
- **📦 Type-Safe Events**: Strongly-typed event schemas as single source of truth
- **🔄 Consumer Groups**: Load distribution across multiple consumers
- **♻️ Automatic Retry**: Exponential backoff retry logic (3 attempts)
- **🎯 At-Least-Once Delivery**: Guaranteed message delivery with acknowledgments
- **⚡ Batch Publishing**: Efficient batch event publishing
- **🏥 Health Checks**: Built-in connection health monitoring
- **✅ Schema Validation**: Automatic JSON schema validation
- **🔒 Schema Governance**: Backward compatibility checking
- **🧪 Fully Tested**: Comprehensive test suite with integration tests
- **📝 Well Documented**: Clear examples and extensive documentation

## Installation

```bash
go get github.com/tclavelloux/promy-event-bus
```

## Quick Start

### Publisher Example

```go
package main

import (
    "context"
    "log"

    eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
    "github.com/tclavelloux/promy-event-bus/events"
    "github.com/tclavelloux/promy-event-bus/events/user"
    "github.com/tclavelloux/promy-event-bus/redis"
)

func main() {
    // Create publisher
    config := eventbus.RedisConfig{
        DSN:      "redis://localhost:6379/0",
        PoolSize: 10,
    }

    publisher, err := redis.NewPublisher(config)
    if err != nil {
        log.Fatal(err)
    }
    defer publisher.Close()

    // Publish an event
    event := user.NewUserRegisteredEvent("user-123", "john@example.com")

    ctx := context.Background()
    if err := publisher.Publish(ctx, events.StreamUsers, event); err != nil {
        log.Printf("Failed to publish: %v", err)
    }

    log.Println("Event published successfully!")
}
```

### Subscriber Example

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
    "github.com/tclavelloux/promy-event-bus/events"
    "github.com/tclavelloux/promy-event-bus/redis"
)

func main() {
    // Create subscriber
    config := eventbus.Config{
        Redis: eventbus.RedisConfig{
            DSN:      "redis://localhost:6379/0",
            PoolSize: 10,
        },
        Consumer: eventbus.ConsumerConfig{
            Group:          "notification-service-consumers",
            ConsumerID:     "worker-1",
            BatchSize:      10,
            BlockDuration:  2 * time.Second,
            MaxConcurrency: 5,
        },
    }

    subscriber, err := redis.NewSubscriber(config)
    if err != nil {
        log.Fatal(err)
    }
    defer subscriber.Close()

    // Define event handler
    handler := func(ctx context.Context, event eventbus.Event) error {
        log.Printf("Processing event: %s (type: %s)", event.EventID(), event.EventType())

        // Your business logic here

        return nil // Return nil on success, error to trigger retry
    }

    // Setup graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("Shutting down...")
        cancel()
    }()

    // Start consuming
    if err := subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
        Stream:         events.StreamUsers,
        ConsumerGroup:  config.Consumer.Group,
        ConsumerID:     config.Consumer.ConsumerID,
        Handler:        handler,
        BatchSize:      config.Consumer.BatchSize,
        BlockDuration:  config.Consumer.BlockDuration,
        MaxConcurrency: config.Consumer.MaxConcurrency,
    }); err != nil && err != context.Canceled {
        log.Fatal(err)
    }
}
```

## Event Schemas

The library provides predefined event schemas as a single source of truth:

### User Events
- `UserRegisteredEvent` - When a new user registers
- `UserPreferencesUpdatedEvent` - When user preferences are updated
- `UserLocationUpdatedEvent` - When user location is updated

### Promotion Events
- `PromotionCreatedEvent` - When a promotion is created
- `PromotionUpdatedEvent` - When a promotion is updated
- `PromotionDeletedEvent` - When a promotion is deleted

### Product Events
- `ProductIdentifiedEvent` - When a product is identified by AI

### Streams
- `events:users` - User-related events
- `events:promotions` - Promotion-related events
- `events:products` - Product-related events

## Schema Governance

The event bus includes **built-in schema validation** and **backward compatibility checking**.

### Schema Validation

Events are automatically validated when published:

```go
event := user.NewUserRegisteredEvent("user-123", "john@example.com")
// Validation happens automatically - invalid events are rejected
publisher.Publish(ctx, events.StreamUsers, event)
```

### Backward Compatibility

The library prevents breaking changes:

```bash
# Run compatibility tests
go test ./eventbus/... -run Compatibility
```

✅ **Safe changes:**
- Add optional fields
- Remove optional fields
- Widen constraints

❌ **Breaking changes:**
- Remove required fields
- Change field types
- Add required fields without defaults

## Configuration

### Publisher Configuration

```go
config := eventbus.RedisConfig{
    DSN:             "redis://localhost:6379/0", // Redis connection string
    PoolSize:        10,                         // Connection pool size
    MaxRetries:      3,                          // Max retry attempts
    MinRetryBackoff: 100 * time.Millisecond,    // Min retry delay
    MaxRetryBackoff: 3 * time.Second,           // Max retry delay
    DialTimeout:     5 * time.Second,           // Connection timeout
    ReadTimeout:     3 * time.Second,           // Read timeout
    WriteTimeout:    3 * time.Second,           // Write timeout
}
```

### Subscriber Configuration

```go
config := eventbus.Config{
    Redis: eventbus.RedisConfig{
        DSN:      "redis://localhost:6379/0",
        PoolSize: 10,
    },
    Consumer: eventbus.ConsumerConfig{
        Group:          "my-service-consumers", // Consumer group name
        ConsumerID:     "worker-1",             // Unique consumer ID
        BatchSize:      10,                     // Messages per batch
        BlockDuration:  2 * time.Second,        // Block duration
        MaxConcurrency: 5,                      // Concurrent processing
    },
}
```


## Testing

### Run Tests

```bash
# Run all tests (requires Redis running)
make test

# Run tests with Docker Compose (starts Redis automatically)
make test-integration

# Run only short tests (excludes integration tests)
make test-short

# Generate coverage report
make coverage
```

### Start Redis for Testing

```bash
# Start Redis via Docker Compose
make docker-up

# Stop Redis
make docker-down

# View Redis logs
make docker-logs
```

### Run Examples

```bash
# Terminal 1: Start subscriber
make example-subscriber

# Terminal 2: Publish events
make example-publisher
```

## Best Practices

### 1. Event Validation

Always validate events before publishing (automatic):

```go
event := user.NewUserRegisteredEvent(userID, email)
// Validation happens automatically
publisher.Publish(ctx, events.StreamUsers, event)
```

### 2. Error Handling

Return errors from handlers to trigger retry:

```go
handler := func(ctx context.Context, event eventbus.Event) error {
    if err := processEvent(event); err != nil {
        return err // Will retry up to 3 times
    }
    return nil // Success
}
```

### 3. Graceful Shutdown

Always handle shutdown signals:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-sigChan
    cancel()
}()
```

### 4. Consumer Groups

Use meaningful consumer group names:

```go
// Good: service-specific group name
Group: "notification-service-consumers"

// Bad: generic group name
Group: "consumers"
```

### 5. Batch Publishing

Use batch publishing for multiple events:

```go
events := []eventbus.Event{event1, event2, event3}
publisher.PublishBatch(ctx, stream, events) // More efficient than 3 separate Publish calls
```

### 6. Schema Versioning

Explicitly version your schemas:

```go
event := user.NewUserRegisteredEvent(userID, email)
event.SchemaVersion = "1.1" // Set schema version
```

## Retry Strategy

The subscriber implements automatic retry with exponential backoff:

- **Attempt 1**: Immediate
- **Attempt 2**: 100ms delay
- **Attempt 3**: 500ms delay

After 3 failed attempts, the message is acknowledged to prevent infinite loops.

> **Note**: Dead Letter Queue (DLQ) support is planned for future releases.

## Performance

- **Publisher**: Non-blocking, fire-and-forget pattern
- **Subscriber**: Configurable concurrency for parallel processing
- **Batch Operations**: Pipeline support for high throughput
- **Connection Pooling**: Reusable connections for efficiency

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   promy-event-bus                       │
│              (Abstraction Layer - Go Module)            │
├─────────────────────────────────────────────────────────┤
│  ┌──────────────┐         ┌──────────────┐            │
│  │ EventPublisher│         │EventSubscriber│            │
│  │  (Interface)  │         │  (Interface)  │            │
│  └───────┬──────┘         └───────┬───────┘            │
│          │                        │                     │
│  ┌───────▼──────────────────────▼────────┐            │
│  │     Implementation Layer                │            │
│  │  ┌─────────────┐   ┌────────────────┐ │            │
│  │  │   Redis     │   │  Pub/Sub       │ │            │
│  │  │  Streams    │   │  (Future)      │ │            │
│  │  └─────────────┘   └────────────────┘ │            │
│  └─────────────────────────────────────────┘            │
│                                                         │
│  ┌──────────────────────────────────────┐              │
│  │     Schema Governance                │              │
│  │  ┌──────────────┐  ┌──────────────┐ │              │
│  │  │  Validator   │  │   Registry   │ │              │
│  │  └──────────────┘  └──────────────┘ │              │
│  └──────────────────────────────────────┘              │
└─────────────────────────────────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
┌───────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐
│ promy-product│  │ promy-user  │  │promy-notif  │
│  (Publisher) │  │ (Publisher) │  │(Subscriber) │
└──────────────┘  └─────────────┘  └─────────────┘
```

## Development

```bash
# Install dependencies
go mod tidy

# Format code
make fmt

# Run linter
make lint

# Run vet
make vet

# Clean build artifacts
make clean
```

## Roadmap

- [x] Redis Streams implementation
- [x] Publisher with batch support
- [x] Subscriber with consumer groups
- [x] Basic retry logic (3 attempts)
- [x] Event schemas
- [x] Integration tests
- [x] Schema validation
- [x] Backward compatibility checking
- [ ] Dead Letter Queue (DLQ)
- [ ] Advanced retry strategies
- [ ] Metrics and observability
- [ ] Google Cloud Pub/Sub implementation
- [ ] Tracing integration
- [ ] Schema Registry HTTP service (optional)

## Documentation

- [Schema Governance Guide](SCHEMA_GOVERNANCE.md) - Detailed schema management guide
- [Examples](examples/) - Working code examples
- [Tests](eventbus/) - Comprehensive test suite

## Contributing

Contributions are welcome! This library is part of the Promy microservices ecosystem.

## License

See [LICENSE](LICENSE) file for details.

## Related Projects

- [promy-product](https://github.com/tclavelloux/promy-product) - Promotion catalog service
- [promy-user](https://github.com/tclavelloux/promy-user) - User management service
- [promy-identifier](https://github.com/tclavelloux/promy-identifier) - AI product identification service

---

**Made with ❤️ for the Promy ecosystem**
