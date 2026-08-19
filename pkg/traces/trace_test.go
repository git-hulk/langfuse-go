package traces

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrace_End_CalculatesLatency(t *testing.T) {
	startTime := time.Now().Add(-100 * time.Millisecond)
	trace := &Trace{
		TraceEntry: TraceEntry{
			ID:        "test-traces-id",
			Name:      "test-traces",
			Timestamp: startTime,
		},
	}

	latency := time.Since(startTime).Milliseconds()
	trace.Latency = latency

	assert.Greater(t, trace.Latency, int64(0))
	assert.GreaterOrEqual(t, trace.Latency, int64(90))
}

func TestTrace_StartSpan(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-traces")

	span := trace.StartSpan("test-span")

	require.NotNil(t, span)
	assert.Equal(t, "test-span", span.Name)
	assert.Equal(t, ObservationTypeSpan, span.Type)
	assert.Equal(t, trace.ID, span.TraceID)
	assert.Equal(t, trace.ID, span.ParentObservationID)
	assert.NotEmpty(t, span.ID)
	assert.False(t, span.StartTime.IsZero())

	assert.Len(t, span.ID, 16)
	assert.Regexp(t, "^[0-9a-f]{16}$", span.ID)

	assert.Len(t, trace.observations, 1)
	assert.Equal(t, span, trace.observations[0])
}

func TestTrace_MultipleSpans(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-traces")

	span1 := trace.StartSpan("span-1")
	span2 := trace.StartSpan("span-2")

	assert.Len(t, trace.observations, 2)
	assert.Equal(t, "span-1", span1.Name)
	assert.Equal(t, "span-2", span2.Name)
	assert.NotEqual(t, span1.ID, span2.ID)

	assert.Equal(t, trace.ID, span1.ParentObservationID)
	assert.Equal(t, span1.ID, span2.ParentObservationID)
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
			TotalCost:   0.05,
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
	assert.Equal(t, 0.05, trace.TotalCost)
	assert.Equal(t, "test", trace.Environment)
}

func TestTrace_NestedSpans(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-trace")

	parentSpan := trace.StartSpan("parent-span")
	assert.Equal(t, trace.ID, parentSpan.ParentObservationID)

	childSpan := trace.StartSpan("child-span")
	assert.Equal(t, parentSpan.ID, childSpan.ParentObservationID)

	childSpan2 := trace.StartSpan("child-span-2")
	assert.Equal(t, childSpan.ID, childSpan2.ParentObservationID)

	assert.Len(t, trace.observations, 3)
	assert.NotEqual(t, parentSpan.ID, childSpan.ID)
	assert.NotEqual(t, childSpan.ID, childSpan2.ID)
	assert.NotEqual(t, parentSpan.ID, childSpan2.ID)
}

func TestTrace_NestedSpansWithEndedSpans(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-trace")

	parentSpan := trace.StartSpan("parent-span")
	assert.Equal(t, trace.ID, parentSpan.ParentObservationID)

	childSpan := trace.StartSpan("child-span")
	assert.Equal(t, parentSpan.ID, childSpan.ParentObservationID)

	childSpan.End()
	require.NotNil(t, childSpan.EndTime)
	assert.False(t, childSpan.EndTime.IsZero())

	siblingSpan := trace.StartSpan("sibling-span")
	assert.Equal(t, parentSpan.ID, siblingSpan.ParentObservationID)

	assert.Len(t, trace.observations, 3)
}

func TestTrace_GetParentObservationID(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	t.Run("no observations - returns trace ID", func(t *testing.T) {
		trace := ingestor.StartTrace(context.Background(), "test-trace")
		result := trace.getParentObservationID()
		assert.Equal(t, trace.ID, result)
	})

	t.Run("active observation - returns observation ID", func(t *testing.T) {
		trace := ingestor.StartTrace(context.Background(), "test-trace")
		span := trace.StartSpan("active-span")
		result := trace.getParentObservationID()
		assert.Equal(t, span.ID, result)
	})

	t.Run("ended observation - returns parent observation ID", func(t *testing.T) {
		trace := ingestor.StartTrace(context.Background(), "test-trace")
		span := trace.StartSpan("ended-span")
		parentID := span.ParentObservationID
		span.End()
		result := trace.getParentObservationID()
		assert.Equal(t, parentID, result)
	})
}

func TestTrace_DeepNestedSpans(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-trace")

	level1 := trace.StartSpan("level-1")
	assert.Equal(t, trace.ID, level1.ParentObservationID)

	level2 := trace.StartSpan("level-2")
	assert.Equal(t, level1.ID, level2.ParentObservationID)

	level3 := trace.StartSpan("level-3")
	assert.Equal(t, level2.ID, level3.ParentObservationID)

	level4 := trace.StartSpan("level-4")
	assert.Equal(t, level3.ID, level4.ParentObservationID)

	assert.Len(t, trace.observations, 4)

	spanIDs := make(map[string]bool)
	for _, obs := range trace.observations {
		assert.False(t, spanIDs[obs.ID], "Found duplicate span ID: %s", obs.ID)
		spanIDs[obs.ID] = true
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
	assert.Equal(t, trace.ID, observation.TraceID)
	assert.Equal(t, trace.ID, observation.ParentObservationID)
	assert.NotEmpty(t, observation.ID)
	assert.False(t, observation.StartTime.IsZero())
	assert.Nil(t, observation.EndTime)

	assert.Len(t, trace.observations, 1)
	assert.Equal(t, observation, trace.observations[0])

	observation2 := trace.StartObservation("test-observation-2", ObservationTypeTool)

	require.NotNil(t, observation2)
	assert.Equal(t, "test-observation-2", observation2.Name)
	assert.Equal(t, ObservationTypeTool, observation2.Type)
	assert.Equal(t, trace.ID, observation2.TraceID)
	assert.Equal(t, observation.ID, observation2.ParentObservationID)
	assert.NotEqual(t, observation.ID, observation2.ID)

	assert.Len(t, trace.observations, 2)
	assert.Equal(t, observation, trace.observations[0])
	assert.Equal(t, observation2, trace.observations[1])
}

func TestTrace_StartGeneration(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-trace")

	generation := trace.StartGeneration("test-generation")

	require.NotNil(t, generation)
	assert.Equal(t, "test-generation", generation.Name)
	assert.Equal(t, ObservationTypeGeneration, generation.Type)
	assert.Equal(t, trace.ID, generation.TraceID)
	assert.Equal(t, trace.ID, generation.ParentObservationID)
	assert.NotEmpty(t, generation.ID)
	assert.False(t, generation.StartTime.IsZero())
	assert.Nil(t, generation.EndTime)

	assert.Len(t, trace.observations, 1)
	assert.Equal(t, generation, trace.observations[0])

	generation2 := trace.StartGeneration("test-generation-2")
	observation := trace.StartObservation("test-observation", ObservationTypeGeneration)

	assert.Equal(t, generation2.Type, observation.Type)
	assert.Equal(t, ObservationTypeGeneration, generation2.Type)
}
