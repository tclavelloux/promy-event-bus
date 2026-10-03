# promy-event-bus — Redis Streams event bus library for the Promy platform

[![CI](https://github.com/tclavelloux/promy-event-bus/actions/workflows/ci.yml/badge.svg)](https://github.com/tclavelloux/promy-event-bus/actions/workflows/ci.yml)

<!-- TOC -->
* [Overview](#overview)
* [Architecture](#architecture)
* [Boundaries](#boundaries)
* [Resources](#resources)
<!-- TOC -->

## Overview

Go library providing a Redis Streams event bus for the Promy microservices. It handles at-least-once delivery, consumer groups, exponential backoff retry, dead-letter queue (DLQ) routing, and schema governance. Services import this module. It ships no HTTP server and no service binary; the only binary is the DLQ operator CLI (`cmd/dlq`).

**Module:** `github.com/tclavelloux/promy-event-bus`

```bash
go get github.com/tclavelloux/promy-event-bus
```

### Implemented

- Publish — `EventPublisher` (`Publish`, `PublishBatch`) on Redis Streams, with struct-tag and `Validate()` checks
- Subscribe — `EventSubscriber` with consumer groups, batch size, block duration, bounded concurrency
- Retry — 3 attempts with exponential backoff
- DLQ — exhausted events routed to `events:dlq` as `DLQEntry`; `cmd/dlq` inspects and replays
- Event schema registry — YAML contracts under `registry/`, validated in CI
- Test doubles — `testutil` (`MockPublisher`, `MockSubscriber`, `TestEvent`)

### Planned

- Transactional outbox for Tier 1 publishing — deferred (see [HOWTO.md](HOWTO.md#tier-1-pattern-publish-before-commit))

## Architecture

### Project Structure

```
eventbus/       Public interfaces, types, config, validation, DLQEntry
streams/        Stream name constants (StreamUsers, StreamDLQ, etc.)
redis/          Redis Streams implementation of EventPublisher & EventSubscriber
testutil/       MockPublisher, MockSubscriber, TestEvent for downstream testing
cmd/dlq/        DLQ inspect & replay CLI tool
registry/       Event schema registry (YAML contracts, CI validation)
scripts/        Lint/tidy hook wrappers, validate-registry.sh
examples/       Runnable publisher/subscriber demos
```

### Dispatch

- Publisher: `redis/publisher.go` runs `eventbus.ValidateStruct` (struct tags), then `event.Validate()`, before writing to the stream. `PublishBatch` validates every event first.
- Subscriber: `redis/subscriber.go` hands each handler a `rawEvent` (id, type, timestamp from metadata; payload via `Data()`). Services deserialize by event type.
- Retries run in-process: sleep, re-add to the stream with `attempt+1`, ACK the original.
- DLQ routing goes through `SubscriptionConfig.DLQPublisher`.

See [HOWTO.md](HOWTO.md) for Quick Start, configuration, Yokai integration, and the contributor workflow.

## Boundaries

### Public API

| Symbol | Package | Contract |
|---|---|---|
| `EventPublisher` | [`eventbus`](eventbus/publisher.go) | `Publish`, `PublishBatch` (all or none), `Close`, `Health` |
| `EventSubscriber` | [`eventbus`](eventbus/subscriber.go) | `Subscribe` (blocks until ctx cancelled), `Close`, `Health`; `SubscriptionConfig` carries stream, group, consumer ID, handler, batch/block/concurrency, `DLQPublisher`, `DLQService` |
| `Event` | [`eventbus`](eventbus/event.go) | `EventType()`, `EventID()`, `EventTime()`, `Data() string`, `Validate()`; `BaseEvent` provides the common fields |
| `Config` | [`eventbus`](eventbus/config.go) | `Redis` (`RedisConfig`) and `Consumer` (`ConsumerConfig`: group, consumer ID, `Defaults`, per-stream `Streams` overrides); `StreamConfig(stream)` resolves effective values |
| `redis.NewPublisher` / `redis.NewSubscriber` | [`redis`](redis/) | Redis implementations of the interfaces |

### Streams

Stream constants live in [`streams/streams.go`](streams/streams.go); owners are declared in `registry/streams/*/stream.yaml`.

| Stream | Owner | Purpose |
|--------|-------|---------|
| `events:users` | promy-user | User lifecycle events |
| `events:subscriptions` | promy-subscription | Subscription lifecycle events |
| `events:promotions` | promy-product | Promotion events |
| `events:products` | promy-product | Product catalogue events |
| `events:identifications` | promy-identifier | AI identification results |
| `events:dlq` | platform (multi-writer) | Dead-letter queue |

- One stream, one owner: only the owner publishes business events to it.
- `events:dlq` is a failure sink. Any service may write to it on retry exhaustion; never publish business events to it.

### Retry and Dead-Letter Queue

| Attempt | Delay |
|---------|-------|
| 1 | 0 ms |
| 2 | 100 ms |
| 3 | 500 ms |

- Delays beyond attempt 3 cap at 10 s (`calculateBackoff` in [`redis/subscriber.go`](redis/subscriber.go)); the subscriber stops at 3 attempts.
- After 3 attempts with `DLQPublisher` set, the event is wrapped in a `DLQEntry` and published to `events:dlq`. Otherwise it is dropped and ACKed.

DLQ entry format:

```json
{
  "original_stream": "events:users",
  "original_event_id": "550e8400-e29b-41d4-a716-446655440000",
  "original_event_type": "user.registered",
  "original_payload": "{\"user_id\":\"u-1\",\"email\":\"thomas@example.com\"}",
  "failure_reason": "timeout calling email service",
  "failed_at": "2026-05-25T14:30:00Z",
  "failed_service": "promy-crm",
  "attempts_exhausted": 3
}
```

### DLQ operator CLI

- [`cmd/dlq`](cmd/dlq/main.go) has two subcommands: `inspect` (stats) and `replay` (re-publish to the original stream, then delete the DLQ entry).
- Run via `make dlq-inspect` and `make dlq-replay`. Filters, flags, and examples: [HOWTO.md](HOWTO.md#operator-tooling).

### Event Schema Registry

`registry/streams/` is the canonical source of truth for which events exist. Each stream has a `stream.yaml` (`stream`, `owner`, `description`); each event has its own YAML contract. `registry/streams/dlq/` has a `stream.yaml` and no events.

```
registry/streams/
  users/
    stream.yaml
    events/
      user.registered.yaml
      user.preferences.updated.yaml
      user.location.updated.yaml
  promotions/
    stream.yaml
    events/
      promotion.created.yaml
      promotion.updated.yaml
  ...
```

Adding a new event:

1. Open a PR adding `registry/streams/<domain>/events/<event-name>.yaml`.
2. Follow the schema: `name`, `tier`, `description`, `fields` (with `type`, `format`, `required`, `description`), `example`.
3. [`registry.yaml`](.github/workflows/registry.yaml) runs [`scripts/validate-registry.sh`](scripts/validate-registry.sh) on PRs touching `registry/**`. Fix every reported error before merging.
4. PR merged = the event contract is official.
5. Implement the event struct in your service's `internal/events/` package.

A new stream also needs `registry/streams/<domain>/stream.yaml` and a constant in [`streams/streams.go`](streams/streams.go).

Naming conventions (enforced by CI):

| Rule | Example |
|---|---|
| Event name: dot-separated snake_case segments (past-tense verb by convention, not checked) | `user.registered`, `user.preferences.updated` |
| Field names: snake_case | `user_id`, `discounted_price` |
| Field `type`: `string`, `number`, `boolean`, `object`, `array` | |
| Field `format` (optional): `uuid`, `email`, `date-time`, `uri` | |
| `name` in YAML must match the filename | `user.registered.yaml` -> `name: user.registered` |
| `tier` must be `1` (business-critical) or `2` (best-effort) | |

### Not owned by this library

- Event payload structs and event type constants: each producing service owns them.
- Business logic.
- Consumer DTOs and subscription topology: each consuming service owns them.

Full split: [HOWTO.md > Scope Split](HOWTO.md#scope-split).

## Resources

- [HOWTO.md](HOWTO.md) — Quick Start, configuration, Yokai integration guide, development and CI
- [examples/](examples/) — runnable publisher/subscriber demos
- [registry/](registry/) — event schema registry
- [promy-product](https://github.com/tclavelloux/promy-product) — promotion catalog service
- [promy-user](https://github.com/tclavelloux/promy-user) — user management service
- [promy-identifier](https://github.com/tclavelloux/promy-identifier) — AI product identification service
- [promy-crm](https://github.com/tclavelloux/promy-crm) — CRM service

<!-- readme-updated-at: ac7645c -->
