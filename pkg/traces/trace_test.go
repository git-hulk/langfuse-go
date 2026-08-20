package traces

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrace_StartSpan(t *testing.T) {
	ingestor, _ := newTestIngestor(t)
	defer ingestor.Close()

	ctx, trace := ingestor.StartTrace(context.Background(), "test-traces")

	_, span := trace.StartSpan(ctx, "test-span")

	require.NotNil(t, span)
	assert.Equal(t, "test-span", span.Name)
	assert.Equal(t, ObservationTypeSpan, span.Type)

	spanID := span.SpanContext().SpanID().String()
	assert.NotEmpty(t, spanID)
	assert.Len(t, spanID, 16)
	assert.Regexp(t, "^[0-9a-f]{16}$", spanID)
}

func TestTrace_MultipleSpans(t *testing.T) {
	ingestor, _ := newTestIngestor(t)
	defer ingestor.Close()

	ctx, trace := ingestor.StartTrace(context.Background(), "test-traces")

	_, span1 := trace.StartSpan(ctx, "span-1")
	_, span2 := trace.StartSpan(ctx, "span-2")

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

func TestTrace_SpansHaveUniqueIDs(t *testing.T) {
	ingestor, _ := newTestIngestor(t)
	defer ingestor.Close()

	ctx, trace := ingestor.StartTrace(context.Background(), "test-trace")

	spanIDs := make(map[string]bool)
	for _, name := range []string{"span-1", "span-2", "span-3", "span-4"} {
		_, span := trace.StartSpan(ctx, name)
		id := span.SpanContext().SpanID().String()
		assert.False(t, spanIDs[id], "Found duplicate span ID: %s", id)
		spanIDs[id] = true
	}
}

func TestTrace_StartObservation(t *testing.T) {
	ingestor, _ := newTestIngestor(t)
	defer ingestor.Close()

	ctx, trace := ingestor.StartTrace(context.Background(), "test-trace")

	_, observation := trace.StartObservation(ctx, "test-observation", ObservationTypeAgent)

	require.NotNil(t, observation)
	assert.Equal(t, "test-observation", observation.Name)
	assert.Equal(t, ObservationTypeAgent, observation.Type)
	assert.NotEmpty(t, observation.SpanContext().SpanID().String())

	_, observation2 := trace.StartObservation(ctx, "test-observation-2", ObservationTypeTool)

	require.NotNil(t, observation2)
	assert.Equal(t, "test-observation-2", observation2.Name)
	assert.Equal(t, ObservationTypeTool, observation2.Type)
	assert.NotEqual(t, observation.SpanContext().SpanID(), observation2.SpanContext().SpanID())
}

func TestTrace_StartGeneration(t *testing.T) {
	ingestor, _ := newTestIngestor(t)
	defer ingestor.Close()

	ctx, trace := ingestor.StartTrace(context.Background(), "test-trace")

	_, generation := trace.StartGeneration(ctx, "test-generation")

	require.NotNil(t, generation)
	assert.Equal(t, "test-generation", generation.Name)
	assert.Equal(t, ObservationTypeGeneration, generation.Type)
	assert.NotEmpty(t, generation.SpanContext().SpanID().String())

	_, generation2 := trace.StartGeneration(ctx, "test-generation-2")
	_, observation := trace.StartObservation(ctx, "test-observation", ObservationTypeGeneration)

	assert.Equal(t, generation2.Type, observation.Type)
	assert.Equal(t, ObservationTypeGeneration, generation2.Type)
}
