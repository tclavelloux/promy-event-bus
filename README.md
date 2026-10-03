# promy-event-bus

A Go library providing a Redis Streams-based event bus for the Promy microservices platform. It handles at-least-once delivery, consumer groups, exponential backoff retry, dead-letter queue routing, and schema governance. Services import this module — it contains no HTTP server or binary (aside from the DLQ ops tool).

**Module:** `github.com/tclavelloux/promy-event-bus`

## Installation

```bash
go get github.com/tclavelloux/promy-event-bus
```

## Quick Start

### Publishing

```go
package main

import (
    "context"
    "log"

    eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
    "github.com/tclavelloux/promy-event-bus/redis"
    "github.com/tclavelloux/promy-event-bus/streams"
)

func main() {
    publisher, err := redis.NewPublisher(eventbus.RedisConfig{
        DSN:      "redis://localhost:6379/0",
        PoolSize: 10,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer publisher.Close()

    // Event structs live in each service (not in this library).
    // They must implement eventbus.Event.
    event := myservice.NewUserRegisteredEvent("user-123", "john@example.com")

    if err := publisher.Publish(context.Background(), streams.StreamUsers, event); err != nil {
        log.Printf("Failed to publish: %v", err)
    }
}
```

### Subscribing

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
    "github.com/tclavelloux/promy-event-bus/redis"
    "github.com/tclavelloux/promy-event-bus/streams"
)

func main() {
    config := eventbus.Config{
        Redis: eventbus.RedisConfig{DSN: "redis://localhost:6379/0", PoolSize: 10},
        Consumer: eventbus.ConsumerConfig{
            Group:      "notification-service",
            ConsumerID: "worker-1",
            Defaults: eventbus.ConsumerStreamConfig{
                BatchSize:      10,
                BlockDuration:  2 * time.Second,
                MaxConcurrency: 5,
            },
        },
    }

    subscriber, err := redis.NewSubscriber(config)
    if err != nil {
        log.Fatal(err)
    }
    defer subscriber.Close()

    // Optional: create a publisher for DLQ routing on exhausted retries
    dlqPublisher, _ := redis.NewPublisher(config.Redis)

    handler := func(ctx context.Context, event eventbus.Event) error {
        log.Printf("Processing: %s (%s)", event.EventID(), event.EventType())
        // return error to trigger retry; nil = success
        return nil
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() { <-sigChan; cancel() }()

    streamCfg := config.Consumer.StreamConfig(streams.StreamUsers)

    err = subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
        Stream:         streams.StreamUsers,
        ConsumerGroup:  config.Consumer.Group,
        ConsumerID:     config.Consumer.ConsumerID,
        Handler:        handler,
        BatchSize:      streamCfg.BatchSize,
        BlockDuration:  streamCfg.BlockDuration,
        MaxConcurrency: streamCfg.MaxConcurrency,
        DLQPublisher:   dlqPublisher, // nil = silent drop after max retries
        DLQService:     "notification-service",
    })
    if err != nil && err != context.Canceled {
        log.Fatal(err)
    }
}
```

## Streams

Stream constants live in the `streams` package:

| Stream | Owner | Purpose |
|--------|-------|---------|
| `events:users` | promy-user | User lifecycle events |
| `events:subscriptions` | promy-subscription | Subscription lifecycle events |
| `events:promotions` | promy-product | Promotion events |
| `events:products` | promy-product | Product catalogue events |
| `events:identifications` | promy-identifier | AI identification results |
| `events:dlq` | platform (multi-writer) | Dead-letter queue |

## Retry and Dead-Letter Queue

The subscriber retries failed messages with exponential backoff:

| Attempt | Delay |
|---------|-------|
| 1 | 0 ms |
| 2 | 100 ms |
| 3 | 500 ms |

After 3 attempts, if `DLQPublisher` is configured on the `SubscriptionConfig`, the event is wrapped in a `DLQEntry` and published to `events:dlq`. Otherwise it is silently dropped and ACKed.

### DLQ Entry Format

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

### DLQ Replay Tooling

Operators can inspect and replay DLQ entries using the CLI tool under `cmd/dlq/`:

```bash
# Inspect DLQ stats
make dlq-inspect

# Replay all entries for a specific stream
make dlq-replay stream=events:users

# Replay by event type
make dlq-replay type=user.registered

# Replay a single entry by Redis message ID
make dlq-replay id=1685000000000-0

# Dry-run (list without replaying)
make dlq-replay stream=events:subscriptions dry-run=true

# Replay all
make dlq-replay all=true
```

The replay tool re-publishes the original payload to the original stream, then deletes the DLQ entry. At-least-once semantics apply.

## Event Schema Registry

The `registry/streams/` directory is the canonical source of truth for what events exist on the platform. Each stream has a `stream.yaml` and each event has its own YAML file defining the contract.

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

### Adding a new event

1. Open a PR adding `registry/streams/<domain>/events/<event-name>.yaml`
2. Follow the schema: `name`, `tier`, `description`, `fields` (with `type`, `format`, `required`, `description`), `example`
3. CI runs `scripts/validate-registry.sh` — the PR cannot merge until it passes
4. PR merged = the event contract is official
5. Implement the event struct in your service's `internal/events/` package

### Naming conventions (enforced by CI)

| Rule | Example |
|---|---|
| Event name: dot-separated, verb in past tense | `user.registered`, `subscription.cancelled` |
| Field names: snake_case | `user_id`, `discounted_price` |
| Field `type`: `string`, `number`, `boolean`, `object`, `array` | |
| Field `format` (optional): `uuid`, `email`, `date-time`, `uri` | |
| `name` in YAML must match the filename | `user.registered.yaml` -> `name: user.registered` |
| `tier` must be `1` (business-critical) or `2` (best-effort) | |

## Configuration

### Per-Stream Consumer Overrides

The `ConsumerConfig` supports per-stream tuning. Defaults apply to all streams; non-zero fields in the `Streams` map override them:

```yaml
consumer:
  group: "my-service"
  consumer_id: "worker-1"
  defaults:
    batch_size: 10
    block_duration: 2s
    max_concurrency: 5
  streams:
    events:users:
      max_concurrency: 10  # higher concurrency for this stream only
