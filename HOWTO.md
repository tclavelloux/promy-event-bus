# HOWTO: Integrating promy-event-bus in a Yokai Service

> **Version**: 2.2 - publish-before-commit pattern for Tier 1 (May 2026)
>
> **Changelog:**
> - v2.2: Tier 1 publish pattern documented as "publish-before-commit" (publish inside open DB transaction, commit only after Redis confirms); replaces naive synchronous publish that had a partial-failure window; transactional outbox deferred to Phase 7
> - v2.1: Stream ownership map - consumers column removed (each service manages its own subscriptions independently); `events:products` ownership corrected to `promy-product`; `events:identifications` added as `promy-identifier`'s own stream; single-owner rule made explicit
> - v2.0: `promy-event-bus` scope reduced to infrastructure only - event payload structs moved to each service; `Data() string` added to `Event` interface; tiers renamed (Tier 1 = critical, Tier 2 = best-effort); DLQ introduced; consumer DTOs replace shared event structs; per-stream config structure cleaned up

---

## Table of Contents

1. [Scope Split: What promy-event-bus Owns vs. What Your Service Owns](#scope-split)
2. [Stream Ownership Map](#stream-ownership-map)
3. [Prerequisites](#prerequisites)
4. [Dependency Installation](#dependency-installation)
5. [Configuration](#configuration)
6. [Docker Compose Setup](#docker-compose-setup)
7. [Event Criticality Tiers](#event-criticality-tiers)
8. [Publisher Integration](#publisher-integration)
9. [Subscriber Integration](#subscriber-integration)
10. [Multi-Stream Worker Topology](#multi-stream-worker-topology)
11. [FX Wiring & Registration](#fx-wiring--registration)
12. [Bootstrap & TestBootstrapper](#bootstrap--testbootstrapper)
13. [Testing Strategy](#testing-strategy)
14. [Graceful Degradation](#graceful-degradation)
15. [Reference: Naming Conventions](#reference-naming-conventions)
16. [Reference: Config Keys](#reference-config-keys)
17. [FAQ](#faq)

---

## Scope Split

This is the most important section to understand before writing any code.

### What `promy-event-bus` owns

`promy-event-bus` is a **pure infrastructure library**. It owns:

| Responsibility | Examples |
|---|---|
| Redis connection management | Publisher, Subscriber, connection pool config |
| The `Event` interface | `EventID()`, `EventType()`, `Data() string` (see prerequisite below) |
| Stream name constants | `StreamUsers`, `StreamPromotions`, etc. |
| Naming conventions | Consumer group pattern, consumer ID pattern (documented, not enforced) |

`promy-event-bus` does **NOT** own:

- Event payload struct definitions (e.g., `UserRegisteredEvent`, `PromotionCreatedEvent`)
- Event type string constants (e.g., `"user.registered"`, `"promotion.created"`)
- Any business logic

### What each service owns

Each service owns its **domain events** - the schema and the constants:

```
promy-user/
  internal/
    events/
      types.go              <-- const EventUserRegistered = "user.registered"
      user_registered.go    <-- struct UserRegisteredEvent { UserID, Email, ... }
      user_prefs_updated.go <-- struct UserPrefsUpdatedEvent { ... }

promy-product/
  internal/
    events/
      types.go
      promotion_created.go
      promotion_updated.go
      product_created.go

promy-subscription/
  internal/
    events/
      types.go
      subscription_started.go
      subscription_cancelled.go

promy-identifier/
  internal/
    events/
      types.go
      product_identified.go  <-- output of ML/LLM identification pipeline
```

### What each consumer owns

Each consuming service defines its **own DTOs** for the events it cares about.
It does NOT import event structs from the producing service.

```
promy-crm/
  internal/
    dto/
      user_registered_dto.go   <-- CRM's own view of a user.registered payload
      promotion_created_dto.go <-- CRM's own view of a promotion.created payload
```

**Why not import from the producer?**
Importing `promy-user`'s event struct in `promy-crm` creates a hard dependency:
every schema change in `promy-user` forces a `promy-crm` update, even for fields
CRM does not use. Consumer DTOs only declare the fields the consumer actually needs -
unknown fields are silently ignored during JSON unmarshalling.

### [WARNING] Prerequisite: `Data() string` must be added to the `Event` interface

The current `Event` interface in `promy-event-bus` does not expose the raw JSON payload.
**This must be fixed before any subscriber can deserialize event data.**

```go
// promy-event-bus/eventbus/event.go -- required change
type Event interface {
    EventID()   string
    EventType() string
    Timestamp() time.Time
    Data()      string  // ADD THIS: returns the raw JSON payload string
}
```

Open a PR on `promy-event-bus` with this change before integrating any subscriber
that needs to act on event content (which is most of them).

---

## Stream Ownership Map

**One stream, one owner.** Each stream is owned by exactly one service.
Only the owning service publishes to it. Which services consume a given stream
is each consuming service's own concern - it is not tracked here.

| Stream | Owner (Producer) | Default Tier |
|---|---|---|
| `events:users` | `promy-user` | Tier 1 + Tier 2 (mixed - see Event Criticality Tiers) |
| `events:subscriptions` | `promy-subscription` | Tier 1 |
| `events:promotions` | `promy-product` | Tier 2 |
| `events:products` | `promy-product` | Tier 2 |
| `events:identifications` | `promy-identifier` | Tier 2 |
| `events:dlq` | Any service (failure path only) | n/a |

**Rules:**
- A service publishes **only to its own stream**. Never to another service's stream.
- **One stream, one owner** - two services must never publish to the same stream. If two services produce events in the same domain, each gets its own stream (see FAQ: "Can two services share a stream?").
- The `events:dlq` stream is the only exception: any service may publish to it when a Tier 1 event handler exhausts its retries (see Event Criticality Tiers).
- When you add a new stream, add it to this table and open a PR on `promy-event-bus` to register its name constant.

---

## Prerequisites

- Your service is built with Yokai (`fxcore`, `fxhttpserver`, `fxsql`, etc.)
- The shared Redis instance is running via `promy-event-bus`'s docker-compose (`eventbus-network`)
- You have identified whether your service is a **publisher**, a **subscriber**, or **both**
- You have identified the **criticality tier** of each event your service produces (see Event Criticality Tiers)
- **`promy-event-bus` has `Data() string` on its `Event` interface** - do not start subscriber implementation without this

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

All event bus configuration lives under `modules.event_bus` in `configs/config.yaml`.

```yaml
modules:
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
      consumer_id: ${RAILWAY_REPLICA_ID:-${HOSTNAME:-<your-service>-worker-1}}
      defaults:
        batch_size: 50
        block_duration: 2s
        max_concurrency: 10
      streams:
        # Per-stream overrides -- only specify what differs from defaults
        events:users:
          max_concurrency: 1   # strict ordering: user.registered before user.preferences.updated
        # events:promotions:   # inherits defaults, no override needed
```

### Key rules

| Key | Convention | Example |
|---|---|---|
| `consumer.group` | `<service-name>-consumers` | `crm-service-consumers` |
| `consumer.consumer_id` | Railway replica ID -> hostname -> static fallback | `${RAILWAY_REPLICA_ID:-${HOSTNAME:-crm-worker-1}}` |
| `consumer.defaults.*` | Applies to all streams unless overridden | n/a |
| `consumer.streams.<name>.*` | Per-stream overrides | `events:users.max_concurrency: 1` |

### Why `RAILWAY_REPLICA_ID` first?

The consumer ID **must be unique per replica**. Two consumers sharing the same ID
in the same group means Redis delivers messages to only one of them - the other sits idle.

- **Railway**: use `${RAILWAY_REPLICA_ID}` - stable, unique per replica, survives restarts
- **Kubernetes**: use `${HOSTNAME}` (pod name) - unique per pod
- **Local dev**: the static fallback is fine (single replica)

### `.env` file additions

```env
REDIS_TRACKING_HOST=promy-redis
REDIS_TRACKING_PORT=6379
REDIS_TRACKING_PASSWORD=
REDIS_TRACKING_DB=0
# RAILWAY_REPLICA_ID is injected automatically by Railway in production
```

---

## Docker Compose Setup

```yaml
services:
  <your-service>-app-server:
    networks:
      - <your-service>-app-network
      - eventbus-network
    volumes:
      - .:/app
      - ../promy-event-bus:/promy-event-bus

networks:
  <your-service>-app-network:
    driver: bridge
  eventbus-network:
    external: true   # owned by promy-event-bus/docker-compose.yaml
```

Start `promy-event-bus` first so the network exists:
```bash
cd ../promy-event-bus && docker compose up -d
```

---

## Event Criticality Tiers

Before writing any producer or consumer code, classify every event your service
emits into one of two tiers. The tier determines the publish pattern and the
subscriber's failure handling.

| | **Tier 1 - Business-Critical** | **Tier 2 - Best-Effort** |
|---|---|---|
| **Definition** | Event loss = broken user contract or data inconsistency | Event loss = degraded experience, self-healing |
| **Acceptable loss?** | No | Yes |
| **Publish pattern** | Synchronous (blocks until Redis confirms) | Fire-and-forget goroutine |
| **Subscriber on failure** | Route to `events:dlq` after retries exhausted | Log warning, drop |
| **Examples** | `subscription.started`, `user.registered` (onboarding) | Push notifications, analytics, search indexing |

### Event tier reference

| Event | Stream | Tier | Rationale |
|---|---|---|---|
| `user.registered` | `events:users` | **Tier 1** | Drives onboarding - welcome email must be sent |
| `user.preferences.updated` | `events:users` | Tier 2 | Personalization - stale for a moment is acceptable |
| `user.location.updated` | `events:users` | Tier 2 | High-frequency, self-correcting |
| `subscription.started` | `events:subscriptions` | **Tier 1** | Triggers access grant and welcome flow |
| `subscription.cancelled` | `events:subscriptions` | **Tier 1** | Triggers access revocation |
| `promotion.created` | `events:promotions` | Tier 2 | Push notification - user can discover promotion later |
| `promotion.updated` | `events:promotions` | Tier 2 | Cache invalidation - next fetch self-heals |
| `product.created` | `events:products` | Tier 2 | Enrichment pipeline trigger |
| `product.identified` | `events:identifications` | Tier 2 | ML output - missing identification caught on next pass |

### [WARNING] Rule: Tier 1 events must not use `PublishEvent()` (fire-and-forget)

The `PublishEvent()` helper (see Publisher Integration) is **Tier 2 only**.
For Tier 1 events, use the synchronous publish path and handle errors explicitly.

```go
// Tier 2 -- fire-and-forget, loss acceptable
eb.PublishEvent(ctx, s.publisher, promystreams.StreamPromotions, event)

// Tier 1 -- publish-before-commit (see pattern below)
if err := s.publisher.Publish(ctx, promystreams.StreamSubscriptions, event); err != nil {
    return fmt.Errorf("failed to publish critical event, rolling back: %w", err)
}
```

### Tier 1 pattern: Publish-Before-Commit

For Tier 1 events, use the **publish-before-commit** pattern to avoid the partial-failure
window where the DB commits but the event is lost.

```go
func (s *DefaultService) StartSubscription(ctx context.Context, sub *Subscription) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    if err := s.repository.CreateWithTx(ctx, tx, sub); err != nil {
        return err
    }

    // Publish BEFORE commit -- if Redis fails, the DB transaction rolls back cleanly.
    // No data is persisted, no event is orphaned. The caller gets an error and can retry.
    event := localevents.NewSubscriptionStartedEvent(sub.ID, sub.UserID)
    if err := s.publisher.Publish(ctx, promystreams.StreamSubscriptions, event); err != nil {
        return fmt.Errorf("publish failed, rolling back: %w", err)
    }

    // Only commit after the event is durably persisted in Redis.
    return tx.Commit()
}
```

**Failure matrix:**

| Publish | Commit | Outcome | Acceptable? |
|---|---|---|---|
| Fails | Not attempted | Clean rollback, caller gets error, can retry | Yes |
| Succeeds | Fails | Orphaned event (entity doesn't exist yet) | Rare; consumers must handle "entity not found" gracefully |
| Succeeds | Succeeds | Happy path | Yes |

**Tradeoffs:**
- You hold a DB transaction open during a Redis round-trip (~1-5ms). Negligible at current scale.
- The "Commit fails after Publish" edge case is extremely rare (disk full, connection drop mid-commit). Consumers that encounter a missing entity should log a warning and skip -- the event is effectively a no-op.
- Long-term: a transactional outbox (Phase 7) eliminates even this edge case.

**[WARNING] Rule:** Never use `PublishEvent()` (fire-and-forget) for Tier 1 events. Always use the synchronous publish-before-commit pattern above.

---

## Publisher Integration

### Defining events in your service

Event structs and type constants live in your service, not in `promy-event-bus`.

```
promy-user/
  internal/
    events/
      types.go
      user_registered.go
      user_prefs_updated.go
```

**`internal/events/types.go`** - type constants owned by the producer:
```go
package events

const (
    EventUserRegistered         = "user.registered"
    EventUserPreferencesUpdated = "user.preferences.updated"
    EventUserLocationUpdated    = "user.location.updated"
)
```

**`internal/events/user_registered.go`** - event struct owned by the producer:
```go
package events

import (
    "encoding/json"
    "time"

    "github.com/google/uuid"
)

type UserRegisteredEvent struct {
    id        string
    userID    string
    email     string
    createdAt time.Time
}

func NewUserRegisteredEvent(userID, email string) *UserRegisteredEvent {
    return &UserRegisteredEvent{
        id:        uuid.NewString(),
        userID:    userID,
        email:     email,
        createdAt: time.Now(),
    }
}

// Implement eventbus.Event interface
func (e *UserRegisteredEvent) EventID()   string    { return e.id }
func (e *UserRegisteredEvent) EventType() string    { return EventUserRegistered }
func (e *UserRegisteredEvent) Timestamp() time.Time { return e.createdAt }
func (e *UserRegisteredEvent) Data()      string {
    b, _ := json.Marshal(map[string]string{
        "user_id": e.userID,
        "email":   e.email,
    })
    return string(b)
}
```

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

// NewEventPublisher creates a publisher from configuration.
// Returns nil if Redis DSN is empty (graceful degradation for test/dev environments).
func NewEventPublisher(cfg *config.Config) eventbus.EventPublisher {
    dsn := cfg.GetString("modules.event_bus.redis.dsn")
    if dsn == "" {
        return nil
    }

    publisher, err := redis.NewPublisher(eventbus.RedisConfig{
        DSN:             dsn,
        PoolSize:        cfg.GetInt("modules.event_bus.redis.pool_size"),
        MaxRetries:      cfg.GetInt("modules.event_bus.redis.max_retries"),
        MinRetryBackoff: cfg.GetDuration("modules.event_bus.redis.min_retry_backoff"),
        MaxRetryBackoff: cfg.GetDuration("modules.event_bus.redis.max_retry_backoff"),
        DialTimeout:     cfg.GetDuration("modules.event_bus.redis.dial_timeout"),
        ReadTimeout:     cfg.GetDuration("modules.event_bus.redis.read_timeout"),
        WriteTimeout:    cfg.GetDuration("modules.event_bus.redis.write_timeout"),
    })
    if err != nil {
        return nil
    }
    return publisher
}

// PublishEvent publishes a Tier 2 (best-effort) event asynchronously.
//
// [WARNING] DO NOT use this for Tier 1 (business-critical) events.
// For Tier 1, call publisher.Publish() synchronously and handle the error.
//
// KNOWN LIMITATION: under sustained Redis outage at high request rate, goroutines
// accumulate (bounded by dial_timeout x req/s). Acceptable at current scale.
// Future improvement: replace with a bounded channel + single background drainer.
func PublishEvent(ctx context.Context, publisher eventbus.EventPublisher, stream string, event eventbus.Event) {
    logger := log.CtxLogger(ctx)

    if publisher == nil {
        logger.Warn().
            Str("eventId", event.EventID()).
            Str("eventType", event.EventType()).
            Msg("publisher not available, skipping event publication")
        return
    }

    //nolint:contextcheck,gosec
    go func() {
        if err := publisher.Publish(context.Background(), stream, event); err != nil {
            logger.Error().Err(err).
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
    "fmt"

    eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
    promystreams "github.com/tclavelloux/promy-event-bus/streams"
    localevents "github.com/tclavelloux/promy-user/internal/events"
    eb "github.com/<org>/promy-user/internal/infra/eventbus"
)

type DefaultService struct {
    repository Repository
    publisher  eventbus.EventPublisher
}

func (s *DefaultService) Register(ctx context.Context, user *User) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    if err := s.repository.CreateWithTx(ctx, tx, user); err != nil {
        return err
    }

    // user.registered is Tier 1 -- publish-before-commit pattern
    event := localevents.NewUserRegisteredEvent(user.ID, user.Email)
    if err := s.publisher.Publish(ctx, promystreams.StreamUsers, event); err != nil {
        return fmt.Errorf("publish failed, rolling back: %w", err)
    }

    return tx.Commit()
}

func (s *DefaultService) UpdateLocation(ctx context.Context, userID string, lat, lng float64) error {
    if err := s.repository.UpdateLocation(ctx, userID, lat, lng); err != nil {
        return err
    }

    // user.location.updated is Tier 2 -- fire-and-forget
    event := localevents.NewUserLocationUpdatedEvent(userID, lat, lng)
    eb.PublishEvent(ctx, s.publisher, promystreams.StreamUsers, event)

    return nil
}
```

---

## Subscriber Integration

### Architecture: Yokai Worker pattern

Subscribers run as **Yokai Workers** (`fxworker`). This is not optional.
Workers are long-running goroutines managed by the Yokai worker pool:
integrated with health checks, started on boot, stopped on shutdown,
and excluded from tests via `TestBootstrapper`.

### Consumer DTOs - define what you need, nothing more

Each consuming service defines its own structs for the event payloads it uses.
Do not import event structs from the producing service.

```
promy-crm/
  internal/
    dto/
      user_registered_dto.go
      subscription_started_dto.go
      promotion_created_dto.go
```

**`internal/dto/user_registered_dto.go`:**
```go
package dto

// UserRegisteredDTO is CRM's view of a user.registered event payload.
// Only declare fields CRM actually needs -- unknown fields are safely ignored
// during JSON unmarshalling.
type UserRegisteredDTO struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
}
```

### File: `internal/infra/eventbus/subscriber.go`

```go
package eventbus

import (
    "github.com/ankorstore/yokai/config"
    eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
    "github.com/tclavelloux/promy-event-bus/redis"
)

// NewEventSubscriber creates a subscriber from configuration.
// Returns nil if Redis DSN is empty (graceful degradation).
func NewEventSubscriber(cfg *config.Config) eventbus.EventSubscriber {
    dsn := cfg.GetString("modules.event_bus.redis.dsn")
    if dsn == "" {
        return nil
    }

    sub, err := redis.NewSubscriber(eventbus.Config{
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
            Group:      cfg.GetString("modules.event_bus.consumer.group"),
            ConsumerID: cfg.GetString("modules.event_bus.consumer.consumer_id"),
        },
    })
    if err != nil {
        return nil
    }
    return sub
}
```

### File: `internal/worker/subscriber/<stream>_worker.go`

One worker file per stream. The worker handles ALL event types on its stream
and dispatches internally. See Multi-Stream Worker Topology for the rationale.

```go
package subscriber

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/ankorstore/yokai/config"
    "github.com/ankorstore/yokai/log"
    eventbus "github.com/tclavelloux/promy-event-bus/eventbus"
    promystreams "github.com/tclavelloux/promy-event-bus/streams"
    "github.com/<org>/promy-crm/internal/dto"
)

// UsersWorker subscribes to all events on events:users.
type UsersWorker struct {
    subscriber eventbus.EventSubscriber
    config     *config.Config
    // emailService email.Service  <-- inject domain services here
}

func NewUsersWorker(
    subscriber eventbus.EventSubscriber,
    config *config.Config,
) *UsersWorker {
    return &UsersWorker{subscriber: subscriber, config: config}
}

func (w *UsersWorker) Name() string { return "users-subscriber" }

func (w *UsersWorker) Run(ctx context.Context) error {
    logger := log.CtxLogger(ctx)

    if w.subscriber == nil {
        logger.Warn().Msg("users-subscriber disabled: Redis not available")
        return nil
    }

    logger.Info().Msg("starting users event subscriber")

    return w.subscriber.Subscribe(ctx, eventbus.SubscriptionConfig{
        Stream:         promystreams.StreamUsers,
        ConsumerGroup:  w.config.GetString("modules.event_bus.consumer.group"),
        ConsumerID:     w.config.GetString("modules.event_bus.consumer.consumer_id"),
        Handler:        w.handleEvent,
        BatchSize:      w.config.GetInt("modules.event_bus.consumer.defaults.batch_size"),
        BlockDuration:  w.config.GetDuration("modules.event_bus.consumer.defaults.block_duration"),
        // Per-stream override: events:users requires strict ordering
        MaxConcurrency: w.config.GetInt("modules.event_bus.consumer.streams.events:users.max_concurrency"),
    })
}

// handleEvent dispatches by event type.
// ALWAYS return nil for unknown or unimplemented event types -- never return
// an error for an event you do not handle, as it triggers retries and
// eventually DLQ routing for an event you intentionally ignore.
func (w *UsersWorker) handleEvent(ctx context.Context, event eventbus.Event) error {
    switch event.EventType() {
    case "user.registered":
        return w.handleUserRegistered(ctx, event)
    case "user.preferences.updated":
        // Not yet implemented -- ACK silently
        return nil
    default:
        log.CtxLogger(ctx).Warn().
            Str("type", event.EventType()).
            Msg("unknown user event type -- ACKing to prevent redelivery")
        return nil
    }
}

func (w *UsersWorker) handleUserRegistered(ctx context.Context, event eventbus.Event) error {
    // Deserialize using CRM's own DTO -- not promy-user's struct
    var payload dto.UserRegisteredDTO
    if err := json.Unmarshal([]byte(event.Data()), &payload); err != nil {
        return fmt.Errorf("failed to deserialize user.registered payload: %w", err)
    }

    // Your business logic here
    // return w.emailService.SendWelcomeEmail(ctx, payload.Email)
    return nil
}
```

---

## Multi-Stream Worker Topology

### Rule 1: One worker per stream

`Subscribe()` is a **blocking call**. You cannot call it twice in the same `Run()`.
Each stream needs its own worker (one goroutine per stream in the Yokai worker pool).

```
internal/worker/subscriber/
    users_worker.go         <-- subscribes to events:users
    promotions_worker.go    <-- subscribes to events:promotions
    subscriptions_worker.go <-- subscribes to events:subscriptions
```

**Why NOT one worker per event type on the same stream?**

If `UserRegisteredWorker` and `UserPrefsWorker` both subscribe to `events:users`
with the same consumer group, Redis distributes messages between them randomly.
Each worker receives ~50% of messages - including messages of the wrong type
that it discards. Half your events are silently no-op'd. This is wrong.

[OK] **Correct**: one worker owns the full stream, dispatches via `switch event.EventType()`.

### Rule 2: One consumer group per service, reused across all streams

The group name is the same for all workers in the same service.
On different streams, the same group name creates independent groups in Redis -
they do not interfere.

```yaml
consumer:
  group: crm-service-consumers   # same name for events:users, events:promotions, etc.
```

Fan-out between services works because each service uses a **different group name**:

```
events:promotions
    +-- group: crm-service-consumers     -> promy-crm gets every message
    +-- group: product-service-consumers -> promy-product gets every message (independently)
```

### Rule 3: One consumer ID per replica, shared across all workers

Two workers in the same process share the same consumer ID. This is correct:
the `(stream, group, consumerID)` triple is what Redis uses to track state.
Workers on different streams with the same consumer ID are independent.

```
promy-crm replica A (consumer_id: crm-abc)
  +-- UsersWorker         -> (events:users,         crm-service-consumers, crm-abc)
  +-- PromotionsWorker    -> (events:promotions,     crm-service-consumers, crm-abc)  <-- different stream, safe
  +-- SubscriptionsWorker -> (events:subscriptions,  crm-service-consumers, crm-abc)
```

### Rule 4: `max_concurrency: 1` for streams requiring event ordering

With `max_concurrency > 1`, messages in the same batch are processed concurrently
and may complete out of order. For `events:users`, this matters:

```
user.registered  must be processed BEFORE  user.preferences.updated
```

Set `max_concurrency: 1` on `events:users` in your config. The throughput cost
is negligible at current scale.

### Rule 5: Error handling by tier

| Event tier | Handler returns error | Outcome |
|---|---|---|
| Tier 1 | `error` | Subscriber retries (3x), then routes to `events:dlq` |
| Tier 2 | `error` | Subscriber retries (3x), then drops (logs warning) |
| Any tier | `nil` for unknown type | ACKed immediately, no retry |

For Tier 1 events, route to DLQ after exhausting retries:

```go
func (w *SubscriptionsWorker) handleEvent(ctx context.Context, event eventbus.Event) error {
    err := w.processEvent(ctx, event)
    if err != nil {
        // Tier 1: route to DLQ -- returning nil ACKs the original message.
        // The DLQ entry is now the retry surface for ops/automated replay.
        w.dlqPublisher.Publish(context.Background(), "events:dlq", wrapDLQ(event, err))
        return nil
    }
    return nil
}
```

### Topology diagram

```
promy-crm (1 replica, consumer_id: crm-abc)
+--------------------------------------------------------------------+
|  Yokai Worker Pool                                                 |
|                                                                    |
|  UsersWorker         PromotionsWorker       SubscriptionsWorker    |
|  stream:events:users  stream:events:promo   stream:events:subs     |
|  concurrency: 1       concurrency: 10       concurrency: 5         |
|                                                                    |
|  Shared *redis.Subscriber (one TCP connection pool)                |
+----------+-------------------+----------------------+--------------+
           | XREADGROUP        | XREADGROUP           | XREADGROUP
           v                   v                      v
+--------------------------------------------------------------------+
|  Redis                                                             |
|  events:users              events:promotions  events:subscriptions |
|  +-- crm-svc-consumers     +-- crm-svc-cons.  +-- crm-svc-cons.   |
|                             +-- product-svc-consumers              |
+--------------------------------------------------------------------+
```

---

## FX Wiring & Registration

### File: `internal/infra/register.go`

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
            fx.Annotate(eb.NewEventPublisher, fx.As(new(eventbus.EventPublisher))),
            fx.Annotate(eb.NewEventSubscriber, fx.As(new(eventbus.EventSubscriber))),
        ),
    )
}
```

Omit `NewEventSubscriber` if your service only publishes; omit `NewEventPublisher` if it only subscribes.

### File: `internal/worker/register.go`

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
        fxhealthcheck.AsCheckerProbe(healthcheck.NewWorkerProbe),
        fxworker.AsWorker(subscriber.NewUsersWorker),
        fxworker.AsWorker(subscriber.NewPromotionsWorker),
        fxworker.AsWorker(subscriber.NewSubscriptionsWorker),
    )
}
```

---

## Bootstrap & TestBootstrapper

Workers **must not run during tests**. They try to connect to Redis and block
on `Subscribe()`. Tests have no Redis available.

```go
// internal/bootstrap.go

// Bootstrapper -- production (includes workers)
var Bootstrapper = fxcore.NewBootstrapper().WithOptions(
    fxsql.FxSQLModule,
    fxhttpserver.FxHttpServerModule,
    fxworker.FxWorkerModule,               // PRODUCTION ONLY

    domain.RegisterDomainComponents(),
    infra.RegisterInfraComponents(),
    worker.RegisterWorkerComponents(),      // PRODUCTION ONLY

    Register(),
    Router(),
)

// TestBootstrapper -- tests (workers excluded)
var TestBootstrapper = fxcore.NewBootstrapper().WithOptions(
    fxsql.FxSQLModule,
    fxhttpserver.FxHttpServerModule,
    // NO fxworker.FxWorkerModule
    // NO worker.RegisterWorkerComponents()

    domain.RegisterDomainComponents(),
    infra.RegisterInfraComponents(),  // publisher/subscriber return nil gracefully (no DSN in test config)

    Register(),
    Router(),
)
```

---

## Testing Strategy

### Unit tests for event handlers

Test handlers directly - no Redis needed, no worker bootstrapping.

```go
package subscriber_test

func TestUsersWorker_HandleUserRegistered(t *testing.T) {
    mockEmailService := &emailmock.MockService{}
    mockEmailService.On("SendWelcomeEmail", mock.Anything, "thomas@example.com").Return(nil)

    worker := subscriber.NewUsersWorker(nil, cfg, mockEmailService)

    // Construct a fake event using a test helper implementing the Event interface
    event := testhelper.NewFakeEvent("user.registered", `{"user_id":"u-1","email":"thomas@example.com"}`)

    err := worker.HandleEventForTest(ctx, event)  // export handler for testing
    require.NoError(t, err)
    mockEmailService.AssertExpectations(t)
}
```

### What `make test` guarantees

- Handler logic is correct (unit tests with mocks)
- FX wiring compiles and resolves (TestBootstrapper boots without workers)
- No Redis connection is attempted (empty DSN -> nil constructors -> graceful skip)

---

## Graceful Degradation

Redis is an **optional dependency at boot time** for Tier 2 events.
For Tier 1 events, Redis unavailability at publish time causes the request to fail
(this is intentional - the caller must know the event was not delivered).

| Scenario | Tier 1 behavior | Tier 2 behavior |
|---|---|---|
| Redis DSN empty (test/dev) | Publisher returns nil, `Publish()` not called, request proceeds | `PublishEvent()` no-ops with warning log |
| Redis unreachable at publish | `Publish()` returns error, request returns 500 | Fire-and-forget goroutine logs error, drops event |
| Redis unreachable at subscribe | Worker logs warning, exits `Run()` cleanly | Same |
| Redis recovers | Next publish/subscribe succeeds automatically | Same |

---

## Reference: Naming Conventions

### Package layout

```
internal/
+-- events/                    # Producer only: event structs + type constants
|   +-- types.go
|   +-- user_registered.go
|   +-- ...
+-- dto/                       # Consumer only: deserialization structs
|   +-- user_registered_dto.go
|   +-- ...
+-- infra/
|   +-- eventbus/
|   |   +-- publisher.go       # NewEventPublisher + PublishEvent helper
|   |   +-- subscriber.go      # NewEventSubscriber
|   +-- register.go
+-- worker/
    +-- subscriber/
    |   +-- users_worker.go
    |   +-- promotions_worker.go
    |   +-- ...
    +-- register.go
```

### Naming rules

| Entity | Convention | Example |
|---|---|---|
| Event type constant | `"<domain>.<verb>"` in past tense | `"user.registered"` |
| Stream constant (in promy-event-bus) | `Stream<Domain>` | `StreamUsers` |
| Stream value | `"events:<domain>"` | `"events:users"` |
| Consumer group | `<service-name>-consumers` | `crm-service-consumers` |
| Consumer ID | `${RAILWAY_REPLICA_ID:-${HOSTNAME:-<svc>-worker-1}}` | `crm-abc123` |
| Worker struct | `<StreamDomain>Worker` | `UsersWorker` |
| Worker constructor | `New<StreamDomain>Worker` | `NewUsersWorker` |
| Worker file | `<stream_domain>_worker.go` | `users_worker.go` |
| Worker `Name()` | `"<stream-domain>-subscriber"` | `"users-subscriber"` |
| Producer event package | `internal/events/` | n/a |
| Consumer DTO package | `internal/dto/` | n/a |

---

## Reference: Config Keys

| YAML Path | Go accessor | Type |
|---|---|---|
| `modules.event_bus.redis.dsn` | `cfg.GetString(...)` | string |
| `modules.event_bus.redis.pool_size` | `cfg.GetInt(...)` | int |
| `modules.event_bus.redis.max_retries` | `cfg.GetInt(...)` | int |
| `modules.event_bus.redis.min_retry_backoff` | `cfg.GetDuration(...)` | duration |
| `modules.event_bus.redis.max_retry_backoff` | `cfg.GetDuration(...)` | duration |
| `modules.event_bus.redis.dial_timeout` | `cfg.GetDuration(...)` | duration |
| `modules.event_bus.redis.read_timeout` | `cfg.GetDuration(...)` | duration |
| `modules.event_bus.redis.write_timeout` | `cfg.GetDuration(...)` | duration |
| `modules.event_bus.consumer.group` | `cfg.GetString(...)` | string |
| `modules.event_bus.consumer.consumer_id` | `cfg.GetString(...)` | string |
| `modules.event_bus.consumer.defaults.batch_size` | `cfg.GetInt(...)` | int |
| `modules.event_bus.consumer.defaults.block_duration` | `cfg.GetDuration(...)` | duration |
| `modules.event_bus.consumer.defaults.max_concurrency` | `cfg.GetInt(...)` | int |
| `modules.event_bus.consumer.streams.<name>.max_concurrency` | `cfg.GetInt(...)` | int |
| `modules.event_bus.consumer.streams.<name>.batch_size` | `cfg.GetInt(...)` | int |

---

## FAQ

### Q: Can I use `PublishEvent()` for a `subscription.started` event?

**No.** `subscription.started` is Tier 1. Use synchronous `publisher.Publish()` and
return the error to the caller if it fails. The user's HTTP request should return 500
rather than silently losing a business-critical event.

### Q: Where do event type string constants live now?

In the **producing service**, under `internal/events/types.go`. They are not in
`promy-event-bus`. If you are a consumer, hardcode the string in your switch statement
(e.g., `case "user.registered":`) or define your own local constants - do not
import from the producing service.

### Q: Are events delivered in order?

Within a single consumer with `max_concurrency: 1`, yes. With higher concurrency,
messages in the same batch may complete out of order. Use `max_concurrency: 1`
for streams where event ordering matters (see `events:users`).

### Q: What happens if my handler returns an error?

The subscriber retries 3 times with exponential backoff. After 3 failures:
- **Tier 1 event**: your handler should route to `events:dlq` and return `nil`
- **Tier 2 event**: the subscriber drops the message and logs a warning

Design your handlers to be **idempotent** - they may be called more than once
for the same event (Redis delivers at-least-once).

### Q: How do I make my handler idempotent?

Use `event.EventID()` as a deduplication key stored in your DB with a unique constraint.
Check before processing; skip (return nil) if already seen.

### Q: How do I add a new event type?

1. Add the event struct to `internal/events/` in the **producing service**
2. Add the type constant to `internal/events/types.go` in the producing service
3. Publish using the new struct - no changes to `promy-event-bus` needed
4. Notify consuming teams of the new event type and its payload schema
5. Consuming services add a new `case` in their worker's `handleEvent` switch
   (and add a DTO if they need the payload)

### Q: How do I add a new stream?

1. Add a row to the Stream Ownership Map
2. Open a PR on `promy-event-bus` to add the stream name constant (e.g., `StreamMyDomain = "events:mydomain"`)
3. Implement producer and/or subscriber as described in this guide

### Q: Can two services share a stream (multiple producers)?

**No.** One stream, one owner. If `promy-identifier` and `promy-product` both need
to publish product-related events, they each own their own stream:

```
events:products        -> owned by promy-product    (product.created, product.updated)
events:identifications -> owned by promy-identifier (product.identified)
```

Consumers subscribe to whichever streams they need independently.
Mixing two producers on one stream breaks ownership accountability:
there is no single team responsible for the stream's contract, schema evolution,
or incident response.

### Q: How do I know if Tier 1 events are being lost in production?

Monitor the `events:dlq` stream length (`XLEN events:dlq`) and alert when it grows.
Also monitor error logs for `"failed to publish event"` (Tier 1 synchronous path
returns 500, which is already surfaced in HTTP metrics).