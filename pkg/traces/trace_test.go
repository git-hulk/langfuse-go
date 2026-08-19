package traces

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrace_StartSpan(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-traces")

	span := trace.StartSpan("test-span")

	require.NotNil(t, span)
	assert.Equal(t, "test-span", span.Name)
	assert.Equal(t, ObservationTypeSpan, span.Type)

	spanID := span.SpanContext().SpanID().String()
	assert.NotEmpty(t, spanID)
	assert.Len(t, spanID, 16)
	assert.Regexp(t, "^[0-9a-f]{16}$", spanID)

	assert.Equal(t, span, trace.lastObservation)
}

func TestTrace_MultipleSpans(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-traces")

	span1 := trace.StartSpan("span-1")
	span2 := trace.StartSpan("span-2")

	assert.Equal(t, span2, trace.lastObservation)
	assert.Equal(t, "span-1", span1.Name)
	assert.Equal(t, "span-2", span2.Name)
	assert.NotEqual(t, span1.SpanContext().SpanID(), span2.SpanContext().SpanID())
}

func TestTrace_Fields(t *testing.T) {
	trace := &Trace{
		TraceEntry: TraceEntry{
			ID:          "test-id",
			Name:        "test-name",
			SessionID:   "session-123",
			Release:     "v1.0.0",
			Version:     "1.0",
			UserID:      "user-456",
			Metadata:    map[string]any{"key": "value"},
			Tags:        []string{"tag1", "tag2"},
			Environment: "test",
		},
	}

	assert.Equal(t, "test-id", trace.ID)
	assert.Equal(t, "test-name", trace.Name)
	assert.Equal(t, "session-123", trace.SessionID)
	assert.Equal(t, "v1.0.0", trace.Release)
	assert.Equal(t, "1.0", trace.Version)
	assert.Equal(t, "user-456", trace.UserID)
	assert.Equal(t, map[string]any{"key": "value"}, trace.Metadata)
	assert.Equal(t, []string{"tag1", "tag2"}, trace.Tags)
	assert.Equal(t, "test", trace.Environment)
}

func TestTrace_NestedSpans(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-trace")

	parentSpan := trace.StartSpan("parent-span")
	childSpan := trace.StartSpan("child-span")
	childSpan2 := trace.StartSpan("child-span-2")

	assert.Equal(t, childSpan2, trace.lastObservation)
	assert.NotEqual(t, parentSpan.SpanContext().SpanID(), childSpan.SpanContext().SpanID())
	assert.NotEqual(t, childSpan.SpanContext().SpanID(), childSpan2.SpanContext().SpanID())
	assert.NotEqual(t, parentSpan.SpanContext().SpanID(), childSpan2.SpanContext().SpanID())
}

func TestTrace_NestedSpansWithEndedSpans(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-trace")

	trace.StartSpan("parent-span")

	childSpan := trace.StartSpan("child-span")
	childSpan.End()

	siblingSpan := trace.StartSpan("sibling-span")

	assert.Equal(t, siblingSpan, trace.lastObservation)
}

func TestTrace_DeepNestedSpans(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-trace")

	level1 := trace.StartSpan("level-1")
	level2 := trace.StartSpan("level-2")
	level3 := trace.StartSpan("level-3")
	level4 := trace.StartSpan("level-4")

	assert.Equal(t, level4, trace.lastObservation)

	spanIDs := make(map[string]bool)
	for _, obs := range []*Observation{level1, level2, level3, level4} {
		id := obs.SpanContext().SpanID().String()
		assert.False(t, spanIDs[id], "Found duplicate span ID: %s", id)
		spanIDs[id] = true
	}
}

func TestTrace_StartObservation(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-trace")

	observation := trace.StartObservation("test-observation", ObservationTypeAgent)

	require.NotNil(t, observation)
	assert.Equal(t, "test-observation", observation.Name)
	assert.Equal(t, ObservationTypeAgent, observation.Type)
	assert.NotEmpty(t, observation.SpanContext().SpanID().String())

	assert.Equal(t, observation, trace.lastObservation)

	observation2 := trace.StartObservation("test-observation-2", ObservationTypeTool)

	require.NotNil(t, observation2)
	assert.Equal(t, "test-observation-2", observation2.Name)
	assert.Equal(t, ObservationTypeTool, observation2.Type)
	assert.NotEqual(t, observation.SpanContext().SpanID(), observation2.SpanContext().SpanID())

	assert.Equal(t, observation2, trace.lastObservation)
}

func TestTrace_StartGeneration(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-trace")

	generation := trace.StartGeneration("test-generation")

	require.NotNil(t, generation)
	assert.Equal(t, "test-generation", generation.Name)
	assert.Equal(t, ObservationTypeGeneration, generation.Type)
	assert.NotEmpty(t, generation.SpanContext().SpanID().String())

	assert.Equal(t, generation, trace.lastObservation)

	generation2 := trace.StartGeneration("test-generation-2")
	observation := trace.StartObservation("test-observation", ObservationTypeGeneration)

	assert.Equal(t, generation2.Type, observation.Type)
	assert.Equal(t, ObservationTypeGeneration, generation2.Type)
}
