package traces

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestFromTraceID(t *testing.T) {
	gen := NewIDGenerator()
	traceID := gen.GenerateTraceID()
	gotTraceID, err := FromTraceID(traceID.String())
	require.NoError(t, err)
	require.Equal(t, traceID, gotTraceID)
}

func TestFromSpanID(t *testing.T) {
	gen := NewIDGenerator()
	spanID := gen.GenerateSpanID()
	gotSpanID, err := FromSpanID(spanID.String())
	require.NoError(t, err)
	require.Equal(t, spanID, gotSpanID)
}

func TestIDGenerator_GenerateTraceID(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	generator := NewIDGenerator()

	// Generate multiple trace IDs
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := generator.GenerateTraceID()
		idStr := id.String()

		// Check that ID is not all zeros
		require.NotEqual(t, "00000000000000000000000000000000", idStr)

		// Check uniqueness
		if ids[idStr] {
			logger.Error("generated duplicate trace ID",
				zap.String("id", idStr),
				zap.Int("iteration", i))
		}
		require.False(t, ids[idStr], "Generated duplicate trace ID: %s", idStr)
		ids[idStr] = true

		// Check proper length (32 hex characters)
		require.Len(t, idStr, 32)
	}

	logger.Info("successfully generated unique trace IDs", zap.Int("count", 1000))
}

func TestIDGenerator_GenerateSpanID(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	generator := NewIDGenerator()

	// Generate multiple span IDs
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := generator.GenerateSpanID()
		idStr := id.String()

		// Check that ID is not all zeros
		require.NotEqual(t, "0000000000000000", idStr)

		// Check uniqueness
		if ids[idStr] {
			logger.Error("generated duplicate span ID",
				zap.String("id", idStr),
				zap.Int("iteration", i))
		}
		require.False(t, ids[idStr], "Generated duplicate span ID: %s", idStr)
		ids[idStr] = true

		// Check proper length (16 hex characters)
		require.Len(t, idStr, 16)
	}

	logger.Info("successfully generated unique span IDs", zap.Int("count", 1000))
}

func TestIDGenerator_Concurrency(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	generator := NewIDGenerator()
	traceIDs := make(chan string, 100)
	spanIDs := make(chan string, 100)

	// Generate IDs concurrently
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				traceIDs <- generator.GenerateTraceID().String()
				spanIDs <- generator.GenerateSpanID().String()
			}
		}()
	}

	// Collect trace IDs and check uniqueness
	traceIDMap := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := <-traceIDs
		if traceIDMap[id] {
			logger.Error("found duplicate trace ID in concurrent test", zap.String("id", id))
		}
		require.False(t, traceIDMap[id], "Found duplicate trace ID: %s", id)
		traceIDMap[id] = true
	}

	// Collect span IDs and check uniqueness
	spanIDMap := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := <-spanIDs
		if spanIDMap[id] {
			logger.Error("found duplicate span ID in concurrent test", zap.String("id", id))
		}
		require.False(t, spanIDMap[id], "Found duplicate span ID: %s", id)
		spanIDMap[id] = true
	}

	logger.Info("successfully completed concurrent ID generation test")
}

func TestIngestor_StartTrace(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	ingestor := NewIngestor(client)

	trace := ingestor.StartTrace("test-trace")

	require.NotNil(t, trace)
	require.Equal(t, "test-trace", trace.Name)
	require.NotEmpty(t, trace.ID)
	require.Len(t, trace.ID, 32) // Trace ID should be 32 hex characters
	require.NotEmpty(t, trace.Timestamp)
	require.Empty(t, trace.observations)
	require.Equal(t, ingestor, trace.ingestor)

	logger.Info("successfully started trace",
		zap.String("trace_id", trace.ID),
		zap.String("trace_name", trace.Name))
}

func TestIngestor_StartTrace_UniqueIDs(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	ingestor := NewIngestor(client)

	// Generate multiple traces and ensure unique IDs
	traceIDs := make(map[string]bool)
	for i := 0; i < 100; i++ {
		trace := ingestor.StartTrace("test-trace")
		if traceIDs[trace.ID] {
			logger.Error("found duplicate trace ID in ingestor test",
				zap.String("id", trace.ID),
				zap.Int("iteration", i))
		}
		require.False(t, traceIDs[trace.ID], "Found duplicate trace ID: %s", trace.ID)
		traceIDs[trace.ID] = true
	}

	logger.Info("successfully generated unique trace IDs via ingestor", zap.Int("count", 100))
}

func TestIngestor_Send(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		traces         []*Trace
		wantError      bool
	}{
		{
			name: "successful send",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "POST", r.Method)
				require.Equal(t, "/ingestion", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success": true}`))
			},
			traces: []*Trace{
				{
					TraceEntry: TraceEntry{
						ID:        "trace-1",
						Name:      "test-trace",
						Timestamp: time.Now(),
					},
				},
			},
			wantError: false,
		},
		{
			name: "server error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal server error"}`))
			},
			traces: []*Trace{
				{
					TraceEntry: TraceEntry{
						ID:        "trace-1",
						Name:      "test-trace",
						Timestamp: time.Now(),
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			client := resty.New().SetBaseURL(server.URL)
			ingestor := NewIngestor(client)

			err := ingestor.Send(context.Background(), tt.traces)
			if tt.wantError {
				require.Error(t, err)
				logger.Info("expected error occurred", zap.Error(err))
			} else {
				require.NoError(t, err)
				logger.Info("successfully sent traces")
			}
		})
	}
}
