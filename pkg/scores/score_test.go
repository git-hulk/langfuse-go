package scores

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"

	"github.com/git-hulk/langfuse-go/pkg/common"
)

func TestCreateScoreRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		request CreateScoreRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with trace ID",
			request: CreateScoreRequest{
				Name:    "accuracy",
				Value:   0.95,
				TraceID: "trace-123",
			},
			wantErr: false,
		},
		{
			name: "valid request with session ID",
			request: CreateScoreRequest{
				Name:      "quality",
				Value:     "excellent",
				SessionID: "session-456",
			},
			wantErr: false,
		},
		{
			name: "valid request with observation ID",
			request: CreateScoreRequest{
				Name:          "relevance",
				Value:         1.0,
				TraceID:       "trace-123",
				ObservationID: "obs-789",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			request: CreateScoreRequest{
				Value:   0.8,
				TraceID: "trace-123",
			},
			wantErr: true,
			errMsg:  "'name' is required",
		},
		{
			name: "missing value",
			request: CreateScoreRequest{
				Name:    "accuracy",
				TraceID: "trace-123",
			},
			wantErr: true,
			errMsg:  "'value' is required",
		},
		{
			name: "missing all IDs",
			request: CreateScoreRequest{
				Name:  "accuracy",
				Value: 0.8,
			},
			wantErr: true,
			errMsg:  "at least one of 'traceId', 'sessionId', or 'datasetRunID' is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

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
			name:   "page and limit only",
			params: ListParams{Page: 2, Limit: 25},
			want:   "page=2&limit=25",
		},
		{
			name: "with user ID and name",
			params: ListParams{
				UserID: "user-123",
				Name:   "accuracy",
			},
			want: "userId=user-123&name=accuracy",
		},
		{
			name: "with timestamps",
			params: ListParams{
				FromTimestamp: mustParseTime("2023-01-01T10:00:00Z"),
				ToTimestamp:   mustParseTime("2023-01-02T10:00:00Z"),
			},
			want: "fromTimestamp=2023-01-01T10%3A00%3A00Z&toTimestamp=2023-01-02T10%3A00%3A00Z",
		},
		{
			name: "with environment filter",
			params: ListParams{
				Environment: []string{"production", "staging"},
			},
			want: "environment=production&environment=staging",
		},
		{
			name: "with source and data type",
			params: ListParams{
				Source:   ScoreSourceAPI,
				DataType: ScoreDataTypeNumeric,
			},
			want: "source=API&dataType=NUMERIC",
		},
		{
			name: "with operator and value",
			params: ListParams{
				Operator: ">=",
				Value:    0.8,
			},
			want: "operator=%3E%3D&value=0.8",
		},
		{
			name: "with score IDs",
			params: ListParams{
				ScoreIDs: []string{"score-1", "score-2", "score-3"},
			},
			want: "scoreIds=score-1%2Cscore-2%2Cscore-3",
		},
		{
			name: "with config and queue IDs",
			params: ListParams{
				ConfigID: "config-123",
				QueueID:  "queue-456",
			},
			want: "configId=config-123&queueId=queue-456",
		},
		{
			name: "with trace tags",
			params: ListParams{
				TraceTags: []string{"experiment", "production"},
			},
			want: "traceTags=experiment&traceTags=production",
		},
		{
			name: "all parameters",
			params: ListParams{
				Page:          1,
				Limit:         10,
				UserID:        "user-123",
				Name:          "quality",
				FromTimestamp: mustParseTime("2023-01-01T10:00:00Z"),
				ToTimestamp:   mustParseTime("2023-01-02T10:00:00Z"),
				Environment:   []string{"production"},
				Source:        ScoreSourceEval,
				Operator:      ">=",
				Value:         0.9,
				ScoreIDs:      []string{"score-1"},
				ConfigID:      "config-123",
				QueueID:       "queue-456",
				DataType:      ScoreDataTypeBoolean,
				TraceTags:     []string{"test"},
			},
			want: "page=1&limit=10&userId=user-123&name=quality&fromTimestamp=2023-01-01T10%3A00%3A00Z&toTimestamp=2023-01-02T10%3A00%3A00Z&environment=production&source=EVAL&operator=%3E%3D&value=0.9&scoreIds=score-1&configId=config-123&queueId=queue-456&dataType=BOOLEAN&traceTags=test",
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

	t.Run("successful list scores", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v2/scores", r.URL.Path)
			require.Equal(t, "GET", r.Method)

			query := r.URL.Query()
			require.Equal(t, "2", query.Get("page"))
			require.Equal(t, "5", query.Get("limit"))

			scores := ListScores{
				Metadata: common.ListMetadata{
					Page:       2,
					Limit:      5,
					TotalItems: 15,
					TotalPages: 3,
				},
				Data: []Score{
					{
						ID:        "score-1",
						Name:      "accuracy",
						Source:    ScoreSourceAPI,
						TraceID:   "trace-123",
						CreatedAt: mustParseTime("2023-01-01T10:00:00Z"),
						UpdatedAt: mustParseTime("2023-01-01T10:00:00Z"),
						DataType:  ScoreDataTypeNumeric,
						Value:     0.95,
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(scores)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		scoreClient := NewClient(client)

		params := ListParams{Page: 2, Limit: 5}
		result, err := scoreClient.List(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 1, len(result.Data))
		require.Equal(t, 15, result.Metadata.TotalItems)
		require.Equal(t, 3, result.Metadata.TotalPages)
	})

	t.Run("list with filters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v2/scores", r.URL.Path)
			require.Equal(t, "GET", r.Method)

			query := r.URL.Query()
			require.Equal(t, "user-123", query.Get("userId"))
			require.Equal(t, "accuracy", query.Get("name"))
			require.Equal(t, "API", query.Get("source"))

			scores := ListScores{
				Metadata: common.ListMetadata{
					Page:       1,
					Limit:      50,
					TotalItems: 2,
					TotalPages: 1,
				},
				Data: []Score{},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(scores)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		scoreClient := NewClient(client)

		params := ListParams{
			UserID: "user-123",
			Name:   "accuracy",
			Source: ScoreSourceAPI,
		}
		result, err := scoreClient.List(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 0, len(result.Data))
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		scoreClient := NewClient(client)

		result, err := scoreClient.List(ctx, ListParams{})
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "500")
	})
}

func TestClient_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get score", func(t *testing.T) {
		scoreID := "score-123"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v2/scores/"+scoreID, r.URL.Path)
			require.Equal(t, "GET", r.Method)

			score := Score{
				ID:        scoreID,
				Name:      "quality",
				Source:    ScoreSourceAnnotation,
				TraceID:   "trace-456",
				CreatedAt: mustParseTime("2023-01-01T10:00:00Z"),
				UpdatedAt: mustParseTime("2023-01-01T10:00:00Z"),
				DataType:  ScoreDataTypeBoolean,
				Value:     1.0,
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(score)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		scoreClient := NewClient(client)

		result, err := scoreClient.Get(ctx, scoreID)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, scoreID, result.ID)
		require.Equal(t, "quality", result.Name)
		require.Equal(t, ScoreDataTypeBoolean, result.DataType)
		require.Equal(t, 1.0, result.Value)
	})

	t.Run("get with empty score ID", func(t *testing.T) {
		client := resty.New()
		scoreClient := NewClient(client)

		result, err := scoreClient.Get(ctx, "")
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "'scoreID' is required", err.Error())
	})

	t.Run("score not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Score not found"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		scoreClient := NewClient(client)

		result, err := scoreClient.Get(ctx, "nonexistent-score")
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "404")
	})
}

