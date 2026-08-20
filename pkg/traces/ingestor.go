package traces

import (
	"context"
	"encoding/base64"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Ingestor struct {
	provider *sdktrace.TracerProvider
	tracer   tracer
}

type IngestorOption func(*ingestorConfig)

type ingestorConfig struct {
	spanProcessor sdktrace.SpanProcessor
}

func WithSpanProcessor(sp sdktrace.SpanProcessor) IngestorOption {
	return func(c *ingestorConfig) {
		c.spanProcessor = sp
	}
}

func NewIngestor(host, publicKey, secretKey string, opts ...IngestorOption) (*Ingestor, error) {
	cfg := &ingestorConfig{}
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
	t := provider.Tracer("langfuse-go")

	return &Ingestor{
		provider: provider,
		tracer:   t,
	}, nil
}

// StartTrace creates a new trace rooted at a new OTel span.
//
// Returns the context carrying the trace span and the Trace itself. Pass the returned
// context to the trace's StartXXX methods to nest observations under this trace.
func (oi *Ingestor) StartTrace(ctx context.Context, name string) (context.Context, *Trace) {
	ctx, span := oi.tracer.Start(ctx, name)
	sc := span.SpanContext()
	return ctx, &Trace{
		Span:   span,
		tracer: oi.tracer,
		TraceEntry: TraceEntry{
			ID:   sc.TraceID().String(),
			Name: name,
		},
	}
}

func (oi *Ingestor) Flush() {
	oi.provider.ForceFlush(context.Background())
}

func (oi *Ingestor) Close() error {
	return oi.provider.Shutdown(context.Background())
}
