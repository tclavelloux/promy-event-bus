# HOWTO: Integrating promy-event-bus in a Yokai Service

This guide is the single source of truth for integrating promy-event-bus (Redis Streams) into any Yokai-based microservice. It covers both publishing and subscribing, with exact file paths, naming conventions, config keys, FX wiring, and test strategies.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Dependency Installation](#dependency-installation)
3. [Configuration](#configuration)
4. [Docker Compose Setup](#docker-compose-setup)
5. [Publisher Integration](#publisher-integration)
6. [Subscriber Integration](#subscriber-integration)
7. [FX Wiring & Registration](#fx-wiring--registration)
8. [Bootstrap & TestBootstrapper](#bootstrap--testbootstrapper)
9. [Event Handler Implementation](#event-handler-implementation)
10. [Testing Strategy](#testing-strategy)
11. [Graceful Degradation](#graceful-degradation)
12. [Reference: Naming Conventions](#reference-naming-conventions)
13. [Reference: Event Schemas](#reference-event-schemas)
14. [Reference: Config Keys](#reference-config-keys)
15. [FAQ](#faq)

---

## Prerequisites

- Your service is built with Yokai (fxcore, fxhttpserver, fxsql, etc.)
- The shared Redis instance is running via `promy-event-bus`'s docker-compose (`eventbus-network`)
- You have identified whether your service is a **publisher**, a **subscriber**, or **both**

---

## Dependency Installation

```bash
go get github.com/tclavelloux/promy-event-bus
go get github.com/redis/go-redis/v9
```

If your service subscribes to events (not just publishes), you also need the Yokai worker module:

```bash
go get github.com/ankorstore/yokai/fxworker
```

---

## Configuration

All event bus configuration lives under the `modules.event_bus` key in `configs/config.yaml`. This is a **hard convention** — every service MUST use this exact path to ensure consistency across the platform.

### Full config block (copy-paste into your `configs/config.yaml`)

```yaml
modules:
  # ... other modules (core, http, sql, log, trace, metrics) ...
  event_bus:
    redis:
      dsn: redis://default:${REDIS_TRACKING_PASSWORD}@${REDIS_TRACKING_HOST}:${REDIS_TRACKING_PORT}/${REDIS_TRACKING_DB}
      pool_size: 20
      max_retries: 5
      min_retry_backoff: 100ms
      max_retry_backoff: 5s
      dial_timeout: 10s
      read_timeout: 5s
      write_timeout: 5s
    consumer:
      group: <your-service>-consumers    # e.g., "crm-service-consumers"
      consumer_id: ${CONSUMER_ID:-<your-service>-worker-1}
      batch_size: 50
      block_duration: 2s
      max_concurrency: 10
```

### Key rules

| Key | Convention | Example |
|-----|-----------|---------|
| `modules.event_bus.redis.dsn` | Always use env vars for credentials | `redis://default:${REDIS_TRACKING_PASSWORD}@...` |
| `modules.event_bus.consumer.group` | `<service-name>-consumers` | `crm-service-consumers` |
| `modules.event_bus.consumer.consumer_id` | env-var with hostname fallback — **MUST be unique per replica** | `${CONSUMER_ID:-${HOSTNAME}-crm-worker}` |

### .env file additions

```env
REDIS_TRACKING_HOST=promy-redis
REDIS_TRACKING_PORT=6379
REDIS_TRACKING_PASSWORD=
REDIS_TRACKING_DB=0
CONSUMER_ID=${HOSTNAME:-<your-service>-worker-1}
```

> **Scaling warning**: if you run multiple replicas, each MUST have a unique `CONSUMER_ID`. Two consumers sharing the same ID will cause only one to receive messages — the other sits idle. In Kubernetes/Railway, use the pod name or instance ID (e.g., `${RAILWAY_REPLICA_ID}` or `${HOSTNAME}`).

For local development (connecting to the shared `eventbus-network`), `REDIS_TRACKING_HOST=promy-redis` and `REDIS_TRACKING_PORT=6379` are the defaults.

---

## Docker Compose Setup

Your service must attach to the `eventbus-network` to reach the shared Redis instance hosted by `promy-event-bus`.

### Add to your `docker-compose.yaml`

```yaml
services:
  <your-service>-app-server:
    # ... existing build/ports/volumes ...
    networks:
      - <your-service>-app-network    # your internal network
      - eventbus-network               # shared Redis Streams access
    volumes:
      - .:/app
      - ../promy-event-bus:/promy-event-bus  # Mount for local go.mod replace

networks:
  <your-service>-app-network:
    driver: bridge
  eventbus-network:
    external: true  # Defined by promy-event-bus docker-compose
```

### Important

- The `eventbus-network` is **always external** in your service. It is owned and defined by `promy-event-bus/docker-compose.yaml` with `name: eventbus-network`.
- You must start `promy-event-bus` first (`cd ../promy-event-bus && docker compose up -d`) so the network exists.
- The `../promy-event-bus:/promy-event-bus` volume mount enables local `go.mod` replace directives during development.

---

## Publisher Integration

Use the publisher when your service produces domain events that other services need to react to (e.g., promy-product publishes `promotion.created`, promy-user publishes `user.registered`).

### File: `internal/infra/eventbus/publisher.go`

```go
package eventbus

import (
	"context"

	"github.com/ankorstore/yokai/config"
	"github.com/ankorstore/yokai/log"
	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/redis"
)

// NewEventPublisher creates a new event publisher from configuration.
// Returns nil if Redis is not configured or if creation fails (graceful degradation).
// The calling service MUST handle nil publisher gracefully by skipping event publication.
func NewEventPublisher(cfg *config.Config) eventbus.EventPublisher {
	dsn := cfg.GetString("modules.event_bus.redis.dsn")
	if dsn == "" {
		return nil
	}

	redisConfig := eventbus.RedisConfig{
		DSN:             dsn,
		PoolSize:        cfg.GetInt("modules.event_bus.redis.pool_size"),
		MaxRetries:      cfg.GetInt("modules.event_bus.redis.max_retries"),
		MinRetryBackoff: cfg.GetDuration("modules.event_bus.redis.min_retry_backoff"),
		MaxRetryBackoff: cfg.GetDuration("modules.event_bus.redis.max_retry_backoff"),
		DialTimeout:     cfg.GetDuration("modules.event_bus.redis.dial_timeout"),
		ReadTimeout:     cfg.GetDuration("modules.event_bus.redis.read_timeout"),
		WriteTimeout:    cfg.GetDuration("modules.event_bus.redis.write_timeout"),
	}

	publisher, err := redis.NewPublisher(redisConfig)
	if err != nil {
		return nil
	}

	return publisher
}

// PublishEvent publishes an event asynchronously (fire-and-forget).
// If publisher is nil, the event is silently skipped with a warning log.
//
// WHY FIRE-AND-FORGET:
// The goroutine decouples the HTTP request lifecycle from the Redis round-trip.
// Redis XADD is synchronous at the transport layer — when Publish() returns nil,
// the event IS persisted in the stream. The goroutine simply means the HTTP handler
// does not block on this side-effect before responding to the client.
//
// TRADEOFF: if Redis is temporarily unreachable at the moment of publish, the event
// is lost (only logged). For our use case (push notifications, analytics), this is
// acceptable. For guaranteed delivery, a transactional outbox pattern would be needed —
// this is explicitly out of scope until post-launch (Phase 7+).
//
// GOROUTINE SAFETY: each call spawns one goroutine bounded by dial_timeout (10s).
// Under sustained Redis outage at high request rate, goroutines accumulate.
// Future improvement: replace with a bounded worker pool or channel-based buffer.
func PublishEvent(ctx context.Context, publisher eventbus.EventPublisher, stream string, event eventbus.Event) {
	logger := log.CtxLogger(ctx)

	if publisher == nil {
		logger.Warn().
			Str("eventId", event.EventID()).
			Str("eventType", event.EventType()).
			Msg("publisher not available, skipping event publication")
		return
	}

	//nolint:contextcheck,gosec // Fire-and-forget: background context ensures publish completes even if request is cancelled
	go func() {
		if err := publisher.Publish(context.Background(), stream, event); err != nil {
			logger.Error().
				Err(err).
				Str("eventId", event.EventID()).
				Str("eventType", event.EventType()).
				Msg("failed to publish event")
		} else {
			logger.Info().
				Str("eventId", event.EventID()).
				Str("eventType", event.EventType()).
				Msg("event published successfully")
		}
	}()
}
```

### Usage in a domain service

```go
package user

import (
	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/events"
	userevents "github.com/tclavelloux/promy-event-bus/events/user"
	eb "github.com/<org>/<repo>/internal/infra/eventbus"
)

type DefaultService struct {
	repository Repository
	publisher  eventbus.EventPublisher  // injected via FX — may be nil
}

func (s *DefaultService) Register(ctx context.Context, user *User) error {
	// ... validation, repository.Create() ...

	// Publish event (fire-and-forget, nil-safe)
	event := userevents.NewUserRegisteredEvent(user.ID, user.Email)
	eb.PublishEvent(ctx, s.publisher, events.StreamUsers, event)

	return nil
}
```

---

## Subscriber Integration

Use the subscriber when your service needs to react to events produced by other services (e.g., promy-crm subscribes to `promotion.created` to send notifications).

### Architecture: Yokai Worker pattern

Subscribers are implemented as **Yokai Workers** (`fxworker`). This is NOT optional — do not use raw FX lifecycle hooks.

A worker is:
- A long-running goroutine managed by the Yokai worker pool
- Automatically started on app boot and stopped on shutdown
- Integrated with health checks (liveness/readiness probes)
- Excluded from tests via a separate `TestBootstrapper`

### File: `internal/infra/eventbus/subscriber.go`

```go
package eventbus

import (
	"github.com/ankorstore/yokai/config"
	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/redis"
)

// NewEventSubscriber creates a new event subscriber from configuration.
// Returns nil if Redis is not configured or unreachable (graceful degradation).
// Workers MUST check for nil before calling Subscribe().
func NewEventSubscriber(cfg *config.Config) eventbus.EventSubscriber {
	dsn := cfg.GetString("modules.event_bus.redis.dsn")
	if dsn == "" {
		return nil
	}

	ebConfig := eventbus.Config{
		Redis: eventbus.RedisConfig{
			DSN:             dsn,
			PoolSize:        cfg.GetInt("modules.event_bus.redis.pool_size"),
			MaxRetries:      cfg.GetInt("modules.event_bus.redis.max_retries"),
			MinRetryBackoff: cfg.GetDuration("modules.event_bus.redis.min_retry_backoff"),
			MaxRetryBackoff: cfg.GetDuration("modules.event_bus.redis.max_retry_backoff"),
			DialTimeout:     cfg.GetDuration("modules.event_bus.redis.dial_timeout"),
			ReadTimeout:     cfg.GetDuration("modules.event_bus.redis.read_timeout"),
			WriteTimeout:    cfg.GetDuration("modules.event_bus.redis.write_timeout"),
		},
		Consumer: eventbus.ConsumerConfig{
			Group:          cfg.GetString("modules.event_bus.consumer.group"),
			ConsumerID:     cfg.GetString("modules.event_bus.consumer.consumer_id"),
			BatchSize:      cfg.GetInt("modules.event_bus.consumer.batch_size"),
			BlockDuration:  cfg.GetDuration("modules.event_bus.consumer.block_duration"),
			MaxConcurrency: cfg.GetInt("modules.event_bus.consumer.max_concurrency"),
		},
	}

	sub, err := redis.NewSubscriber(ebConfig)
	if err != nil {
		return nil
	}

	return sub
}
```

### File: `internal/worker/subscriber/<event_name>_worker.go`

Each event stream you subscribe to gets its own worker file. Naming convention: `<stream>_<event_type>_worker.go`.

Example for `promotion.created`:

```go
package subscriber

import (
	"context"

	"github.com/ankorstore/yokai/config"
	"github.com/ankorstore/yokai/log"
	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	"github.com/tclavelloux/promy-event-bus/events"
)

// PromotionCreatedWorker subscribes to promotion.created events from Redis Streams.
type PromotionCreatedWorker struct {
	subscriber eventbus.EventSubscriber
	config     *config.Config
	// Inject your domain services here for business logic:
	// messageService message.Service
}

// NewPromotionCreatedWorker creates a new promotion.created event subscriber worker.
func NewPromotionCreatedWorker(
	subscriber eventbus.EventSubscriber,
	config *config.Config,
	// messageService message.Service,
) *PromotionCreatedWorker {
	return &PromotionCreatedWorker{
		subscriber: subscriber,
		config:     config,
	}
}

// Name returns the worker name. Convention: "<event-type>-subscriber".
func (w *PromotionCreatedWorker) Name() string {
	return "promotion-created-subscriber"
}

// Run starts the blocking subscription loop.
// If the subscriber is nil (Redis unavailable), the worker exits gracefully
// without error — this allows the service to start even without Redis.
func (w *PromotionCreatedWorker) Run(ctx context.Context) error {
	logger := log.CtxLogger(ctx)

	if w.subscriber == nil {
		logger.Warn().Msg("promotion-created subscriber disabled: Redis not available")
		return nil
	}

	logger.Info().Msg("starting promotion-created event subscriber")

	return w.subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
		Stream:         events.StreamPromotions,
		ConsumerGroup:  w.config.GetString("modules.event_bus.consumer.group"),
		ConsumerID:     w.config.GetString("modules.event_bus.consumer.consumer_id"),
		Handler:        w.handleEvent,
		BatchSize:      w.config.GetInt("modules.event_bus.consumer.batch_size"),
		BlockDuration:  w.config.GetDuration("modules.event_bus.consumer.block_duration"),
		MaxConcurrency: w.config.GetInt("modules.event_bus.consumer.max_concurrency"),
	})
}

// handleEvent processes a single event from the stream.
func (w *PromotionCreatedWorker) handleEvent(ctx context.Context, event eventbus.Event) error {
	logger := log.CtxLogger(ctx)
	logger.Info().
		Str("eventId", event.EventID()).
		Str("eventType", event.EventType()).
		Msg("processing promotion.created event")

	// Your business logic here.
	// See "Event Handler Implementation" section below for deserialization patterns.

	return nil
}
```

---

## FX Wiring & Registration

### File: `internal/infra/register.go`

Register both publisher and subscriber as FX-provided interfaces:

```go
package infra

import (
	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
	eb "github.com/<org>/<repo>/internal/infra/eventbus"
	"go.uber.org/fx"
)

func RegisterInfraComponents() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				eb.NewEventPublisher,
				fx.As(new(eventbus.EventPublisher)),
			),
			fx.Annotate(
				eb.NewEventSubscriber,
				fx.As(new(eventbus.EventSubscriber)),
			),
		),
	)
}
```

If your service only publishes (no subscriber), omit the `NewEventSubscriber` line.
If your service only subscribes (no publisher), omit the `NewEventPublisher` line.

### File: `internal/worker/register.go`

Register your worker(s) with the Yokai worker pool:

```go
package worker

import (
	"github.com/ankorstore/yokai/fxhealthcheck"
	"github.com/ankorstore/yokai/fxworker"
	"github.com/ankorstore/yokai/worker/healthcheck"
	"github.com/<org>/<repo>/internal/worker/subscriber"
	"go.uber.org/fx"
)

func RegisterWorkerComponents() fx.Option {
	return fx.Options(
		// Worker pool health check (exposes worker liveness)
		fxhealthcheck.AsCheckerProbe(healthcheck.NewWorkerProbe),
		// Register each subscriber worker
		fxworker.AsWorker(subscriber.NewPromotionCreatedWorker),
		// fxworker.AsWorker(subscriber.NewUserRegisteredWorker),  // add more as needed
	)
}
```

---

## Bootstrap & TestBootstrapper

This is a critical architectural decision: **workers MUST NOT run during tests**.

Workers try to connect to Redis and block indefinitely on `Subscribe()`. Tests have no Redis available (they use embedded in-memory MySQL, not Docker). If workers boot during tests, they either fail to connect (test hangs/crashes) or spin idle consuming resources.

### File: `internal/bootstrap.go`

```go
package internal

import (
	"github.com/ankorstore/yokai/fxworker"
	"github.com/<org>/<repo>/internal/worker"
	// ... other imports ...
)

// Bootstrapper is used in production — includes workers.
var Bootstrapper = fxcore.NewBootstrapper().WithOptions(
	fxsql.FxSQLModule,
	fxhttpserver.FxHttpServerModule,
	fxworker.FxWorkerModule,               // <-- PRODUCTION ONLY

	domain.RegisterDomainComponents(),
	infra.RegisterInfraComponents(),
	worker.RegisterWorkerComponents(),      // <-- PRODUCTION ONLY

	Register(),
	Router(),
)

// TestBootstrapper is used in tests — excludes workers.
var TestBootstrapper = fxcore.NewBootstrapper().WithOptions(
	fxsql.FxSQLModule,
	fxhttpserver.FxHttpServerModule,
	// NO fxworker.FxWorkerModule
	// NO worker.RegisterWorkerComponents()

	domain.RegisterDomainComponents(),
	infra.RegisterInfraComponents(),       // Still registers publisher/subscriber constructors
	                                        // They return nil gracefully (no Redis in test env)

	Register(),
	Router(),
)

// RunTest uses TestBootstrapper.
func RunTest(tb testing.TB, options ...fx.Option) {
	tb.Helper()
	// ... config, fxgomysqlserver, migrations, seeds ...
	TestBootstrapper.RunTestApp(tb, /* options */)
}
```

### Why `infra.RegisterInfraComponents()` stays in TestBootstrapper

Domain services depend on `eventbus.EventPublisher` via FX injection. If we remove `infra.RegisterInfraComponents()` from tests, FX fails with "missing type" errors. The publisher constructor returns `nil` when `modules.event_bus.redis.dsn` is empty (which it is in test config), so the nil-safe `PublishEvent()` helper handles it gracefully.

---

## Event Handler Implementation

When the subscriber's handler receives an event, it arrives as a `eventbus.Event` interface (specifically a `*rawEvent` from the Redis subscriber). The payload is a JSON string accessible via type assertion.

### Deserializing the event payload

The `rawEvent` struct (internal to promy-event-bus/redis) exposes the payload as a string field. To access it from your handler, assert to the raw interface and unmarshal:

```go
import (
	"encoding/json"

	promotion "github.com/tclavelloux/promy-event-bus/events/promotion"
)

func (w *PromotionCreatedWorker) handleEvent(ctx context.Context, event eventbus.Event) error {
	// The event's underlying data is JSON. Access the raw payload:
	type dataCarrier interface {
		EventType() string
		EventID() string
	}

	// For now, use the event metadata to route, then deserialize from
	// the raw Redis message. The promy-event-bus library passes a rawEvent
	// whose .data field contains the full JSON payload.
	//
	// Pattern: type-switch on EventType() for multi-event streams.
	switch event.EventType() {
	case events.EventPromotionCreated:
		return w.handlePromotionCreated(ctx, event)
	default:
		logger.Warn().Str("type", event.EventType()).Msg("unhandled event type")
		return nil  // Acknowledge unknown events to prevent redelivery
	}
}
```

> **Note**: The current `rawEvent` struct in promy-event-bus does not export the `.data` field. If you need to access the raw JSON payload for deserialization, you have two options:
> 1. Submit a PR to promy-event-bus adding a `Data() string` method to the `Event` interface
> 2. Perform your business logic based solely on the event metadata (ID, type, timestamp) and fetch the full entity from the source service API
>
> For MVP, option 2 is simpler — the handler knows the promotion ID from the event and can call promy-product's API if enrichment is needed.

---

## Testing Strategy

### Unit tests for event handlers

Test handlers in isolation by mocking dependencies. No Redis needed.

```go
package subscriber_test

func TestPromotionCreatedWorker_HandleEvent(t *testing.T) {
	// Mock your domain service (e.g., messageService)
	mockService := &messagemock.MockService{}
	mockService.On("Create", mock.Anything, mock.Anything).Return(nil)

	worker := subscriber.NewPromotionCreatedWorker(
		nil,  // subscriber is nil — we're testing the handler, not the connection
		cfg,
		mockService,
	)

	// Call the handler directly (export it or test via Run with a fake subscriber)
	// ...

	mockService.AssertExpectations(t)
}
```

### Integration tests (optional, requires Redis)

If you need end-to-end verification:
- Use `github.com/alicebob/miniredis/v2` for an in-process Redis mock
- Or use testcontainers for a real Redis

These tests should NOT be part of the standard `make test` pipeline. Gate them behind a build tag:

```go
//go:build integration

package integration_test
```

### What `make test` guarantees

- Handler logic is correct (unit tests with mocks)
- FX wiring compiles and resolves (TestBootstrapper boots without workers)
- No Redis connection is attempted (DSN is empty in test config, constructors return nil)

---

## Graceful Degradation

Redis is treated as an **optional dependency** — but this comes with an explicit tradeoff that you must understand before adopting it.

### Event Criticality Tiers

Not all events are equal. Before using the fire-and-forget pattern, classify your event:

| Tier | Guarantee | Loss acceptable? | Pattern | Examples |
|------|-----------|-------------------|---------|----------|
| **Best-effort** | Fire-and-forget. Event may be lost if Redis is down at publish time. | Yes | `PublishEvent()` goroutine (this guide) | Push notifications, analytics, cache invalidation |
| **Business-critical** | Event loss = data inconsistency or broken user contract. | No | Transactional Outbox or synchronous publish with request failure | Billing triggers, inter-service state sync, audit trails |

**All current promy events are best-effort (Tier 1).** Push notifications, promotion indexing, and analytics enrichment tolerate occasional loss — the user can refresh, and the data converges eventually.

**If you ever add a Tier 2 event** (e.g., a subscription payment trigger, a compliance audit event), **do NOT use the `PublishEvent()` helper**. Instead, write the event to an outbox table inside your business transaction, and have a separate worker drain the outbox to Redis. This is out of scope for now (Phase 7+) but this boundary must be respected.

### Rules (for best-effort events)

1. **Publisher constructor returns nil** when DSN is empty or Redis is unreachable
2. **Subscriber constructor returns nil** when DSN is empty or Redis is unreachable
3. **`PublishEvent()` checks for nil publisher** before calling `Publish()` — logs a warning and returns
4. **Workers check for nil subscriber** in `Run()` — log a warning and return nil (no error)
5. **Tests never have Redis** — the empty DSN in test config triggers nil returns automatically
6. **The service starts and serves HTTP normally** even if Redis is completely down

### Why this matters

- `make test` works without Docker (embedded MySQL only)
- Local development works without running promy-event-bus's Redis
- Production survives transient Redis outages (HTTP API stays up, events are skipped)
- Railway deploys succeed even if Redis is temporarily unreachable during boot

### What you lose (consciously)

- **No delivery guarantee** — if Redis is down at publish time, the event is gone (logged, not retried)
- **No backpressure** — under sustained Redis failure, goroutines accumulate (bounded by dial_timeout × request rate). Monitor publish error logs in production.
- **No replay** — there is no outbox, WAL, or DLQ to recover lost events from

These are acceptable tradeoffs for a side-project at current scale. Revisit if any event drives business-critical state transitions.

---

## Reference: Naming Conventions

### Package layout

```
internal/
├── infra/
│   ├── eventbus/              # Publisher + subscriber constructors
│   │   ├── publisher.go       # NewEventPublisher + PublishEvent helper
│   │   └── subscriber.go     # NewEventSubscriber
│   ├── register.go            # FX registration for all infra components
│   └── ...                    # Other infra (fcm/, brevo/, etc.)
├── worker/
│   ├── subscriber/            # One file per event worker
│   │   ├── promotion_created_worker.go
│   │   └── user_registered_worker.go
│   └── register.go            # FX registration for all workers
└── ...
```

### Naming rules

| Entity | Convention | Example |
|--------|-----------|---------|
| Package for event bus code | `internal/infra/eventbus/` | — |
| Package for workers | `internal/worker/subscriber/` | — |
| Worker struct name | `<EventType>Worker` | `PromotionCreatedWorker` |
| Worker constructor | `New<EventType>Worker` | `NewPromotionCreatedWorker` |
| Worker file name | `<event_type>_worker.go` | `promotion_created_worker.go` |
| Worker `Name()` return | `"<event-type>-subscriber"` | `"promotion-created-subscriber"` |
| Consumer group | `<service-name>-consumers` | `crm-service-consumers` |
| Consumer ID | `<service-name>-worker-<n>` | `crm-worker-1` |

### Note on promy-user's `tracking/` package

promy-user currently places publisher/subscriber under `internal/infra/tracking/`. This was an early naming choice. **New services should use `internal/infra/eventbus/`** for clarity. promy-user will be migrated to match in a future cleanup.

---

## Reference: Event Schemas

Defined in `github.com/tclavelloux/promy-event-bus/events/`:

### Streams

| Constant | Value | Domain |
|----------|-------|--------|
| `events.StreamPromotions` | `"events:promotions"` | Promotion lifecycle |
| `events.StreamProducts` | `"events:products"` | Product identification |
| `events.StreamUsers` | `"events:users"` | User lifecycle & preferences |

### Event Types

| Constant | Value | Publisher |
|----------|-------|-----------|
| `events.EventPromotionCreated` | `"promotion.created"` | promy-product |
| `events.EventPromotionUpdated` | `"promotion.updated"` | promy-product |
| `events.EventProductIdentified` | `"product.identified"` | promy-identifier |
| `events.EventUserRegistered` | `"user.registered"` | promy-user |
| `events.EventUserPreferencesUpdated` | `"user.preferences.updated"` | promy-user |
| `events.EventUserLocationUpdated` | `"user.location.updated"` | promy-user |

### Event Structs

Each event struct is in its domain package under `events/`:

- `events/promotion.PromotionCreatedEvent` — promotion ID, name, distributor, leaflet, price, dates, image
- `events/user.UserRegisteredEvent` — user ID, email
- `events/user.UserPreferenceUpdatedEvent` — user ID, distributors, categories
- `events/user.UserLocationUpdatedEvent` — user ID, latitude, longitude
- `events/product.ProductIdentifiedEvent` — promotion ID, product name, type ID, category ID, confidence

### Schema Evolution

All events carry a `version` field in metadata (currently hardcoded to `"1.0"`). When evolving event schemas:

1. **Additive changes** (new optional fields): safe — bump version to `"1.1"`, old consumers ignore unknown fields via `json:",omitempty"` and lenient unmarshalling
2. **Breaking changes** (removing/renaming fields, changing types): create a NEW event type (e.g., `promotion.created.v2`) rather than modifying the existing one. Old consumers continue processing `promotion.created` until migrated.
3. **Consumers MUST tolerate unknown fields** — use `json.Decoder` with no strict mode, or ignore unknown keys. Never fail on an unexpected field.

---

## Reference: Config Keys

Complete mapping between YAML path and `config.GetXxx()` calls:

| YAML Path | Go accessor | Type | Default |
|-----------|------------|------|---------|
| `modules.event_bus.redis.dsn` | `cfg.GetString(...)` | string | (empty = disabled) |
| `modules.event_bus.redis.pool_size` | `cfg.GetInt(...)` | int | 10 |
| `modules.event_bus.redis.max_retries` | `cfg.GetInt(...)` | int | 3 |
| `modules.event_bus.redis.min_retry_backoff` | `cfg.GetDuration(...)` | duration | 8ms |
| `modules.event_bus.redis.max_retry_backoff` | `cfg.GetDuration(...)` | duration | 512ms |
| `modules.event_bus.redis.dial_timeout` | `cfg.GetDuration(...)` | duration | 5s |
| `modules.event_bus.redis.read_timeout` | `cfg.GetDuration(...)` | duration | 3s |
| `modules.event_bus.redis.write_timeout` | `cfg.GetDuration(...)` | duration | 3s |
| `modules.event_bus.consumer.group` | `cfg.GetString(...)` | string | — |
| `modules.event_bus.consumer.consumer_id` | `cfg.GetString(...)` | string | — |
| `modules.event_bus.consumer.batch_size` | `cfg.GetInt(...)` | int | 10 |
| `modules.event_bus.consumer.block_duration` | `cfg.GetDuration(...)` | duration | 1s |
| `modules.event_bus.consumer.max_concurrency` | `cfg.GetInt(...)` | int | 5 |

---

## FAQ

### Q: Should I await Redis acknowledgement before returning 200 to the client?

**No.** The `PublishEvent()` helper uses a goroutine (fire-and-forget at the application layer). However, Redis XADD is synchronous at the transport layer — when `Publish()` returns nil inside the goroutine, the event IS durably persisted in the stream. The goroutine simply decouples the HTTP response from the Redis round-trip (~1-2ms).

**Tradeoff**: if Redis is unreachable at the exact moment of publish, the event is lost (logged but not retried). For our use case (push notifications, analytics side-effects), this is acceptable. A transactional outbox pattern would guarantee delivery but adds significant complexity — deferred to Phase 7+.

### Q: Are events delivered in order?

**Within a single consumer with `max_concurrency: 1`**, yes — Redis Streams preserves insertion order. **With `max_concurrency > 1`** (the default in this guide is 10), messages are dispatched to a worker pool and may be processed out of order.

If your handler requires strict ordering (e.g., `user.registered` must be processed before `user.preferences.updated` for the same user), either:
- Set `max_concurrency: 1` (trades throughput for ordering)
- Partition by entity ID in your handler logic (check that the entity exists before processing dependent events)

For most promy use cases, out-of-order processing is fine because handlers are idempotent and operate on independent entities.

### Q: Can I subscribe to multiple streams in one service?

**Yes.** Create one worker per stream (or per event type if you want finer granularity). Each worker independently calls `subscriber.Subscribe()` on its target stream. The Yokai worker pool runs them all concurrently.

### Q: What happens if my handler returns an error?

The promy-event-bus subscriber has built-in retry logic (3 attempts with exponential backoff: 0ms, 100ms, 500ms). After 3 failures, the message is **ACKed and dropped** — there is no dead-letter queue (DLQ is a future enhancement). This means a persistently failing message is permanently lost after 3 attempts.

**Consequence**: design your handlers to be idempotent AND to fail gracefully. If your handler depends on an external API that may be temporarily down, consider returning nil (skip) with a warning log rather than an error, to avoid exhausting retries on transient infrastructure issues.

### Q: How do I make my handler idempotent?

Use `event.EventID()` as a deduplication key. Before processing, check if you've already handled this event:

```go
func (w *PromotionCreatedWorker) handleEvent(ctx context.Context, event eventbus.Event) error {
    // Option A: "processed events" table with unique constraint on event_id
    if w.repository.EventAlreadyProcessed(ctx, event.EventID()) {
        return nil // skip duplicate
    }
    // ... process ...
    w.repository.MarkEventProcessed(ctx, event.EventID())
    return nil
}
```

For simpler cases (handler is naturally idempotent, e.g., upserting a record by external ID), explicit dedup may not be needed.

### Q: How do I add a new event type to the system?

1. Define the event struct in `promy-event-bus/events/<domain>/<event_type>.go`
2. Add the constant to `promy-event-bus/events/types.go`
3. Tag a new release of promy-event-bus
4. `go get github.com/tclavelloux/promy-event-bus@latest` in the publishing service
5. `go get github.com/tclavelloux/promy-event-bus@latest` in consuming service(s)

### Q: How do I know if events are being dropped in production?

Currently: only via error logs (`"failed to publish event"`). This is a known gap.

**Recommended future improvement** (not yet implemented): expose a Prometheus counter `eventbus_publish_errors_total{stream, event_type}` and alert when it exceeds a threshold. Until then, ensure your logging pipeline (e.g., Railway logs, Loki) has an alert rule on the error message pattern.

At minimum, periodically check `XLEN` on your streams to confirm they're growing as expected, and `XINFO GROUPS <stream>` to verify consumer lag isn't growing unbounded.

### Q: Why is the package called `eventbus` and not `tracking`?

`tracking` was an early naming choice in promy-user (the first service to integrate). `eventbus` is the canonical name going forward — it directly describes what the package does without implying a specific use case (analytics tracking vs. event-driven communication are both valid uses).
