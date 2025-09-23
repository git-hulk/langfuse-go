package datasets

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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

// TestClient_CreateDatasetRunItems tests the CreateDatasetRunItems method.
func TestClient_CreateDatasetRunItems(t *testing.T) {
	ctx := context.Background()

	t.Run("successful create dataset run items", func(t *testing.T) {
		runName := "test-run"
		datasetItemId := "dataset-item-1"
		traceId := "trace-1"
		observationId := "observation-1"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/dataset-run-items", r.URL.Path)
			require.Equal(t, "POST", r.Method)

			var req CreateDatasetRunItemRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			require.Equal(t, runName, req.RunName)
			require.Equal(t, datasetItemId, req.DatasetItemID)
			require.Equal(t, traceId, req.TraceID)
			require.Equal(t, observationId, req.ObservationID)

			runItem := DatasetRunItem{
				ID:             "run-item-1",
				DatasetRunID:   "run-123",
				DatasetRunName: runName,
				DatasetItemID:  datasetItemId,
				TraceID:        traceId,
				ObservationID:  observationId,
				CreatedAt:      mustParseTime("2023-01-01T10:00:00Z"),
				UpdatedAt:      mustParseTime("2023-01-01T10:00:00Z"),
			}

			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(runItem)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		req := CreateDatasetRunItemRequest{
			RunName:       runName,
			DatasetItemID: datasetItemId,
			TraceID:       traceId,
			ObservationID: observationId,
		}
		result, err := datasetClient.CreateDatasetRunItems(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "run-item-1", result.ID)
		require.Equal(t, "run-123", result.DatasetRunID)
		require.Equal(t, runName, result.DatasetRunName)
		require.Equal(t, datasetItemId, result.DatasetItemID)
		require.Equal(t, traceId, result.TraceID)
		require.Equal(t, observationId, result.ObservationID)
	})

	t.Run("create dataset run items with empty runName", func(t *testing.T) {
		client := resty.New()
		datasetClient := NewClient(client)

		req := CreateDatasetRunItemRequest{
			TraceID: "trace-1",
		}
		result, err := datasetClient.CreateDatasetRunItems(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "'runName' is required", err.Error())
	})

	t.Run("create dataset run items with empty traceId", func(t *testing.T) {
		client := resty.New()
		datasetClient := NewClient(client)

		req := CreateDatasetRunItemRequest{
			RunName: "test-run",
		}
		result, err := datasetClient.CreateDatasetRunItems(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "'traceId' is required", err.Error())
	})

	t.Run("create dataset run items with HTTP error", func(t *testing.T) {
		client := resty.New().SetBaseURL("http://non-existent-domain-123456.com")
		datasetClient := NewClient(client)

		req := CreateDatasetRunItemRequest{
			RunName: "test-run",
			TraceID: "trace-1",
		}
		result, err := datasetClient.CreateDatasetRunItems(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("create dataset run items with API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(`{"error": "Invalid request"}`))
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		req := CreateDatasetRunItemRequest{
			RunName: "test-run",
			TraceID: "trace-1",
		}
		result, err := datasetClient.CreateDatasetRunItems(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "failed to dataset run items")
		require.Contains(t, err.Error(), "400")
	})
}

// TestClient_ListDatasetRunItems tests the ListDatasetRunItems method.
func TestClient_ListDatasetRunItems(t *testing.T) {
	ctx := context.Background()

	// Test successful retrieval of dataset run items list
	t.Run("successful list dataset run items", func(t *testing.T) {
		datasetId := "dataset-123"
		runName := "test-run"
		page := 1
		limit := 10

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/dataset-run-items", r.URL.Path)
			require.Equal(t, "GET", r.Method)

			// Verify query parameters
			query := r.URL.Query()
			require.Equal(t, datasetId, query.Get("datasetId"))
			require.Equal(t, runName, query.Get("runName"))
			require.Equal(t, strconv.Itoa(page), query.Get("page"))
			require.Equal(t, strconv.Itoa(limit), query.Get("limit"))

			// Return mock response
			listResponse := ListDatasetRunItems{
				Metadata: common.ListMetadata{
					Page:       page,
					Limit:      limit,
					TotalItems: 2,
					TotalPages: 1,
				},
				Data: []DatasetRunItem{
					{
						ID:             "run-item-1",
						DatasetRunID:   "run-123",
						DatasetRunName: runName,
						DatasetItemID:  "dataset-item-1",
						TraceID:        "trace-1",
						CreatedAt:      mustParseTime("2023-01-01T10:00:00Z"),
						UpdatedAt:      mustParseTime("2023-01-01T10:00:00Z"),
					},
					{
						ID:             "run-item-2",
						DatasetRunID:   "run-123",
						DatasetRunName: runName,
						DatasetItemID:  "dataset-item-2",
						TraceID:        "trace-2",
						CreatedAt:      mustParseTime("2023-01-01T10:01:00Z"),
						UpdatedAt:      mustParseTime("2023-01-01T10:01:00Z"),
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(listResponse)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		params := ListDatasetRunItemsParams{
			DatasetID: datasetId,
			RunName:   runName,
			Page:      page,
			Limit:     limit,
		}
		result, err := datasetClient.ListDatasetRunItems(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, page, result.Metadata.Page)
		require.Equal(t, limit, result.Metadata.Limit)
		require.Equal(t, 2, result.Metadata.TotalItems)
		require.Equal(t, 2, len(result.Data))
		require.Equal(t, "run-item-1", result.Data[0].ID)
		require.Equal(t, "run-item-2", result.Data[1].ID)
	})

	// Test case where datasetId is missing
	t.Run("list dataset run items with empty datasetId", func(t *testing.T) {
		client := resty.New()
		datasetClient := NewClient(client)

		params := ListDatasetRunItemsParams{
			RunName: "test-run",
		}
		result, err := datasetClient.ListDatasetRunItems(ctx, params)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "'datasetId' is required", err.Error())
	})

	// Test case where runName is missing
	t.Run("list dataset run items with empty runName", func(t *testing.T) {
		client := resty.New()
		datasetClient := NewClient(client)

		params := ListDatasetRunItemsParams{
			DatasetID: "dataset-123",
		}
		result, err := datasetClient.ListDatasetRunItems(ctx, params)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "'runName' is required", err.Error())
	})

	// Test case where HTTP request fails
	t.Run("list dataset run items with HTTP error", func(t *testing.T) {
		client := resty.New().SetBaseURL("http://non-existent-domain-123456.com")
		datasetClient := NewClient(client)

		params := ListDatasetRunItemsParams{
			DatasetID: "dataset-123",
			RunName:   "test-run",
		}
		result, err := datasetClient.ListDatasetRunItems(ctx, params)
		require.Error(t, err)
		require.Nil(t, result)
	})

	// Test case where API returns an error status code
	t.Run("list dataset run items with API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(`{"error": "Invalid request"}`))
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		params := ListDatasetRunItemsParams{
			DatasetID: "dataset-123",
			RunName:   "test-run",
		}
		result, err := datasetClient.ListDatasetRunItems(ctx, params)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "list dataset run items failed")
		require.Contains(t, err.Error(), "400")
	})

	// Test default pagination parameters (page=0, limit=0)
	t.Run("list dataset run items with default pagination params", func(t *testing.T) {
		datasetId := "dataset-123"
		runName := "test-run"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/dataset-run-items", r.URL.Path)
			query := r.URL.Query()
			require.Equal(t, datasetId, query.Get("datasetId"))
			require.Equal(t, runName, query.Get("runName"))
			require.Empty(t, query.Get("page"))
			require.Empty(t, query.Get("limit"))

			listResponse := ListDatasetRunItems{
				Metadata: common.ListMetadata{
					Page:       1,
					Limit:      10,
					TotalItems: 1,
					TotalPages: 1,
				},
				Data: []DatasetRunItem{
					{
						ID:             "run-item-1",
						DatasetRunID:   "run-123",
						DatasetRunName: runName,
						DatasetItemID:  "dataset-item-1",
						TraceID:        "trace-1",
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(listResponse)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		datasetClient := NewClient(client)

		params := ListDatasetRunItemsParams{
			DatasetID: datasetId,
			RunName:   runName,
			Page:      0,
			Limit:     0,
		}
		result, err := datasetClient.ListDatasetRunItems(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 1, len(result.Data))
	})
}

// TestCreateDatasetRunItemRequest_validate tests the validate method of CreateDatasetRunItemRequest.
func TestCreateDatasetRunItemRequest_validate(t *testing.T) {
	// Test successful validation - all required fields are provided
	t.Run("successful validation with all required fields", func(t *testing.T) {
		req := CreateDatasetRunItemRequest{
			RunName: "test-run",
			TraceID: "test-trace-id",
		}
		err := req.validate()
		require.NoError(t, err)
	})

	// Test case with missing RunName
	t.Run("validation fails with empty RunName", func(t *testing.T) {
		req := CreateDatasetRunItemRequest{
			TraceID: "test-trace-id",
		}
		err := req.validate()
		require.Error(t, err)
		require.Equal(t, "'runName' is required", err.Error())
	})

	// Test case with missing TraceID
	t.Run("validation fails with empty TraceID", func(t *testing.T) {
		req := CreateDatasetRunItemRequest{
			RunName: "test-run",
		}
		err := req.validate()
		require.Error(t, err)
		require.Equal(t, "'traceId' is required", err.Error())
	})

	// Test case with both required fields missing
	t.Run("validation fails with empty both RunName and TraceID", func(t *testing.T) {
		req := CreateDatasetRunItemRequest{}
		err := req.validate()
		require.Error(t, err)
		require.Equal(t, "'runName' is required", err.Error())
	})

	// Test case with optional fields provided but required fields missing
	t.Run("validation fails with optional fields but missing required fields", func(t *testing.T) {
		req := CreateDatasetRunItemRequest{
			DatasetItemID: "test-dataset-item-id",
			Metadata:      map[string]string{"key": "value"},
			ObservationID: "test-observation-id",
			// Missing RunName and TraceID
		}
		err := req.validate()
		require.Error(t, err)
		require.Equal(t, "'runName' is required", err.Error())
	})

	// Test case with TraceID and optional fields provided but missing RunName
	t.Run("validation fails with TraceID and optional fields but missing RunName", func(t *testing.T) {
		req := CreateDatasetRunItemRequest{
			DatasetItemID: "test-dataset-item-id",
			Metadata:      map[string]string{"key": "value"},
			ObservationID: "test-observation-id",
			TraceID:       "test-trace-id",
			// Missing RunName
		}
		err := req.validate()
		require.Error(t, err)
		require.Equal(t, "'runName' is required", err.Error())
	})

	// Test case with RunName and optional fields provided but missing TraceID
	t.Run("validation fails with RunName and optional fields but missing TraceID", func(t *testing.T) {
		req := CreateDatasetRunItemRequest{
			RunName:       "test-run",
			DatasetItemID: "test-dataset-item-id",
			Metadata:      map[string]string{"key": "value"},
			ObservationID: "test-observation-id",
			// Missing TraceID
		}
		err := req.validate()
		require.Error(t, err)
		require.Equal(t, "'traceId' is required", err.Error())
	})
}

// TestListDatasetRunItemsParams_ToQueryString tests the ToQueryString method of ListDatasetRunItemsParams.
func TestListDatasetRunItemsParams_ToQueryString(t *testing.T) {
	// Test with all parameters set
	t.Run("convert with all parameters set", func(t *testing.T) {
		params := ListDatasetRunItemsParams{
			DatasetID: "dataset-123",
			RunName:   "test-run",
			Page:      2,
			Limit:     20,
		}
		queryStr := params.ToQueryString()

		// Verify all parameters are included in the query string
		require.Contains(t, queryStr, "datasetId=dataset-123")
		require.Contains(t, queryStr, "runName=test-run")
		require.Contains(t, queryStr, "page=2")
		require.Contains(t, queryStr, "limit=20")
	})

	// Test with only required parameters set
	t.Run("convert with only required parameters", func(t *testing.T) {
		params := ListDatasetRunItemsParams{
			DatasetID: "dataset-456",
			RunName:   "another-run",
			// Page and Limit are 0 (default)
		}
		queryStr := params.ToQueryString()

		// Verify only non-zero/non-empty parameters are included
		require.Contains(t, queryStr, "datasetId=dataset-456")
		require.Contains(t, queryStr, "runName=another-run")
		require.NotContains(t, queryStr, "page")
		require.NotContains(t, queryStr, "limit")
	})

	// Test with only pagination parameters set
	t.Run("convert with only pagination parameters", func(t *testing.T) {
		params := ListDatasetRunItemsParams{
			// DatasetID and RunName are empty
			Page:  3,
			Limit: 50,
		}
		queryStr := params.ToQueryString()

		// Verify only non-zero/non-empty parameters are included
		require.Contains(t, queryStr, "page=3")
		require.Contains(t, queryStr, "limit=50")
		require.NotContains(t, queryStr, "datasetId")
		require.NotContains(t, queryStr, "runName")
	})

	// Test with no parameters set (all default values)
	t.Run("convert with no parameters set", func(t *testing.T) {
		params := ListDatasetRunItemsParams{}
		queryStr := params.ToQueryString()

		// Verify the query string is empty when all parameters are default
		require.Empty(t, queryStr)
	})

	// Test with DatasetID and Page set
	t.Run("convert with DatasetID and Page set", func(t *testing.T) {
		params := ListDatasetRunItemsParams{
			DatasetID: "dataset-789",
			// RunName is empty
			Page: 1,
			// Limit is 0
		}
		queryStr := params.ToQueryString()

		// Verify only non-zero/non-empty parameters are included
		require.Contains(t, queryStr, "datasetId=dataset-789")
		require.Contains(t, queryStr, "page=1")
		require.NotContains(t, queryStr, "runName")
		require.NotContains(t, queryStr, "limit")
	})

	// Test with RunName and Limit set
	t.Run("convert with RunName and Limit set", func(t *testing.T) {
		params := ListDatasetRunItemsParams{
			// DatasetID is empty
			RunName: "special-run",
			// Page is 0
			Limit: 100,
		}
		queryStr := params.ToQueryString()

		// Verify only non-zero/non-empty parameters are included
		require.Contains(t, queryStr, "runName=special-run")
		require.Contains(t, queryStr, "limit=100")
		require.NotContains(t, queryStr, "datasetId")
		require.NotContains(t, queryStr, "page")
	})

	// Test with negative Page value
	t.Run("convert with negative Page value", func(t *testing.T) {
		params := ListDatasetRunItemsParams{
			Page: -1,
		}
		queryStr := params.ToQueryString()

		// Verify negative Page value is included (since it's not zero)
		require.Contains(t, queryStr, "page=-1")
	})

	// Test with zero Limit but non-zero Page
	t.Run("convert with zero Limit but non-zero Page", func(t *testing.T) {
		params := ListDatasetRunItemsParams{
			Page:  4,
			Limit: 0,
		}
		queryStr := params.ToQueryString()

		// Verify only non-zero parameters are included
		require.Contains(t, queryStr, "page=4")
		require.NotContains(t, queryStr, "limit")
	})

	// Test with URL encoding of special characters
	t.Run("convert with URL encoding of special characters", func(t *testing.T) {
		params := ListDatasetRunItemsParams{
			DatasetID: "dataset/with/slashes",
			RunName:   "run with spaces & symbols",
		}
		queryStr := params.ToQueryString()

		// Verify special characters are properly URL encoded
		require.Contains(t, queryStr, "datasetId=dataset%2Fwith%2Fslashes")
		require.Contains(t, queryStr, "runName=run+with+spaces+%26+symbols")
	})
}
