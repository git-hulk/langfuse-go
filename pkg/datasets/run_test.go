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

func TestClient_GetDatasetRuns(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get runs", func(t *testing.T) {
		datasetName := "test-dataset"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/datasets/"+datasetName+"/runs", r.URL.Path)
			require.Equal(t, "GET", r.Method)

			query := r.URL.Query()
			require.Equal(t, "1", query.Get("page"))
			require.Equal(t, "10", query.Get("limit"))

			runs := ListDatasetRuns{
				Metadata: common.ListMetadata{
					Page:       1,
					Limit:      10,
					TotalItems: 2,
					TotalPages: 1,
				},
				Data: []DatasetRun{
					{
						ID:          "run-1",
						Name:        "test-run-1",
						Description: "First test run",
						DatasetID:   "dataset-123",
						DatasetName: datasetName,
						CreatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
						UpdatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(runs)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		params := ListParams{Page: 1, Limit: 10}
		result, err := datasetClient.GetDatasetRuns(ctx, datasetName, params)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 1, len(result.Data))
		require.Equal(t, "run-1", result.Data[0].ID)
		require.Equal(t, "test-run-1", result.Data[0].Name)
		require.Equal(t, datasetName, result.Data[0].DatasetName)
	})

	t.Run("get runs with empty dataset name", func(t *testing.T) {
		client := resty.New()
		datasetClient := NewClient(client)

		result, err := datasetClient.GetDatasetRuns(ctx, "", ListParams{})
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "'datasetName' is required", err.Error())
	})
}

func TestClient_GetRun(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get run", func(t *testing.T) {
		datasetName := "test-dataset"
		runName := "test-run"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/datasets/"+datasetName+"/runs/"+runName, r.URL.Path)
			require.Equal(t, "GET", r.Method)

			runWithItems := DatasetRunWithItems{
				DatasetRun: DatasetRun{
					ID:          "run-123",
					Name:        runName,
					Description: "Test run",
					DatasetID:   "dataset-456",
					DatasetName: datasetName,
					CreatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
					UpdatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
				},
				DatasetRunItems: []DatasetRunItem{
					{
						ID:             "item-1",
						DatasetRunID:   "run-123",
						DatasetRunName: runName,
						DatasetItemID:  "dataset-item-1",
						TraceID:        "trace-1",
						CreatedAt:      mustParseTime("2023-01-01T10:00:00Z"),
						UpdatedAt:      mustParseTime("2023-01-01T10:00:00Z"),
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(runWithItems)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		result, err := datasetClient.GetDatasetRun(ctx, datasetName, runName)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "run-123", result.ID)
		require.Equal(t, runName, result.Name)
		require.Equal(t, datasetName, result.DatasetName)
		require.Equal(t, 1, len(result.DatasetRunItems))
		require.Equal(t, "item-1", result.DatasetRunItems[0].ID)
	})

	t.Run("get run with empty dataset name", func(t *testing.T) {
		client := resty.New()
		datasetClient := NewClient(client)

		result, err := datasetClient.GetDatasetRun(ctx, "", "run-name")
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "'datasetName' is required", err.Error())
	})

	t.Run("get run with empty run name", func(t *testing.T) {
		client := resty.New()
		datasetClient := NewClient(client)

		result, err := datasetClient.GetDatasetRun(ctx, "dataset-name", "")
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "'runName' is required", err.Error())
	})
}

func TestClient_DeleteRun(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete run", func(t *testing.T) {
		datasetName := "test-dataset"
		runName := "test-run"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/datasets/"+datasetName+"/runs/"+runName, r.URL.Path)
			require.Equal(t, "DELETE", r.Method)

			deleteResponse := DeleteDatasetRunResponse{
				Message: "Dataset run deleted successfully",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(deleteResponse)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		result, err := datasetClient.DeleteDatasetRun(ctx, datasetName, runName)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "Dataset run deleted successfully", result.Message)
	})

	t.Run("delete run with empty dataset name", func(t *testing.T) {
		client := resty.New()
		datasetClient := NewClient(client)

		result, err := datasetClient.DeleteDatasetRun(ctx, "", "run-name")
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "'datasetName' is required", err.Error())
	})

	t.Run("delete run with empty run name", func(t *testing.T) {
		client := resty.New()
		datasetClient := NewClient(client)

		result, err := datasetClient.DeleteDatasetRun(ctx, "dataset-name", "")
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "'runName' is required", err.Error())
	})
}
