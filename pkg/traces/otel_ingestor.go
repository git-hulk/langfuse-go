package traces

import (
	"context"
	"encoding/base64"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type OtelIngestor struct {
	provider *sdktrace.TracerProvider
	tracer   oteltrace.Tracer
}

type OtelIngestorOption func(*otelIngestorConfig)

type otelIngestorConfig struct {
	spanProcessor sdktrace.SpanProcessor
}

func WithSpanProcessor(sp sdktrace.SpanProcessor) OtelIngestorOption {
	return func(c *otelIngestorConfig) {
		c.spanProcessor = sp
	}
}

func NewOtelIngestor(host, publicKey, secretKey string, opts ...OtelIngestorOption) (*OtelIngestor, error) {
	cfg := &otelIngestorConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	var provider *sdktrace.TracerProvider
	if cfg.spanProcessor != nil {
		provider = sdktrace.NewTracerProvider(
			sdktrace.WithSpanProcessor(cfg.spanProcessor),
		)
	} else {
		authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(publicKey+":"+secretKey))
		exporter, err := otlptracehttp.New(
			context.Background(),
			otlptracehttp.WithEndpointURL(host+"/api/public/otel/v1/traces"),
			otlptracehttp.WithHeaders(map[string]string{
				"Authorization": authHeader,
			}),
		)
		if err != nil {
			return nil, err
		}
		provider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
		)
	}

	otel.SetTracerProvider(provider)
	tracer := provider.Tracer("langfuse-go")

	return &OtelIngestor{
		provider: provider,
		tracer:   tracer,
	}, nil
}

func (oi *OtelIngestor) startTrace(ctx context.Context, name string) *Trace {
	ctx, span := oi.tracer.Start(ctx, name)
	sc := span.SpanContext()
	return &Trace{
		handler:      oi,
		observations: make([]*Observation, 0),
		otelCtx:      ctx,
		TraceEntry: TraceEntry{
			ID:        sc.TraceID().String(),
			Name:      name,
			Timestamp: time.Now(),
		},
	}
}

func (oi *OtelIngestor) endTrace(t *Trace) {
	t.Latency = time.Since(t.Timestamp).Milliseconds()
	if t.otelCtx == nil {
		return
	}
	span := oteltrace.SpanFromContext(t.otelCtx)
	span.SetAttributes(traceAttributes(t)...)
	span.End()
}

func (oi *OtelIngestor) startObservation(t *Trace, name string, typ ObservationType) *Observation {
	parentCtx := oi.getParentContext(t)
	ctx, span := oi.tracer.Start(parentCtx, name)
	sc := span.SpanContext()
	return &Observation{
		TraceID:             t.ID,
		ID:                  sc.SpanID().String(),
		Name:                name,
		Type:                typ,
		ParentObservationID: t.getParentObservationID(),
		StartTime:           time.Now(),
		handler:             oi,
		otelCtx:             ctx,
	}
}

func (oi *OtelIngestor) endObservation(o *Observation) {
	now := time.Now()
	o.EndTime = &now
	if o.otelCtx == nil {
		return
	}
	span := oteltrace.SpanFromContext(o.otelCtx)
	span.SetAttributes(observationAttributes(o)...)
	span.End()
}

func (oi *OtelIngestor) getParentContext(t *Trace) context.Context {
	if len(t.observations) == 0 {
		return t.otelCtx
	}
	last := t.observations[len(t.observations)-1]
	if last.EndTime == nil || last.EndTime.IsZero() {
		return last.otelCtx
	}
	return oi.findParentContext(t, last)
}

func (oi *OtelIngestor) findParentContext(t *Trace, obs *Observation) context.Context {
	for i := len(t.observations) - 1; i >= 0; i-- {
		o := t.observations[i]
		if o.ID == obs.ParentObservationID {
			return o.otelCtx
		}
	}
	return t.otelCtx
}

func (oi *OtelIngestor) flush() {
	oi.provider.ForceFlush(context.Background())
}

func (oi *OtelIngestor) close() error {
	return oi.provider.Shutdown(context.Background())
}

func (oi *OtelIngestor) StartTrace(ctx context.Context, name string) *Trace {
	return oi.startTrace(ctx, name)
}

func (oi *OtelIngestor) Flush() {
	oi.flush()
}

func (oi *OtelIngestor) Close() error {
	return oi.close()
}
