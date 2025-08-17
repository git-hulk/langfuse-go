package traces

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrace_End_CalculatesLatency(t *testing.T) {
	startTime := time.Now().Add(-100 * time.Millisecond)
	trace := &Trace{
		ID:        "test-traces-id",
		Name:      "test-traces",
		Timestamp: startTime,
	}

	latency := time.Since(startTime).Milliseconds()
	trace.Latency = latency

	assert.Greater(t, trace.Latency, int64(0))
	assert.GreaterOrEqual(t, trace.Latency, int64(90))
}

func TestTrace_StartSpan(t *testing.T) {
	trace := &Trace{
		ID:           "test-traces-id",
		Name:         "test-traces",
		observations: []*Observation{},
	}

	span := trace.StartSpan("test-span")

	require.NotNil(t, span)
	assert.Equal(t, "test-span", span.Name)
	assert.Equal(t, ObservationTypeSpan, span.Type)
	assert.Equal(t, "test-traces-id", span.TraceID)
	assert.Equal(t, "test-traces-id", span.ParentObservationID)
	assert.NotEmpty(t, span.ID)
	assert.False(t, span.StartTime.IsZero())

	_, err := uuid.FromString(span.ID)
	assert.NoError(t, err, "Span ID should be a valid UUID")

	assert.Len(t, trace.observations, 1)
	assert.Equal(t, span, trace.observations[0])
}

func TestTrace_MultipleSpans(t *testing.T) {
	trace := &Trace{
		ID:           "test-traces-id",
		Name:         "test-traces",
		observations: []*Observation{},
	}

	span1 := trace.StartSpan("span-1")
	span2 := trace.StartSpan("span-2")

	assert.Len(t, trace.observations, 2)
	assert.Equal(t, "span-1", span1.Name)
	assert.Equal(t, "span-2", span2.Name)
	assert.NotEqual(t, span1.ID, span2.ID)
}

func TestTrace_Fields(t *testing.T) {
	trace := &Trace{
		ID:          "test-id",
		Name:        "test-name",
		SessionID:   "session-123",
		Release:     "v1.0.0",
		Version:     "1.0",
		UserID:      "user-456",
		Metadata:    map[string]interface{}{"key": "value"},
		Tags:        []string{"tag1", "tag2"},
		TotalCost:   0.05,
		Environment: "test",
	}

	assert.Equal(t, "test-id", trace.ID)
	assert.Equal(t, "test-name", trace.Name)
	assert.Equal(t, "session-123", trace.SessionID)
	assert.Equal(t, "v1.0.0", trace.Release)
	assert.Equal(t, "1.0", trace.Version)
	assert.Equal(t, "user-456", trace.UserID)
	assert.Equal(t, map[string]interface{}{"key": "value"}, trace.Metadata)
	assert.Equal(t, []string{"tag1", "tag2"}, trace.Tags)
	assert.Equal(t, 0.05, trace.TotalCost)
	assert.Equal(t, "test", trace.Environment)
}
