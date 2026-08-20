package traces

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func newTestIngestor(t *testing.T) (*Ingestor, *tracetest.InMemoryExporter) {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	sp := sdktrace.NewSimpleSpanProcessor(exporter)
	ingestor, err := NewIngestor("http://localhost", "pk", "sk", WithSpanProcessor(sp))
	require.NoError(t, err)
	return ingestor, exporter
}

func TestNewIngestor_DoesNotOverrideGlobalProvider(t *testing.T) {
	// Simulate a host application (or APM) that has already installed a global
	// TracerProvider. Creating a Langfuse ingestor must leave it untouched.
	sentinel := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(sentinel)

	sp := sdktrace.NewSimpleSpanProcessor(tracetest.NewInMemoryExporter())
	ingestor, err := NewIngestor("http://localhost", "pk", "sk", WithSpanProcessor(sp))
	require.NoError(t, err)
	defer ingestor.Close()

	assert.Same(t, sentinel, otel.GetTracerProvider(),
		"NewIngestor must not override the global TracerProvider")
}

func TestIngestor_StartTrace(t *testing.T) {
	ingestor, _ := newTestIngestor(t)
	defer ingestor.Close()

	ctx, trace := ingestor.StartTrace(context.Background(), "test-trace")

	require.NotNil(t, ctx)
	require.NotNil(t, trace)
	assert.Equal(t, "test-trace", trace.Name)
	assert.NotEmpty(t, trace.ID)
	assert.Len(t, trace.ID, 32)
	assert.Equal(t, trace.SpanContext().SpanID(), oteltrace.SpanFromContext(ctx).SpanContext().SpanID())
}

func TestIngestor_TraceUniqueIDs(t *testing.T) {
	ingestor, _ := newTestIngestor(t)
	defer ingestor.Close()

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		_, trace := ingestor.StartTrace(context.Background(), "test")
		require.False(t, ids[trace.ID], "duplicate trace ID: %s", trace.ID)
		ids[trace.ID] = true
		trace.End()
	}
}

func TestIngestor_EndTrace_ExportsSpan(t *testing.T) {
	ingestor, exporter := newTestIngestor(t)
	defer ingestor.Close()

	_, trace := ingestor.StartTrace(context.Background(), "exported-trace")
	trace.Input = "hello"
	trace.Output = "world"
	trace.UserID = "user-1"
	trace.SessionID = "sess-1"
	trace.Tags = []string{"a", "b"}
	trace.Environment = "test"
	trace.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "exported-trace", span.Name)

	attrMap := make(map[string]any)
	for _, a := range span.Attributes {
		attrMap[string(a.Key)] = a.Value.AsInterface()
	}
	assert.Equal(t, "exported-trace", attrMap[AttrTraceName])
	assert.Equal(t, "hello", attrMap[AttrTraceInput])
	assert.Equal(t, "world", attrMap[AttrTraceOutput])
	assert.Equal(t, "user-1", attrMap[AttrUserID])
	assert.Equal(t, "sess-1", attrMap[AttrSessionID])
	assert.Equal(t, "test", attrMap[AttrEnvironment])
}

func TestIngestor_StartObservation(t *testing.T) {
	ingestor, _ := newTestIngestor(t)
	defer ingestor.Close()

	ctx, trace := ingestor.StartTrace(context.Background(), "trace")
	obsCtx, obs := trace.StartObservation(ctx, "my-agent", ObservationTypeAgent)

	require.NotNil(t, obs)
	assert.Equal(t, "my-agent", obs.Name)
	assert.Equal(t, ObservationTypeAgent, obs.Type)
	obsID := obs.SpanContext().SpanID().String()
	assert.NotEmpty(t, obsID)
	assert.Len(t, obsID, 16)
	assert.Equal(t, obs.SpanContext().SpanID(), oteltrace.SpanFromContext(obsCtx).SpanContext().SpanID())
}