func TestClient_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful create score", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/scores", r.URL.Path)
			require.Equal(t, "POST", r.Method)

			var createReq CreateScoreRequest
			err := json.NewDecoder(r.Body).Decode(&createReq)
			require.NoError(t, err)
			require.Equal(t, "accuracy", createReq.Name)
			require.Equal(t, 0.95, createReq.Value)
			require.Equal(t, "trace-123", createReq.TraceID)

			response := CreateScoreResponse{
				ID: "score-created-456",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		scoreClient := NewClient(client)

		createReq := &CreateScoreRequest{
			Name:    "accuracy",
			Value:   0.95,
			TraceID: "trace-123",
			Comment: "Excellent performance",
		}

		result, err := scoreClient.Create(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "score-created-456", result.ID)
	})

	t.Run("create with validation error", func(t *testing.T) {
		client := resty.New()
		scoreClient := NewClient(client)

		createReq := &CreateScoreRequest{} // Missing required fields
		result, err := scoreClient.Create(ctx, createReq)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "'name' is required")
	})

	t.Run("create with server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid request"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		scoreClient := NewClient(client)

		createReq := &CreateScoreRequest{
			Name:    "accuracy",
			Value:   0.95,
			TraceID: "trace-123",
		}
		result, err := scoreClient.Create(ctx, createReq)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "400")
	})
}

func TestClient_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete score", func(t *testing.T) {
		scoreID := "score-123"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/scores/"+scoreID, r.URL.Path)
			require.Equal(t, "DELETE", r.Method)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		scoreClient := NewClient(client)

		err := scoreClient.Delete(ctx, scoreID)
		require.NoError(t, err)
	})

	t.Run("delete with empty score ID", func(t *testing.T) {
		client := resty.New()
		scoreClient := NewClient(client)

		err := scoreClient.Delete(ctx, "")
		require.Error(t, err)
		require.Equal(t, "'scoreID' is required", err.Error())
	})

	t.Run("delete with server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		scoreClient := NewClient(client)

		err := scoreClient.Delete(ctx, "score-123")
		require.Error(t, err)
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
