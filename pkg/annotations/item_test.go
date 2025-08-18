package annotations

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestItemListParams_ToQueryString(t *testing.T) {
	tests := []struct {
		name   string
		params ItemListParams
		want   string
	}{
		{"with all params", ItemListParams{Status: StatusPending, Page: 1, Limit: 10}, "status=PENDING&page=1&limit=10"},
		{"with status only", ItemListParams{Status: StatusCompleted}, "status=COMPLETED"},
		{"with page and limit", ItemListParams{Page: 1, Limit: 10}, "page=1&limit=10"},
		{"no params", ItemListParams{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.params.ToQueryString())
		})
	}
}

func TestCreateItemRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		request CreateItemRequest
		wantErr bool
		errMsg  string
	}{
		{
			"valid request",
			CreateItemRequest{
				ObjectID:   "trace-123",
				ObjectType: ObjectTypeTrace,
			},
			false,
			"",
		},
		{
			"valid request with status",
			CreateItemRequest{
				ObjectID:   "obs-456",
				ObjectType: ObjectTypeObservation,
				Status:     StatusPending,
			},
			false,
			"",
		},
		{
			"missing objectId",
			CreateItemRequest{
				ObjectType: ObjectTypeTrace,
			},
			true,
			"'objectId' is required",
		},
		{
			"missing objectType",
			CreateItemRequest{
				ObjectID: "trace-123",
			},
			true,
			"'objectType' is required",
		},
		{
			"invalid objectType",
			CreateItemRequest{
				ObjectID:   "trace-123",
				ObjectType: QueueObjectType("INVALID"),
			},
			true,
			"invalid 'objectType': INVALID",
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

func TestUpdateItemRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateItemRequest
		wantErr bool
		errMsg  string
	}{
		{
			"valid request",
			UpdateItemRequest{
				Status: StatusCompleted,
			},
			false,
			"",
		},
		{
			"empty request",
			UpdateItemRequest{},
			false,
			"",
		},
		{
			"invalid status",
			UpdateItemRequest{
				Status: QueueStatus("INVALID"),
			},
			true,
			"invalid 'status': INVALID",
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

func TestItemClient_Get(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/annotation-queues/test-queue-id/items/test-item-id", r.URL.Path)
			item := Item{
				ID:         "test-item-id",
				QueueID:    "test-queue-id",
				ObjectID:   "trace-123",
				ObjectType: ObjectTypeTrace,
				Status:     StatusPending,
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(item)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewItemClient(cli)
	item, err := client.Get(context.Background(), "test-queue-id", "test-item-id")
	require.NoError(t, err)
	require.Equal(t, "test-item-id", item.ID)
	require.Equal(t, "test-queue-id", item.QueueID)
	require.Equal(t, "trace-123", item.ObjectID)
	require.Equal(t, ObjectTypeTrace, item.ObjectType)
	require.Equal(t, StatusPending, item.Status)
}

func TestItemClient_Get_MissingParams(t *testing.T) {
	cli := resty.New()
	client := NewItemClient(cli)

	// Test missing queueID
	_, err := client.Get(context.Background(), "", "test-item-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "'queueID' is required")

	// Test missing itemID
	_, err = client.Get(context.Background(), "test-queue-id", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "'itemID' is required")
}

func TestItemClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/annotation-queues/test-queue-id/items", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"meta":{"page":1,"limit":10,"totalItems":1,"totalPages":1},"data":[{"id":"test-item-id","queueId":"test-queue-id","objectId":"trace-123","objectType":"TRACE","status":"PENDING","createdAt":"2023-01-01T00:00:00Z","updatedAt":"2023-01-01T00:00:00Z"}]}`))
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewItemClient(cli)
	itemList, err := client.List(context.Background(), "test-queue-id", ItemListParams{})
	require.NoError(t, err)
	require.Len(t, itemList.Data, 1)
	require.Equal(t, "test-item-id", itemList.Data[0].ID)
	require.Equal(t, "test-queue-id", itemList.Data[0].QueueID)
	require.Equal(t, "trace-123", itemList.Data[0].ObjectID)
	require.Equal(t, ObjectTypeTrace, itemList.Data[0].ObjectType)
	require.Equal(t, StatusPending, itemList.Data[0].Status)
	// verify meta
	require.Equal(t, 1, itemList.Metadata.Page)
	require.Equal(t, 10, itemList.Metadata.Limit)
	require.Equal(t, 1, itemList.Metadata.TotalItems)
	require.Equal(t, 1, itemList.Metadata.TotalPages)
}

func TestItemClient_List_MissingQueueID(t *testing.T) {
	cli := resty.New()
	client := NewItemClient(cli)
	_, err := client.List(context.Background(), "", ItemListParams{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "'queueID' is required")
}

func TestItemClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/annotation-queues/test-queue-id/items", r.URL.Path)
			require.Equal(t, "POST", r.Method)
			var request CreateItemRequest
			err := json.NewDecoder(r.Body).Decode(&request)
			require.NoError(t, err)
			require.Equal(t, "trace-123", request.ObjectID)
			require.Equal(t, ObjectTypeTrace, request.ObjectType)
			// Return the created item with an ID
			item := Item{
				ID:         "created-item-id",
				QueueID:    "test-queue-id",
				ObjectID:   request.ObjectID,
				ObjectType: request.ObjectType,
				Status:     StatusPending,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(item)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewItemClient(cli)
	createRequest := &CreateItemRequest{
		ObjectID:   "trace-123",
		ObjectType: ObjectTypeTrace,
	}
	item, err := client.Create(context.Background(), "test-queue-id", createRequest)
	require.NoError(t, err)
	require.Equal(t, "created-item-id", item.ID)
	require.Equal(t, "test-queue-id", item.QueueID)
	require.Equal(t, "trace-123", item.ObjectID)
	require.Equal(t, ObjectTypeTrace, item.ObjectType)
	require.Equal(t, StatusPending, item.Status)
}

func TestItemClient_Create_ValidationError(t *testing.T) {
	cli := resty.New()
	client := NewItemClient(cli)
	createRequest := &CreateItemRequest{} // Missing required fields
	_, err := client.Create(context.Background(), "test-queue-id", createRequest)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'objectId' is required")
}

func TestItemClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/annotation-queues/test-queue-id/items/test-item-id", r.URL.Path)
			require.Equal(t, "PATCH", r.Method)
			var request UpdateItemRequest
			err := json.NewDecoder(r.Body).Decode(&request)
			require.NoError(t, err)
			require.Equal(t, StatusCompleted, request.Status)
			// Return the updated item
			item := Item{
				ID:         "test-item-id",
				QueueID:    "test-queue-id",
				ObjectID:   "trace-123",
				ObjectType: ObjectTypeTrace,
				Status:     StatusCompleted,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(item)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewItemClient(cli)
	updateRequest := &UpdateItemRequest{
		Status: StatusCompleted,
	}
	item, err := client.Update(context.Background(), "test-queue-id", "test-item-id", updateRequest)
	require.NoError(t, err)
	require.Equal(t, "test-item-id", item.ID)
	require.Equal(t, StatusCompleted, item.Status)
}

func TestItemClient_Update_MissingParams(t *testing.T) {
	cli := resty.New()
	client := NewItemClient(cli)
	updateRequest := &UpdateItemRequest{Status: StatusCompleted}

	// Test missing queueID
	_, err := client.Update(context.Background(), "", "test-item-id", updateRequest)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'queueID' is required")

	// Test missing itemID
	_, err = client.Update(context.Background(), "test-queue-id", "", updateRequest)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'itemID' is required")
}

func TestItemClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/annotation-queues/test-queue-id/items/test-item-id", r.URL.Path)
			require.Equal(t, "DELETE", r.Method)
			// Return delete response
			response := DeleteItemResponse{
				Success: true,
				Message: "Item deleted successfully",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewItemClient(cli)
	response, err := client.Delete(context.Background(), "test-queue-id", "test-item-id")
	require.NoError(t, err)
	require.True(t, response.Success)
	require.Equal(t, "Item deleted successfully", response.Message)
}

func TestItemClient_Delete_MissingParams(t *testing.T) {
	cli := resty.New()
	client := NewItemClient(cli)

	// Test missing queueID
	_, err := client.Delete(context.Background(), "", "test-item-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "'queueID' is required")

	// Test missing itemID
	_, err = client.Delete(context.Background(), "test-queue-id", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "'itemID' is required")
}
