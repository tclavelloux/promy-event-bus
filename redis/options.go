package redis

import (
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// instrumentationName is the OpenTelemetry tracer name of this package.
const instrumentationName = "github.com/tclavelloux/promy-event-bus/redis"

// Metadata keys carrying the W3C trace context.
const (
	metaTraceparent = "traceparent"
	metaTracestate  = "tracestate"
)

// options holds the observability settings shared by Publisher and Subscriber.
type options struct {
	logger     zerolog.Logger
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

// Option customizes a Publisher or a Subscriber.
type Option func(*options)

// WithLogger sets the logger used by the Subscriber. A nil logger is ignored.
//
// It has no effect on the Publisher: publish errors are returned, never logged.
//
// The logger is the default sink of SubscriptionConfig.OnError:
//   - OnError nil: each swallowed error is logged (Warn for metadata_unmarshal, Error otherwise).
//   - OnError set: OnError is called and the logger is not used for that error.
//
// One log bypasses OnError: an event dropped after max attempts because no DLQPublisher
// is configured is always logged at Warn.
//
// Default: zerolog.Nop().
func WithLogger(logger *zerolog.Logger) Option {
	return func(o *options) {
		if logger != nil {
			o.logger = *logger
		}
	}
}

// WithTracerProvider sets the tracer provider used for publish and process spans. A nil provider is ignored.
//
// Default: otel.GetTracerProvider() at construction. The global provider delegates, so a later
// otel.SetTracerProvider call still takes effect.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(o *options) {
		if tp != nil {
			o.tracer = tp.Tracer(instrumentationName)
		}
	}
}

func newOptions(opts []Option) options {
	o := options{
		logger:     zerolog.Nop(),
		tracer:     otel.GetTracerProvider().Tracer(instrumentationName),
		propagator: propagation.TraceContext{},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	return o
}
