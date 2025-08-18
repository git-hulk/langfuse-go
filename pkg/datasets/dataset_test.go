package datasets

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

// Tests for V2 Datasets API

func TestCreateDatasetRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		request CreateDatasetRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with all fields",
			request: CreateDatasetRequest{
				Name:        "test-dataset",
				Description: "A test dataset",
				Metadata:    map[string]interface{}{"version": "1.0"},
			},
			wantErr: false,
		},
		{
			name: "valid request with name only",
			request: CreateDatasetRequest{
				Name: "minimal-dataset",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			request: CreateDatasetRequest{
				Description: "Dataset without name",
			},
			wantErr: true,
			errMsg:  "'name' is required",
		},
		{
			name: "empty name",
			request: CreateDatasetRequest{
				Name: "",
			},
			wantErr: true,
			errMsg:  "'name' is required",
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

func TestListParams_ToQueryString_V2(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.params.ToQueryString()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestClient_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get dataset", func(t *testing.T) {
		datasetName := "test-dataset"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v2/datasets/"+datasetName, r.URL.Path)
			require.Equal(t, "GET", r.Method)

			dataset := Dataset{
				ID:          "dataset-123",
				Name:        datasetName,
				Description: "Test dataset description",
				Metadata:    map[string]interface{}{"version": "1.0"},
				ProjectID:   "project-456",
				CreatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
				UpdatedAt:   mustParseTime("2023-01-02T11:00:00Z"),
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(dataset)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		result, err := datasetClient.Get(ctx, datasetName)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "dataset-123", result.ID)
		require.Equal(t, datasetName, result.Name)
		require.Equal(t, "Test dataset description", result.Description)
		require.Equal(t, "project-456", result.ProjectID)
	})

	t.Run("get with empty dataset name", func(t *testing.T) {
		client := resty.New()
		datasetClient := NewClient(client)

		result, err := datasetClient.Get(ctx, "")
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "'datasetName' is required", err.Error())
	})

	t.Run("dataset not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Dataset not found"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		result, err := datasetClient.Get(ctx, "nonexistent-dataset")
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "404")
	})
}

func TestClient_List_V2(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list datasets", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v2/datasets", r.URL.Path)
			require.Equal(t, "GET", r.Method)

			query := r.URL.Query()
			require.Equal(t, "2", query.Get("page"))
			require.Equal(t, "5", query.Get("limit"))

			datasets := ListDatasets{
				Metadata: common.ListMetadata{
					Page:       2,
					Limit:      5,
					TotalItems: 15,
					TotalPages: 3,
				},
				Data: []Dataset{
					{
						ID:          "dataset-1",
						Name:        "dataset-one",
						Description: "First dataset",
						ProjectID:   "project-123",
						CreatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
						UpdatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
					},
					{
						ID:        "dataset-2",
						Name:      "dataset-two",
						ProjectID: "project-123",
						CreatedAt: mustParseTime("2023-01-02T10:00:00Z"),
						UpdatedAt: mustParseTime("2023-01-02T10:00:00Z"),
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(datasets)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		params := ListParams{Page: 2, Limit: 5}
		result, err := datasetClient.List(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 2, len(result.Data))
		require.Equal(t, "dataset-1", result.Data[0].ID)
		require.Equal(t, "dataset-one", result.Data[0].Name)
		require.Equal(t, 15, result.Metadata.TotalItems)
		require.Equal(t, 3, result.Metadata.TotalPages)
	})

	t.Run("empty dataset list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			datasets := ListDatasets{
				Metadata: common.ListMetadata{
					Page:       1,
					Limit:      50,
					TotalItems: 0,
					TotalPages: 0,
				},
				Data: []Dataset{},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(datasets)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		result, err := datasetClient.List(ctx, ListParams{})
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 0, len(result.Data))
		require.Equal(t, 0, result.Metadata.TotalItems)
	})
}

func TestClient_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful create dataset", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v2/datasets", r.URL.Path)
			require.Equal(t, "POST", r.Method)

			var createReq CreateDatasetRequest
			err := json.NewDecoder(r.Body).Decode(&createReq)
			require.NoError(t, err)
			require.Equal(t, "new-dataset", createReq.Name)
			require.Equal(t, "A new test dataset", createReq.Description)

			dataset := Dataset{
				ID:          "dataset-created-123",
				Name:        createReq.Name,
				Description: createReq.Description,
				Metadata:    createReq.Metadata,
				ProjectID:   "project-456",
				CreatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
				UpdatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(dataset)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		createReq := &CreateDatasetRequest{
			Name:        "new-dataset",
			Description: "A new test dataset",
			Metadata:    map[string]interface{}{"version": "1.0"},
		}

		result, err := datasetClient.Create(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "dataset-created-123", result.ID)
		require.Equal(t, "new-dataset", result.Name)
		require.Equal(t, "A new test dataset", result.Description)
		require.Equal(t, "project-456", result.ProjectID)
	})

	t.Run("create with validation error", func(t *testing.T) {
		client := resty.New()
		datasetClient := NewClient(client)

		createReq := &CreateDatasetRequest{} // Missing name
		result, err := datasetClient.Create(ctx, createReq)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "'name' is required", err.Error())
	})

	t.Run("create with server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid request"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		createReq := &CreateDatasetRequest{Name: "test-dataset"}
		result, err := datasetClient.Create(ctx, createReq)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "400")
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
