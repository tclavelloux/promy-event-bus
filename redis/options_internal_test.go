package redis

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestOptions_NilKeepsDefaults(t *testing.T) {
	base := newOptions(nil)

	got := newOptions([]Option{WithLogger(nil), WithTracerProvider(nil), nil})

	assert.Equal(t, base.logger.GetLevel(), got.logger.GetLevel())
	assert.Equal(t, zerolog.Nop().GetLevel(), got.logger.GetLevel())
	assert.NotNil(t, got.tracer)
	assert.NotNil(t, got.propagator)
}

func TestOptions_Applied(t *testing.T) {
	logger := zerolog.New(nil).Level(zerolog.ErrorLevel)
	tp := sdktrace.NewTracerProvider()
	defer func() { _ = tp.Shutdown(t.Context()) }()

	got := newOptions([]Option{WithLogger(&logger), WithTracerProvider(tp)})

	assert.Equal(t, zerolog.ErrorLevel, got.logger.GetLevel())
	assert.Equal(t, tp.Tracer(instrumentationName), got.tracer)
}
