package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	eventbus "github.com/tclavelloux/promy-event-bus/eventbus"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	fieldMetadata = "metadata"
	fieldPayload  = "payload"
)

// Publisher implements EventPublisher for Redis Streams.
type Publisher struct {
	client *redis.Client
	config eventbus.RedisConfig
	obs    options
}

// NewPublisher creates a new Redis publisher. Options: WithTracerProvider (WithLogger has no effect here).
func NewPublisher(config eventbus.RedisConfig, opts ...Option) (*Publisher, error) {
	redisOpts, err := redis.ParseURL(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis DSN: %w", err)
	}

	// Apply configuration
	if config.PoolSize > 0 {
		redisOpts.PoolSize = config.PoolSize
	}
	if config.MaxRetries > 0 {
		redisOpts.MaxRetries = config.MaxRetries
	}
	if config.MinRetryBackoff > 0 {
		redisOpts.MinRetryBackoff = config.MinRetryBackoff
	}
	if config.MaxRetryBackoff > 0 {
		redisOpts.MaxRetryBackoff = config.MaxRetryBackoff
	}
	if config.DialTimeout > 0 {
		redisOpts.DialTimeout = config.DialTimeout
	}
	if config.ReadTimeout > 0 {
		redisOpts.ReadTimeout = config.ReadTimeout
	}
	if config.WriteTimeout > 0 {
		redisOpts.WriteTimeout = config.WriteTimeout
	}

	client := redis.NewClient(redisOpts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Publisher{
		client: client,
		config: config,
		obs:    newOptions(opts),
	}, nil
}

// Publish publishes a single event to Redis Streams.
func (p *Publisher) Publish(ctx context.Context, stream string, event eventbus.Event) (err error) {
	// Validate struct tags first (required fields, formats, constraints)
	if err := eventbus.ValidateStruct(event); err != nil {
		return err
	}

	// Validate event business rules (event-specific validation)
	if err := event.Validate(); err != nil {
		return err
	}

	ctx, span := p.startPublishSpan(ctx, stream, event)
	defer span.End()
	defer func() { failSpan(span, err) }()

	values, err := p.buildValues(ctx, event)
	if err != nil {
		return err
	}

	if err := p.client.XAdd(ctx, &redis.XAddArgs{Stream: stream, Values: values}).Err(); err != nil {
		return fmt.Errorf("%w: %w", eventbus.ErrPublishFailed, err)
	}

	return nil
}

// PublishBatch publishes multiple events in a pipeline.
func (p *Publisher) PublishBatch(ctx context.Context, stream string, events []eventbus.Event) error {
	if len(events) == 0 {
		return nil
	}

	// Validate all events first before publishing batch
	if err := validateEvents(events); err != nil {
		return err
	}

	pipe := p.client.Pipeline()
	spans := make([]trace.Span, 0, len(events))
	cmds := make([]*redis.StringCmd, 0, len(events))

	// Every span started is ended exactly once, on every return path.
	defer func() {
		for _, span := range spans {
			span.End()
		}
	}()

	failAll := func(err error) {
		for _, span := range spans {
			failSpan(span, err)
		}
	}

	for _, event := range events {
		spanCtx, span := p.startPublishSpan(ctx, stream, event)
		spans = append(spans, span)

		values, err := p.buildValues(spanCtx, event)
		if err != nil {
			failAll(err)

			return err
		}

		cmds = append(cmds, pipe.XAdd(ctx, &redis.XAddArgs{Stream: stream, Values: values}))
	}

	if _, err := pipe.Exec(ctx); err != nil {
		failCommandSpans(spans, cmds)

		return fmt.Errorf("%w: %w", eventbus.ErrPublishFailed, err)
	}

	return nil
}

// validateEvents runs struct-tag and business validation on every event.
func validateEvents(events []eventbus.Event) error {
	for _, event := range events {
		if err := eventbus.ValidateStruct(event); err != nil {
			return err
		}

		if err := event.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// failCommandSpans fails the span of every command that returned an error.
func failCommandSpans(spans []trace.Span, cmds []*redis.StringCmd) {
	for i, cmd := range cmds {
		failSpan(spans[i], cmd.Err())
	}
}

// startPublishSpan starts the producer span of event.
func (p *Publisher) startPublishSpan(ctx context.Context, stream string, event eventbus.Event) (context.Context, trace.Span) {
	return p.obs.tracer.Start(ctx, stream+".publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "redis"),
			attribute.String("messaging.destination.name", stream),
			attribute.String("messaging.message.id", event.EventID()),
			attribute.String("messaging.operation.type", "send"),
		),
	)
}

// buildValues builds the XADD values (metadata + payload) of event, injecting the trace context of ctx.
func (p *Publisher) buildValues(ctx context.Context, event eventbus.Event) (map[string]any, error) {
	metadata := map[string]any{
		"id":        event.EventID(),
		"type":      event.EventType(),
		"timestamp": event.EventTime().Format(time.RFC3339),
		"version":   "1.0",
		"attempt":   1,
	}

	carrier := propagation.MapCarrier{}
	p.obs.propagator.Inject(ctx, carrier)

	for k, v := range carrier {
		metadata[k] = v
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	payloadJSON, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return map[string]any{
		fieldMetadata: string(metadataJSON),
		fieldPayload:  string(payloadJSON),
	}, nil
}

// failSpan marks span as failed when err is non-nil.
func failSpan(span trace.Span, err error) {
	if err == nil {
		return
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// Close closes the Redis connection.
func (p *Publisher) Close() error {
	return p.client.Close()
}

// Health checks Redis connection health.
func (p *Publisher) Health(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}
