package traces

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func newTestOtelIngestor(t *testing.T) (*OtelIngestor, *tracetest.InMemoryExporter) {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	sp := sdktrace.NewSimpleSpanProcessor(exporter)
	ingestor, err := NewOtelIngestor("http://localhost", "pk", "sk", WithSpanProcessor(sp))
	require.NoError(t, err)
	return ingestor, exporter
}

func TestOtelIngestor_StartTrace(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-trace")

	require.NotNil(t, trace)
	assert.Equal(t, "test-trace", trace.Name)
	assert.NotEmpty(t, trace.ID)
	assert.Len(t, trace.ID, 32)
	assert.NotNil(t, trace.otelCtx)
	assert.Empty(t, trace.observations)
}

func TestOtelIngestor_TraceUniqueIDs(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		trace := ingestor.StartTrace(context.Background(), "test")
		require.False(t, ids[trace.ID], "duplicate trace ID: %s", trace.ID)
		ids[trace.ID] = true
		trace.End()
	}
}

func TestOtelIngestor_EndTrace_ExportsSpan(t *testing.T) {
	ingestor, exporter := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "exported-trace")
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

func TestOtelIngestor_StartObservation(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "trace")
	obs := trace.StartObservation("my-agent", ObservationTypeAgent)

	require.NotNil(t, obs)
	assert.Equal(t, "my-agent", obs.Name)
	assert.Equal(t, ObservationTypeAgent, obs.Type)
	assert.NotEmpty(t, obs.ID)
	assert.Len(t, obs.ID, 16)
	assert.NotNil(t, obs.otelCtx)
}

func TestOtelIngestor_ObservationParentChild(t *testing.T) {
	ingestor, exporter := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "trace")

	span1 := trace.StartSpan("span-1")
	span2 := trace.StartSpan("span-2")

	span2.End()
	span1.End()
	trace.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 3)

	spanByName := make(map[string]tracetest.SpanStub)
	for _, s := range spans {
		spanByName[s.Name] = s
	}
	assert.Equal(t, spanByName["trace"].SpanContext.SpanID(), spanByName["span-1"].Parent.SpanID())
	assert.Equal(t, spanByName["span-1"].SpanContext.SpanID(), spanByName["span-2"].Parent.SpanID())
}

func TestOtelIngestor_EndedSpanParenting(t *testing.T) {
	ingestor, exporter := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "trace")

	parent := trace.StartSpan("parent")
	child := trace.StartSpan("child")
	child.End()

	sibling := trace.StartSpan("sibling")
	sibling.End()
	parent.End()
	trace.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 4)

	spanByName := make(map[string]tracetest.SpanStub)
	for _, s := range spans {
		spanByName[s.Name] = s
	}
	assert.Equal(t, spanByName["parent"].SpanContext.SpanID(), spanByName["sibling"].Parent.SpanID())
}

func TestOtelIngestor_Generation(t *testing.T) {
	ingestor, exporter := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "trace")
	gen := trace.StartGeneration("llm-call")
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

func TestOtelIngestor_FlushAndClose(t *testing.T) {
	ingestor, exporter := newTestOtelIngestor(t)

	trace := ingestor.StartTrace(context.Background(), "flush-test")
	trace.End()

	ingestor.Flush()
	assert.NotEmpty(t, exporter.GetSpans())

	err := ingestor.Close()
	require.NoError(t, err)
}

func TestOtelIngestor_DeepNesting(t *testing.T) {
	ingestor, exporter := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "trace")

	l1 := trace.StartSpan("level-1")
	l2 := trace.StartSpan("level-2")
	l3 := trace.StartSpan("level-3")

	l3.End()
	l2.End()
	l1.End()
	trace.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 4)

	spanByName := make(map[string]tracetest.SpanStub)
	for _, s := range spans {
		spanByName[s.Name] = s
	}

	assert.Equal(t, spanByName["trace"].SpanContext.SpanID(), spanByName["level-1"].Parent.SpanID())
	assert.Equal(t, spanByName["level-1"].SpanContext.SpanID(), spanByName["level-2"].Parent.SpanID())
	assert.Equal(t, spanByName["level-2"].SpanContext.SpanID(), spanByName["level-3"].Parent.SpanID())
}
