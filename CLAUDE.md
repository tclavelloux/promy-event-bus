# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

A standalone Go **library** (no binary, no HTTP server, no Yokai framework). It is imported by other `promy-*` microservices. It provides a Redis Streams–based event bus with at-least-once delivery, consumer groups, retry with exponential backoff, and typed event schemas.

Module: `github.com/tclavelloux/promy-event-bus`

## Commands

```bash
# Tests
make test                  # all tests, race detector, coverage
make test-short            # unit tests only (skips integration)
make test-integration      # start Redis → run all tests → stop Redis
make coverage              # generate coverage.html

# Code quality
make lint                  # golangci-lint
make fmt && make vet && make tidy

# Infrastructure
make up                    # start Redis (docker-compose)
make down                  # stop Redis

# Run a single test
go test -run TestPublisher ./redis/...
go test -run TestPublisher -v -race ./redis/...

# Examples
make example-publisher
make example-subscriber
```

Integration tests require Redis on `localhost:6379/1`. Guard: `if testing.Short() { t.Skip(...) }`. Use `make test-integration` which handles Docker lifecycle, or `make up` before running tests manually.

## Architecture

```
eventbus/    # Public interfaces and types (Event, EventPublisher, EventSubscriber,
             # Config, BaseEvent, sentinel errors, validation singleton)
events/      # Typed event schemas — single source of truth for all domain events
  promotion/ # PromotionCreatedEvent, PromotionUpdatedEvent
  user/      # UserRegisteredEvent, UserPreferencesUpdatedEvent, UserLocationUpdatedEvent
  product/   # ProductIdentifiedEvent
  types.go   # EventXxx and StreamXxx constants
redis/       # Concrete implementation of EventPublisher and EventSubscriber
testutil/    # MockPublisher and MockSubscriber (testify/mock) for downstream services
pkg/ptr/     # Scalar pointer helpers used in event constructors
examples/    # Runnable publisher/subscriber demos; excluded from lint
```

### Event contract

All events embed `eventbus.BaseEvent` (auto UUID, UTC timestamp, source, version `"1.0"`). Each event type lives in its subdomain package and provides:
- A `New*` constructor
- A `Validate()` method with business rules (beyond struct-tag validation)

Validation is two-layered: struct tags (`go-playground/validator`) run in the publisher before every `Publish`, then `Validate()` for business rules.

### Subscriber dispatch

Subscribers receive a `rawEvent` wrapping Redis stream fields. They must type-switch on `event.EventType()` and unmarshal the `payload` JSON field themselves to get domain-specific data. The `metadata` field carries `id`, `type`, `timestamp`, `version`, `attempt`.

### Retry behaviour

Max 3 attempts per message (backoff: 0 ms → 100 ms → 500 ms, capped at 10 s). After max retries the message is ACKed to prevent an infinite loop. No dead-letter queue yet.

### Adding a new event type

1. Add stream constant to `events/types.go` if needed.
2. Create `events/<domain>/<event>.go` with struct embedding `eventbus.BaseEvent`, a `New*` constructor, and a `Validate()` method.
3. Add the event type constant to `events/types.go`.

### Stream naming

`events:<domain>` — e.g., `events:promotions`, `events:users`, `events:products`.
