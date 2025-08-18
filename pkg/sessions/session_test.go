package sessions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/git-hulk/langfuse-go/pkg/traces"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"

	"github.com/git-hulk/langfuse-go/pkg/common"
)

func TestListParams_ToQueryString(t *testing.T) {
	tests := []struct {
		name   string
		params ListParams
		want   string
	}{
		{
			name:   "empty params",
			params: ListParams{},
			want:   "",
		},
		{
			name:   "page only",
			params: ListParams{Page: 2},
			want:   "page=2",
		},
		{
			name:   "limit only",
			params: ListParams{Limit: 25},
			want:   "limit=25",
		},
		{
			name:   "both page and limit",
			params: ListParams{Page: 3, Limit: 15},
			want:   "page=3&limit=15",
		},
		{
			name: "with timestamps",
			params: ListParams{
				Page:          1,
				Limit:         10,
				FromTimestamp: mustParseTime("2023-01-01T10:00:00Z"),
				ToTimestamp:   mustParseTime("2023-01-02T10:00:00Z"),
			},
			want: "page=1&limit=10&fromTimestamp=2023-01-01T10%3A00%3A00Z&toTimestamp=2023-01-02T10%3A00%3A00Z",
		},
		{
			name: "with environments",
			params: ListParams{
				Environment: []string{"production", "staging"},
			},
			want: "environment=production&environment=staging",
		},
		{
			name: "all parameters",
			params: ListParams{
				Page:          1,
				Limit:         5,
				FromTimestamp: mustParseTime("2023-01-01T10:00:00Z"),
				ToTimestamp:   mustParseTime("2023-01-02T10:00:00Z"),
				Environment:   []string{"production"},
			},
			want: "page=1&limit=5&fromTimestamp=2023-01-01T10%3A00%3A00Z&toTimestamp=2023-01-02T10%3A00%3A00Z&environment=production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.params.ToQueryString()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestClient_List(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list sessions", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/sessions", r.URL.Path)
			require.Equal(t, "GET", r.Method)

			query := r.URL.Query()
			require.Equal(t, "2", query.Get("page"))
			require.Equal(t, "5", query.Get("limit"))

			sessions := ListSessions{
				Metadata: common.ListMetadata{
					Page:       2,
					Limit:      5,
					TotalItems: 15,
					TotalPages: 3,
				},
				Data: []Session{
					{
						ID:          "session-1",
						CreatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
						ProjectID:   "project-123",
						Environment: "production",
					},
					{
						ID:        "session-2",
						CreatedAt: mustParseTime("2023-01-02T10:00:00Z"),
						ProjectID: "project-123",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(sessions)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		sessionClient := NewClient(client)

		params := ListParams{Page: 2, Limit: 5}
		result, err := sessionClient.List(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 2, len(result.Data))
		require.Equal(t, "session-1", result.Data[0].ID)
		require.Equal(t, "production", result.Data[0].Environment)
		require.Equal(t, 15, result.Metadata.TotalItems)
		require.Equal(t, 3, result.Metadata.TotalPages)
	})

	t.Run("list with timestamp filters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/sessions", r.URL.Path)
			require.Equal(t, "GET", r.Method)

			query := r.URL.Query()
			require.Equal(t, "2023-01-01T10:00:00Z", query.Get("fromTimestamp"))
			require.Equal(t, "2023-01-02T10:00:00Z", query.Get("toTimestamp"))

			sessions := ListSessions{
				Metadata: common.ListMetadata{
					Page:       1,
					Limit:      50,
					TotalItems: 1,
					TotalPages: 1,
				},
				Data: []Session{
					{
						ID:        "session-1",
						CreatedAt: mustParseTime("2023-01-01T15:00:00Z"),
						ProjectID: "project-123",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(sessions)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		sessionClient := NewClient(client)

		params := ListParams{
			FromTimestamp: mustParseTime("2023-01-01T10:00:00Z"),
			ToTimestamp:   mustParseTime("2023-01-02T10:00:00Z"),
		}
		result, err := sessionClient.List(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 1, len(result.Data))
		require.Equal(t, "session-1", result.Data[0].ID)
	})

	t.Run("list with environment filter", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/sessions", r.URL.Path)
			require.Equal(t, "GET", r.Method)

			query := r.URL.Query()
			envs := query["environment"]
			require.Contains(t, envs, "production")
			require.Contains(t, envs, "staging")

			sessions := ListSessions{
				Metadata: common.ListMetadata{
					Page:       1,
					Limit:      50,
					TotalItems: 2,
					TotalPages: 1,
				},
				Data: []Session{
					{
						ID:          "session-1",
						CreatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
						ProjectID:   "project-123",
						Environment: "production",
					},
					{
						ID:          "session-2",
						CreatedAt:   mustParseTime("2023-01-01T11:00:00Z"),
						ProjectID:   "project-123",
						Environment: "staging",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(sessions)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		sessionClient := NewClient(client)

		params := ListParams{
			Environment: []string{"production", "staging"},
		}
		result, err := sessionClient.List(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 2, len(result.Data))
	})

	t.Run("empty session list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessions := ListSessions{
				Metadata: common.ListMetadata{
					Page:       1,
					Limit:      50,
					TotalItems: 0,
					TotalPages: 0,
				},
				Data: []Session{},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(sessions)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		sessionClient := NewClient(client)

		result, err := sessionClient.List(ctx, ListParams{})
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 0, len(result.Data))
		require.Equal(t, 0, result.Metadata.TotalItems)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		sessionClient := NewClient(client)

		result, err := sessionClient.List(ctx, ListParams{})
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "500")
	})
}

func TestClient_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get session", func(t *testing.T) {
		sessionID := "session-123"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/sessions/"+sessionID, r.URL.Path)
			require.Equal(t, "GET", r.Method)

			sessionWithTraces := SessionWithTraces{
				Session: Session{
					ID:          sessionID,
					CreatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
					ProjectID:   "project-456",
					Environment: "production",
				},
				Traces: []traces.TraceEntry{
					{
						ID:        "trace-1",
						Name:      "test-trace",
						Timestamp: mustParseTime("2023-01-01T10:05:00Z"),
						SessionID: sessionID,
						Input:     map[string]interface{}{"query": "test"},
						Output:    map[string]interface{}{"response": "result"},
					},
					{
						ID:        "trace-2",
						Name:      "another-trace",
						Timestamp: mustParseTime("2023-01-01T10:10:00Z"),
						SessionID: sessionID,
						Tags:      []string{"test", "api"},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(sessionWithTraces)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		sessionClient := NewClient(client)

		result, err := sessionClient.Get(ctx, sessionID)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, sessionID, result.ID)
		require.Equal(t, "project-456", result.ProjectID)
		require.Equal(t, "production", result.Environment)
		require.Equal(t, 2, len(result.Traces))
		require.Equal(t, "trace-1", result.Traces[0].ID)
		require.Equal(t, "test-trace", result.Traces[0].Name)
		require.Equal(t, []string{"test", "api"}, result.Traces[1].Tags)
	})

	t.Run("get with empty session ID", func(t *testing.T) {
		client := resty.New()
		sessionClient := NewClient(client)

		result, err := sessionClient.Get(ctx, "")
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "'sessionID' is required", err.Error())
	})

	t.Run("session not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Session not found"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		sessionClient := NewClient(client)

		result, err := sessionClient.Get(ctx, "nonexistent-session")
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "404")
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		sessionClient := NewClient(client)

		result, err := sessionClient.Get(ctx, "session-123")
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "500")
	})
}

// Helper functions for tests

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
