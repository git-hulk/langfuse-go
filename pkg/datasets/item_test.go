package datasets

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"

	"github.com/git-hulk/langfuse-go/pkg/common"
)

func TestCreateDatasetItemRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		request CreateDatasetItemRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: CreateDatasetItemRequest{
				DatasetName:    "test-dataset",
				Input:          map[string]any{"text": "hello world"},
				ExpectedOutput: map[string]any{"response": "hello back"},
			},
			wantErr: false,
		},
		{
			name: "missing dataset name",
			request: CreateDatasetItemRequest{
				Input:          map[string]any{"text": "hello world"},
				ExpectedOutput: map[string]any{"response": "hello back"},
			},
			wantErr: true,
		},
		{
			name: "empty dataset name",
			request: CreateDatasetItemRequest{
				DatasetName:    "",
				Input:          map[string]any{"text": "hello world"},
				ExpectedOutput: map[string]any{"response": "hello back"},
			},
			wantErr: true,
		},
		{
			name: "dataset name only",
			request: CreateDatasetItemRequest{
				DatasetName: "test-dataset",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestListDatasetItemParams_ToQueryString(t *testing.T) {
	tests := []struct {
		name   string
		params ListDatasetItemParams
		want   string
	}{
		{
			name:   "empty params",
			params: ListDatasetItemParams{},
			want:   "",
		},
		{
			name: "all params",
			params: ListDatasetItemParams{
				Page:        1,
				Limit:       10,
				DatasetName: "test-dataset",
			},
			want: "datasetName=test-dataset&page=1&limit=10",
		},
		{
			name: "partial params",
			params: ListDatasetItemParams{
				Page:  2,
				Limit: 20,
			},
			want: "page=2&limit=20",
		},
		{
			name: "dataset name only",
			params: ListDatasetItemParams{
				DatasetName: "my-dataset",
			},
			want: "datasetName=my-dataset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.params.ToQueryString()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDatasetItemClientMethods(t *testing.T) {
	client := NewClient(nil)
	require.NotNil(t, client)

	ctx := context.Background()

	t.Run("GetDatasetItem requires ID", func(t *testing.T) {
		_, err := client.GetDatasetItem(ctx, "")
		require.Error(t, err)
	})

	t.Run("DeleteDatasetItem requires ID", func(t *testing.T) {
		err := client.DeleteDatasetItem(ctx, "")
		require.Error(t, err)
	})

	t.Run("CreateDatasetItem validates request", func(t *testing.T) {
		_, err := client.CreateDatasetItem(ctx, &CreateDatasetItemRequest{})
		require.Error(t, err)

		validRequest := &CreateDatasetItemRequest{
			DatasetName:    "test-dataset",
			Input:          map[string]any{"text": "hello world"},
			ExpectedOutput: map[string]any{"response": "hello back"},
		}
		err = validRequest.validate()
		require.NoError(t, err)
	})
}

func TestDatasetItemStructures(t *testing.T) {
	t.Run("DatasetItem creation", func(t *testing.T) {
		item := DatasetItem{
			ID:             "item-123",
			DatasetName:    "test-dataset",
			Input:          map[string]any{"text": "hello"},
			ExpectedOutput: map[string]any{"response": "hi"},
			Metadata:       map[string]any{"model": "gpt-4"},
		}

		require.Equal(t, "item-123", item.ID)
		require.Equal(t, "test-dataset", item.DatasetName)
	})

	t.Run("CreateDatasetItemRequest with optional fields", func(t *testing.T) {
		request := CreateDatasetItemRequest{
			DatasetName:         "test-dataset",
			Input:               "simple string input",
			ExpectedOutput:      "simple string output",
			Metadata:            map[string]any{"version": "1.0"},
			ID:                  "custom-id",
			SourceTraceID:       "trace-123",
			SourceObservationID: "obs-456",
			Status:              "active",
		}

		err := request.validate()
		require.NoError(t, err)
		require.Equal(t, "custom-id", request.ID)
		require.Equal(t, "trace-123", request.SourceTraceID)
	})
}

func TestDatasetItemClient_List(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list with all parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/dataset-items", r.URL.Path)
			query := r.URL.Query()
			require.Equal(t, "test-dataset", query.Get("datasetName"))
			require.Equal(t, "1", query.Get("page"))
			require.Equal(t, "10", query.Get("limit"))
			require.Equal(t, "trace-123", query.Get("sourceTraceId"))
			require.Equal(t, "obs-456", query.Get("sourceObservationId"))

			mockResponse := ListDatasetItems{
				Metadata: common.ListMetadata{
					Page:       1,
					Limit:      10,
					TotalItems: 25,
					TotalPages: 3,
				},
				Data: []DatasetItem{
					{
						ID:             "item-1",
						DatasetName:    "test-dataset",
						Input:          map[string]any{"text": "hello"},
						ExpectedOutput: map[string]any{"response": "hi"},
						Metadata:       map[string]any{"model": "gpt-4"},
						Status:         "active",
					},
					{
						ID:             "item-2",
						DatasetName:    "test-dataset",
						Input:          map[string]any{"text": "goodbye"},
						ExpectedOutput: map[string]any{"response": "bye"},
						Metadata:       map[string]any{"model": "gpt-3.5"},
						Status:         "active",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(mockResponse)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		params := ListDatasetItemParams{
			DatasetName:         "test-dataset",
			Page:                1,
			Limit:               10,
			SourceTraceID:       "trace-123",
			SourceObservationID: "obs-456",
		}

		result, err := datasetClient.ListDatasetItems(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 2, len(result.Data))
		require.Equal(t, "item-1", result.Data[0].ID)
		require.Equal(t, "test-dataset", result.Data[0].DatasetName)
		require.Equal(t, 25, result.Metadata.TotalItems)
		require.Equal(t, 3, result.Metadata.TotalPages)
	})

	t.Run("successful list with minimal parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/dataset-items", r.URL.Path)
			query := r.URL.Query()
			require.Equal(t, "minimal-dataset", query.Get("datasetName"))
			require.Empty(t, query.Get("page"))
			require.Empty(t, query.Get("limit"))

			mockResponse := ListDatasetItems{
				Metadata: common.ListMetadata{
					Page:       1,
					Limit:      50,
					TotalItems: 5,
					TotalPages: 1,
				},
				Data: []DatasetItem{
					{
						ID:          "item-1",
						DatasetName: "minimal-dataset",
						Input:       "simple input",
						Status:      "active",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(mockResponse)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		params := ListDatasetItemParams{
			DatasetName: "minimal-dataset",
		}

		result, err := datasetClient.ListDatasetItems(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 1, len(result.Data))
		require.Equal(t, "item-1", result.Data[0].ID)
		require.Equal(t, 5, result.Metadata.TotalItems)
	})

	t.Run("empty list response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/dataset-items", r.URL.Path)

			mockResponse := ListDatasetItems{
				Metadata: common.ListMetadata{
					Page:       1,
					Limit:      50,
					TotalItems: 0,
					TotalPages: 0,
				},
				Data: []DatasetItem{},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(mockResponse)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		params := ListDatasetItemParams{
			DatasetName: "empty-dataset",
		}

		result, err := datasetClient.ListDatasetItems(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 0, len(result.Data))
		require.Equal(t, 0, result.Metadata.TotalItems)
	})

	t.Run("list with no dataset name", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/dataset-items", r.URL.Path)

			mockResponse := ListDatasetItems{
				Metadata: common.ListMetadata{
					Page:       1,
					Limit:      50,
					TotalItems: 10,
					TotalPages: 1,
				},
				Data: []DatasetItem{
					{ID: "item-1", DatasetName: "dataset-1"},
					{ID: "item-2", DatasetName: "dataset-2"},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(mockResponse)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		params := ListDatasetItemParams{
			Page:  1,
			Limit: 50,
		}

		result, err := datasetClient.ListDatasetItems(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 2, len(result.Data))
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		params := ListDatasetItemParams{
			DatasetName: "test-dataset",
		}

		result, err := datasetClient.ListDatasetItems(ctx, params)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "500")
	})

	t.Run("not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Dataset not found"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		params := ListDatasetItemParams{
			DatasetName: "nonexistent-dataset",
		}

		result, err := datasetClient.ListDatasetItems(ctx, params)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "404")
	})
}

func TestDatasetItemClient_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		itemID := "item-123"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/dataset-items/"+itemID, r.URL.Path)
			require.Equal(t, "DELETE", r.Method)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		err := datasetClient.DeleteDatasetItem(ctx, itemID)
		require.NoError(t, err)
	})

	t.Run("delete with empty ID", func(t *testing.T) {
		client := resty.New()
		datasetClient := NewClient(client)

		err := datasetClient.DeleteDatasetItem(ctx, "")
		require.Error(t, err)
		require.Equal(t, "'id' is required", err.Error())
	})

	t.Run("delete nonexistent item", func(t *testing.T) {
		itemID := "nonexistent-item"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/dataset-items/"+itemID, r.URL.Path)
			require.Equal(t, "DELETE", r.Method)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Dataset item not found"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		err := datasetClient.DeleteDatasetItem(ctx, itemID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "404")
	})

	t.Run("delete with server error", func(t *testing.T) {
		itemID := "error-item"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/dataset-items/"+itemID, r.URL.Path)
			require.Equal(t, "DELETE", r.Method)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		err := datasetClient.DeleteDatasetItem(ctx, itemID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "500")
	})

	t.Run("delete with forbidden error", func(t *testing.T) {
		itemID := "forbidden-item"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/dataset-items/"+itemID, r.URL.Path)
			require.Equal(t, "DELETE", r.Method)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Access denied"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		err := datasetClient.DeleteDatasetItem(ctx, itemID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "403")
	})

	t.Run("delete with validation error", func(t *testing.T) {
		itemID := "validation-error-item"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/dataset-items/"+itemID, r.URL.Path)
			require.Equal(t, "DELETE", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid request"))
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		err := datasetClient.DeleteDatasetItem(ctx, itemID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "400")
	})
}
