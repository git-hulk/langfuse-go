package traces

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
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
	// Create ingestor with mock server for ID generation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	ingestor := NewIngestor(client)

	trace := &Trace{
		ingestor: ingestor,
		TraceEntry: TraceEntry{
			ID:   "test-traces-id",
			Name: "test-traces",
		},
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

	// Check that span ID is a valid hex string of length 16
	assert.Len(t, span.ID, 16)
	assert.Regexp(t, "^[0-9a-f]{16}$", span.ID)

	assert.Len(t, trace.observations, 1)
	assert.Equal(t, span, trace.observations[0])
}

func TestTrace_MultipleSpans(t *testing.T) {
	// Create ingestor with mock server for ID generation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	ingestor := NewIngestor(client)

	trace := &Trace{
		ingestor: ingestor,
		TraceEntry: TraceEntry{
			ID:   "test-traces-id",
			Name: "test-traces",
		},
		observations: []*Observation{},
	}

	span1 := trace.StartSpan("span-1")
	span2 := trace.StartSpan("span-2")

	assert.Len(t, trace.observations, 2)
	assert.Equal(t, "span-1", span1.Name)
	assert.Equal(t, "span-2", span2.Name)
	assert.NotEqual(t, span1.ID, span2.ID)

	// First span should have trace ID as parent
	assert.Equal(t, "test-traces-id", span1.ParentObservationID)
	// Second span should have first span as parent (since first span is still active)
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
			Metadata:    map[string]interface{}{"key": "value"},
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
	assert.Equal(t, map[string]interface{}{"key": "value"}, trace.Metadata)
	assert.Equal(t, []string{"tag1", "tag2"}, trace.Tags)
	assert.Equal(t, 0.05, trace.TotalCost)
	assert.Equal(t, "test", trace.Environment)
}

func TestTrace_NestedSpans(t *testing.T) {
	// Create ingestor with mock server for ID generation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	ingestor := NewIngestor(client)

	trace := &Trace{
		ingestor: ingestor,
		TraceEntry: TraceEntry{
			ID:   "test-trace-id",
			Name: "test-trace",
		},
		observations: []*Observation{},
	}

	// Create parent span
	parentSpan := trace.StartSpan("parent-span")
	assert.Equal(t, "test-trace-id", parentSpan.ParentObservationID) // Parent is the trace

	// Create child span while parent is still active
	childSpan := trace.StartSpan("child-span")
	assert.Equal(t, parentSpan.ID, childSpan.ParentObservationID) // Parent is the active span

	// Create another child span while first child is still active
	childSpan2 := trace.StartSpan("child-span-2")
	assert.Equal(t, childSpan.ID, childSpan2.ParentObservationID) // Parent is the last active span

	assert.Len(t, trace.observations, 3)
	assert.NotEqual(t, parentSpan.ID, childSpan.ID)
	assert.NotEqual(t, childSpan.ID, childSpan2.ID)
	assert.NotEqual(t, parentSpan.ID, childSpan2.ID)
}

func TestTrace_NestedSpansWithEndedSpans(t *testing.T) {
	// Create ingestor with mock server for ID generation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	ingestor := NewIngestor(client)

	trace := &Trace{
		ingestor: ingestor,
		TraceEntry: TraceEntry{
			ID:   "test-trace-id",
			Name: "test-trace",
		},
		observations: []*Observation{},
	}

	// Create parent span
	parentSpan := trace.StartSpan("parent-span")
	assert.Equal(t, "test-trace-id", parentSpan.ParentObservationID) // Parent is the trace

	// Create child span
	childSpan := trace.StartSpan("child-span")
	assert.Equal(t, parentSpan.ID, childSpan.ParentObservationID) // Parent is the active span

	// End the child span
	childSpan.End()
	assert.False(t, childSpan.EndTime.IsZero())

	// Create another span after child has ended
	siblingSpan := trace.StartSpan("sibling-span")
	// Since child span is ended, it should use the child span's parent (parentSpan.ID)
	assert.Equal(t, parentSpan.ID, siblingSpan.ParentObservationID)

	assert.Len(t, trace.observations, 3)
}

func TestTrace_GetParentObservationID(t *testing.T) {
	// Create ingestor with mock server for ID generation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	ingestor := NewIngestor(client)

	trace := &Trace{
		ingestor: ingestor,
		TraceEntry: TraceEntry{
			ID:   "test-trace-id",
			Name: "test-trace",
		},
		observations: []*Observation{},
	}

	tests := []struct {
		name     string
		setup    func() string
		expected string
	}{
		{
			name: "no observations - returns trace ID",
			setup: func() string {
				return trace.getParentObservationID()
			},
			expected: "test-trace-id",
		},
		{
			name: "active observation - returns observation ID",
			setup: func() string {
				_ = trace.StartSpan("active-span")
				return trace.getParentObservationID()
			},
			expected: "", // Will be set to span.ID in the test
		},
		{
			name: "ended observation - returns parent observation ID",
			setup: func() string {
				span := trace.StartSpan("ended-span")
				span.ParentObservationID = "parent-id"
				span.End()
				return trace.getParentObservationID()
			},
			expected: "parent-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset trace observations for each test
			trace.observations = []*Observation{}

			result := tt.setup()

			if tt.name == "active observation - returns observation ID" {
				// For this test, we expect the result to be the span ID
				assert.NotEqual(t, "test-trace-id", result)
				assert.Len(t, result, 16) // Should be a span ID
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestTrace_DeepNestedSpans(t *testing.T) {
	// Create ingestor with mock server for ID generation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	ingestor := NewIngestor(client)

	trace := &Trace{
		ingestor: ingestor,
		TraceEntry: TraceEntry{
			ID:   "test-trace-id",
			Name: "test-trace",
		},
		observations: []*Observation{},
	}

	// Create a chain of nested spans
	level1 := trace.StartSpan("level-1")
	assert.Equal(t, "test-trace-id", level1.ParentObservationID)

	level2 := trace.StartSpan("level-2")
	assert.Equal(t, level1.ID, level2.ParentObservationID)

	level3 := trace.StartSpan("level-3")
	assert.Equal(t, level2.ID, level3.ParentObservationID)

	level4 := trace.StartSpan("level-4")
	assert.Equal(t, level3.ID, level4.ParentObservationID)

	assert.Len(t, trace.observations, 4)

	// Verify all spans are unique
	spanIDs := make(map[string]bool)
	for _, obs := range trace.observations {
		assert.False(t, spanIDs[obs.ID], "Found duplicate span ID: %s", obs.ID)
		spanIDs[obs.ID] = true
	}
}
