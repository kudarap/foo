package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	otellogs "github.com/dbdoyc/opentelemetry-logs-go"
	"github.com/dbdoyc/opentelemetry-logs-go/logs"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
)

// TraceLogger injects log tracer on slog.Handler.
// Only use this after initializing open-telemetry provider.
func TraceLogger(l *slog.Logger) *slog.Logger {
	h := l.Handler()
	return slog.New(NewLogHandler(h))
}

const (
	instrumentationName    = "otelslog"
	instrumentationVersion = "0.0.1"
)

type LogHandler struct {
	handler slog.Handler
	logger  logs.Logger
}

// NewLogHandler returns a LogHandler with the given level.
// All methods except Enabled delegate to h.
func NewLogHandler(sh slog.Handler) *LogHandler {
	var h LogHandler
	h.handler = sh
	h.logger = otellogs.GetLoggerProvider().Logger(
		instrumentationName,
		logs.WithInstrumentationVersion(instrumentationVersion),
		logs.WithSchemaURL(semconv.SchemaURL),
	)
	return &h
}

func (h *LogHandler) Handle(ctx context.Context, record slog.Record) error {
	if ctx == nil {
		return h.handler.Handle(ctx, record)
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return h.handler.Handle(ctx, record)
	}
	spanCtx := span.SpanContext()
	if !spanCtx.HasTraceID() {
		return h.handler.Handle(ctx, record)
	}

	traceID := spanCtx.TraceID()
	spanID := spanCtx.SpanID()
	traceFlags := spanCtx.TraceFlags()

	levelStr := record.Level.String()
	levelNum := logs.SeverityNumber(record.Level)
	config := logs.LogRecordConfig{
		Timestamp:         &record.Time,
		ObservedTimestamp: time.Now(),
		TraceId:           &traceID,
		SpanId:            &spanID,
		TraceFlags:        &traceFlags,
		SeverityText:      &levelStr,
		SeverityNumber:    &levelNum,
		Body:              &record.Message,
		InstrumentationScope: &instrumentation.Scope{
			Name:    instrumentationName,
			Version: instrumentationVersion,
		},
	}
	attrs := make([]attribute.KeyValue, 0, record.NumAttrs()+2+3)
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, otelAttr(a))
		return true
	})
	config.Attributes = &attrs

	logRecord := logs.NewLogRecord(config)
	h.logger.Emit(logRecord)

	// Experimental: stdout log tracing
	record.Add("trace_id", traceID)
	return h.handler.Handle(ctx, record)
}

// Enabled implements Handler.Enabled.
func (h *LogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// WithAttrs implements Handler.WithAttrs.
func (h *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LogHandler{
		handler: h.handler.WithAttrs(attrs),
		logger:  h.logger,
	}
}

// WithGroup implements Handler.WithGroup.
func (h *LogHandler) WithGroup(name string) slog.Handler {
	return &LogHandler{
		handler: h.handler.WithGroup(name),
		logger:  h.logger,
	}
}

func otelAttr(attr slog.Attr) attribute.KeyValue {
	key, val := attr.Key, attr.Value
	switch val.Kind() {
	case slog.KindBool:
		return attribute.Bool(key, val.Bool())
	case slog.KindDuration:
		return attribute.Int64(key+"_ns", val.Duration().Nanoseconds())
	case slog.KindFloat64:
		return attribute.Float64(key, val.Float64())
	case slog.KindInt64:
		return attribute.Int64(key, val.Int64())
	case slog.KindString:
		return attribute.String(key, val.String())
	case slog.KindLogValuer:
		return attribute.Stringer(key, val.LogValuer().LogValue())
	}

	return attribute.String(key, fmt.Sprint(val))
}
