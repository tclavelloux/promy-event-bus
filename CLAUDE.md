# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

A standalone Go **library** (no binary, no HTTP server, no Yokai framework). It is imported by other `promy-*` microservices. It provides a Redis Streams–based event bus with at-least-once delivery, consumer groups, retry with exponential backoff, dead-letter routing, and an event schema registry.

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

Integration tests require Redis on `localhost:6389/15` (`make up` maps host 6389 to the container's 6379; override with `REDIS_TEST_DSN`; CI uses `localhost:6379/15`). Tests skip when Redis is unreachable. Guard: `if testing.Short() { t.Skip(...) }`. Use `make test-integration` which handles Docker lifecycle, or `make up` before running tests manually.

## Architecture

```
eventbus/    # Public interfaces and types (Event, EventPublisher, EventSubscriber,
             # Config, BaseEvent, sentinel errors, validation singleton)
streams/     # Stream name constants (StreamUsers, StreamPromotions, ..., StreamDLQ)
registry/    # Event schema registry (registry/streams/<stream>/), checked by scripts/validate-registry.sh
redis/       # Concrete implementation of EventPublisher and EventSubscriber
testutil/    # MockPublisher and MockSubscriber (testify/mock) for downstream services
cmd/dlq/     # DLQ inspection/replay CLI
examples/    # Runnable publisher/subscriber demos; excluded from lint
scripts/     # golangci-lint.sh, go-mod-tidy.sh (pre-commit hooks), validate-registry.sh
```

### Event contract

This library defines the `Event` interface and `BaseEvent` ([`eventbus/event.go`](eventbus/event.go)); concrete event structs and type-string constants live in each producing service.
- Services embed `eventbus.BaseEvent` (use `NewBaseEvent` for UUID, UTC timestamp, source, version `"1.0"`), override `Data()` to return the JSON payload, and add a `Validate()` for business rules.
- The publisher validates struct tags (`go-playground/validator`) before every `Publish`, then calls `Validate()`; see [`eventbus/validation.go`](eventbus/validation.go).
- The contract of each event is declared in [`registry/streams/`](registry/streams/) and checked by `scripts/validate-registry.sh`.

### Subscriber dispatch

Subscribers receive a `rawEvent` wrapping Redis stream fields. They must type-switch on `event.EventType()` and unmarshal the `payload` JSON field themselves to get domain-specific data. The `metadata` field carries `id`, `type`, `timestamp`, `version`, `attempt`.

### Retry behaviour

Max 3 attempts per message (backoff: 0 ms → 100 ms → 500 ms, capped at 10 s). After max retries the message is ACKed to prevent an infinite loop. If `SubscriptionConfig.DLQPublisher` is set, the subscriber first wraps the event in a `DLQEntry` ([`eventbus/dlq.go`](eventbus/dlq.go)) and publishes it to `streams.StreamDLQ`; if nil, the event is dropped. Inspect and replay with `make dlq-inspect` / `make dlq-replay` ([`cmd/dlq/`](cmd/dlq/)).

### Adding a new event type

1. New stream only: add the constant to [`streams/streams.go`](streams/streams.go) and `registry/streams/<domain>/stream.yaml`.
2. Add `registry/streams/<domain>/events/<event-name>.yaml` (`name` = filename, `tier` 1 or 2, `fields`, `example`).
3. Run `bash scripts/validate-registry.sh` (needs `yq`) before opening the PR.
4. Do not add payload structs or type constants here; the producing service owns them (see HOWTO > Scope Split).

### Stream naming

`events:<domain>` — e.g., `events:promotions`, `events:users`, `events:products`.