func TestIngestor_ObservationParentChild(t *testing.T) {
	ingestor, exporter := newTestIngestor(t)
	defer ingestor.Close()

	ctx, trace := ingestor.StartTrace(context.Background(), "trace")

	_, span1 := trace.StartSpan(ctx, "span-1")
	_, span2 := trace.StartSpan(ctx, "span-2")

	span2.End()
	span1.End()
	trace.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 3)

	spanByName := make(map[string]tracetest.SpanStub)
	for _, s := range spans {
		spanByName[s.Name] = s
	}
	traceSpanID := spanByName["trace"].SpanContext.SpanID()
	assert.Equal(t, traceSpanID, spanByName["span-1"].Parent.SpanID())
	assert.Equal(t, traceSpanID, spanByName["span-2"].Parent.SpanID())
}

func TestIngestor_NestedObservations(t *testing.T) {
	ingestor, exporter := newTestIngestor(t)
	defer ingestor.Close()

	ctx, trace := ingestor.StartTrace(context.Background(), "trace")

	agentCtx, agent := trace.StartObservation(ctx, "agent", ObservationTypeAgent)
	_, tool := trace.StartSpan(agentCtx, "tool")

	tool.End()
	agent.End()
	trace.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 3)

	spanByName := make(map[string]tracetest.SpanStub)
	for _, s := range spans {
		spanByName[s.Name] = s
	}
	assert.Equal(t, spanByName["trace"].SpanContext.SpanID(), spanByName["agent"].Parent.SpanID())
	assert.Equal(t, spanByName["agent"].SpanContext.SpanID(), spanByName["tool"].Parent.SpanID())
}

func TestIngestor_Generation(t *testing.T) {
	ingestor, exporter := newTestIngestor(t)
	defer ingestor.Close()

	ctx, trace := ingestor.StartTrace(context.Background(), "trace")
	_, gen := trace.StartGeneration(ctx, "llm-call")
	gen.Model = "gpt-4"
	gen.Input = map[string]string{"prompt": "hello"}
	gen.Output = map[string]string{"response": "hi"}
	gen.Usage = Usage{Input: 10, Output: 20, Total: 30, Unit: UnitTokens}
	gen.End()
	trace.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 2)

	var genSpan tracetest.SpanStub
	for _, s := range spans {
		if s.Name == "llm-call" {
			genSpan = s
			break
		}
	}
	require.NotEmpty(t, genSpan.Name)

	attrMap := make(map[string]any)
	for _, a := range genSpan.Attributes {
		attrMap[string(a.Key)] = a.Value.AsInterface()
	}
	assert.Equal(t, "generation", attrMap[AttrObservationType])
	assert.Equal(t, "gpt-4", attrMap[AttrObservationModelName])
	assert.Contains(t, attrMap[AttrObservationInput], "hello")
	assert.Contains(t, attrMap[AttrObservationOutput], "hi")
	assert.Contains(t, attrMap[AttrObservationUsageDetails], "TOKENS")
}

func TestIngestor_FlushAndClose(t *testing.T) {
	ingestor, exporter := newTestIngestor(t)

	_, trace := ingestor.StartTrace(context.Background(), "flush-test")
	trace.End()

	ingestor.Flush()
	assert.NotEmpty(t, exporter.GetSpans())

	err := ingestor.Close()
	require.NoError(t, err)
}

func TestIngestor_EndedSpanParenting(t *testing.T) {
	ingestor, exporter := newTestIngestor(t)
	defer ingestor.Close()

	ctx, trace := ingestor.StartTrace(context.Background(), "trace")

	_, first := trace.StartSpan(ctx, "first")
	first.End()

	_, second := trace.StartSpan(ctx, "second")
	second.End()
	trace.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 3)

	spanByName := make(map[string]tracetest.SpanStub)
	for _, s := range spans {
		spanByName[s.Name] = s
	}
	assert.Equal(t, spanByName["trace"].SpanContext.SpanID(), spanByName["second"].Parent.SpanID())
}