```

```go
cfg := config.Consumer.StreamConfig("events:users")
// cfg.MaxConcurrency == 10 (overridden), cfg.BatchSize == 10 (default)
```

## Project Structure

```
eventbus/       Public interfaces, types, config, validation, DLQEntry
streams/        Stream name constants (StreamUsers, StreamDLQ, etc.)
redis/          Redis Streams implementation of EventPublisher & EventSubscriber
testutil/       MockPublisher, MockSubscriber, TestEvent for downstream testing
cmd/dlq/        DLQ inspect & replay CLI tool
registry/       Event schema registry (YAML contracts, CI validation)
examples/       Runnable publisher/subscriber demos
```

## Development

```bash
make test              # all tests (Redis on localhost:6389/15 — `make up` starts it via docker-compose;
                        # override with REDIS_TEST_DSN if pointing at a different instance)
make test-short        # unit tests only
make test-integration  # start Redis via Docker, run tests, stop Redis
make coverage          # generate coverage.html
make lint              # golangci-lint
make fmt               # go fmt
make vet               # go vet
make tidy              # go mod tidy
make up                # start Redis + RedisInsight GUI at :5540 (docker-compose)
make down              # stop Redis
make dlq-inspect       # show DLQ stats
make dlq-replay        # replay DLQ entries (see flags above)
make help              # list all targets

# Setup
make setup             # pre-commit install --install-hooks (pre-commit, commit-msg, pre-push)
```

### Local quality gate

`make setup` is not optional. Without it no git hooks are installed, so nothing lints, formats or scans for secrets until CI does.

- Install `pre-commit` first (`brew install pre-commit`); `make setup` fails without it.
- `make setup` installs all three hook types (`default_install_hook_types` in [`.pre-commit-config.yaml`](.pre-commit-config.yaml)).
- A hand-written `.git/hooks/pre-commit` is moved to `pre-commit.legacy` and keeps running. Delete it after setup.

| Stage | Checks |
|---|---|
| `pre-commit` | `no-direct-commit-to-main` (shared hook, blocks commits on `main`), golangci-lint (pinned), `golangci-lint fmt --diff`, config verify, `go mod tidy` check, gitleaks, whitespace/YAML hygiene |
| `commit-msg` | Conventional Commits |
| `pre-push` | `make check-coverage`: full race suite, then thresholds from [`.testcoverage.yml`](.testcoverage.yml) (also read by the CI `coverage` job). Runs on every push, including pushes without `.go` files |

The Go hooks call [`scripts/golangci-lint.sh`](scripts/golangci-lint.sh) and [`scripts/go-mod-tidy.sh`](scripts/go-mod-tidy.sh), which install the version in `.golangci-version` and exec it directly — the same version [`go-lint.yml`](https://github.com/tclavelloux/promy-github-workflows) runs in CI. A locally installed `golangci-lint` is neither needed nor consulted. The Go lint and tidy hooks only report: they fail without touching your files. The `trailing-whitespace` and `end-of-file-fixer` hooks rewrite files and fail the commit; re-stage the fixed files and commit again.

### CI

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs on `pull_request` to `main` (opened, synchronize, ready_for_review, reopened). Draft PRs skip every job. No `push: main` trigger.

- `lint`, `vuln`: reusable workflows from `promy-github-workflows` (`go-lint/v1`, `go-vuln/v1`).
- `test`: `go test -race` against a Redis service container (`REDIS_TEST_DSN=redis://localhost:6379/15`); uploads the coverage profile.
- `coverage`: reusable `go-coverage/v2`; needs `test`; thresholds from [`.testcoverage.yml`](.testcoverage.yml).
- `gitleaks`: scans the PR commit range. No `docker` job: this is a library.
- [`pr-title.yml`](.github/workflows/pr-title.yml) is a separate workflow.
- Squash-merge only: the PR title becomes the commit on `main`. It must be a Conventional Commits string; release-please parses it.

## Documentation

- [HOWTO.md](HOWTO.md) - Integration guide for downstream Yokai services
- [examples/](examples/) - Runnable publisher/subscriber demos
- [registry/](registry/) - Event schema registry

## Related Projects

- [promy-product](https://github.com/tclavelloux/promy-product) - Promotion catalog service
- [promy-user](https://github.com/tclavelloux/promy-user) - User management service
- [promy-identifier](https://github.com/tclavelloux/promy-identifier) - AI product identification service
- [promy-crm](https://github.com/tclavelloux/promy-crm) - CRM service

<!-- readme-updated-at: d71cde2 -->
