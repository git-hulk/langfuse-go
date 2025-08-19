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

func TestQueueListParams_ToQueryString(t *testing.T) {
	tests := []struct {
		name   string
		params QueueListParams
		want   string
	}{
		{"with page and limit", QueueListParams{Page: 1, Limit: 10}, "page=1&limit=10"},
		{"with page only", QueueListParams{Page: 1}, "page=1"},
		{"with limit only", QueueListParams{Limit: 10}, "limit=10"},
		{"no params", QueueListParams{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.params.ToQueryString())
		})
	}
}

func TestCreateQueueRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		request CreateQueueRequest
		wantErr bool
		errMsg  string
	}{
		{
			"valid request",
			CreateQueueRequest{
				Name:           "test-queue",
				ScoreConfigIDs: []string{"config-1", "config-2"},
			},
			false,
			"",
		},
		{
			"missing name",
			CreateQueueRequest{
				ScoreConfigIDs: []string{"config-1"},
			},
			true,
			"'name' is required",
		},
		{
			"missing scoreConfigIDs",
			CreateQueueRequest{
				Name: "test-queue",
			},
			true,
			"'scoreConfigIDs' is required",
		},
		{
			"empty scoreConfigIDs",
			CreateQueueRequest{
				Name:           "test-queue",
				ScoreConfigIDs: []string{},
			},
			true,
			"'scoreConfigIDs' is required and cannot be empty",
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

func TestAssignmentRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		request AssignmentRequest
		wantErr bool
		errMsg  string
	}{
		{
			"valid request",
			AssignmentRequest{UserID: "user-123"},
			false,
			"",
		},
		{
			"missing userID",
			AssignmentRequest{},
			true,
			"'userID' is required",
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

func TestQueueClient_Get(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/annotation-queues/test-queue-id", r.URL.Path)
			queue := Queue{
				ID:             "test-queue-id",
				Name:           "Test Queue",
				ScoreConfigIDs: []string{"config-1", "config-2"},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(queue)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewQueueClient(cli)
	queue, err := client.Get(context.Background(), "test-queue-id")
	require.NoError(t, err)
	require.Equal(t, "test-queue-id", queue.ID)
	require.Equal(t, "Test Queue", queue.Name)
	require.Equal(t, []string{"config-1", "config-2"}, queue.ScoreConfigIDs)
}

func TestQueueClient_Get_MissingQueueID(t *testing.T) {
	cli := resty.New()
	client := NewQueueClient(cli)
	_, err := client.Get(context.Background(), "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "'queueID' is required")
}

func TestQueueClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/annotation-queues", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"meta":{"page":1,"limit":10,"totalItems":1,"totalPages":1},"data":[{"id":"test-queue-id","name":"Test Queue","scoreConfigIds":["config-1"],"createdAt":"2023-01-01T00:00:00Z","updatedAt":"2023-01-01T00:00:00Z"}]}`))
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewQueueClient(cli)
	queueList, err := client.List(context.Background(), QueueListParams{})
	require.NoError(t, err)
	require.Len(t, queueList.Data, 1)
	require.Equal(t, "test-queue-id", queueList.Data[0].ID)
	require.Equal(t, "Test Queue", queueList.Data[0].Name)
	require.Equal(t, []string{"config-1"}, queueList.Data[0].ScoreConfigIDs)
	// verify meta
	require.Equal(t, 1, queueList.Metadata.Page)
	require.Equal(t, 10, queueList.Metadata.Limit)
	require.Equal(t, 1, queueList.Metadata.TotalItems)
	require.Equal(t, 1, queueList.Metadata.TotalPages)
}

func TestQueueClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/annotation-queues", r.URL.Path)
			require.Equal(t, "POST", r.Method)
			var request CreateQueueRequest
			err := json.NewDecoder(r.Body).Decode(&request)
			require.NoError(t, err)
			require.Equal(t, "New Queue", request.Name)
			require.Equal(t, []string{"config-1", "config-2"}, request.ScoreConfigIDs)
			// Return the created queue with an ID
			queue := Queue{
				ID:             "created-queue-id",
				Name:           request.Name,
				Description:    request.Description,
				ScoreConfigIDs: request.ScoreConfigIDs,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(queue)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewQueueClient(cli)
	createRequest := &CreateQueueRequest{
		Name:           "New Queue",
		ScoreConfigIDs: []string{"config-1", "config-2"},
	}
	queue, err := client.Create(context.Background(), createRequest)
	require.NoError(t, err)
	require.Equal(t, "created-queue-id", queue.ID)
	require.Equal(t, "New Queue", queue.Name)
	require.Equal(t, []string{"config-1", "config-2"}, queue.ScoreConfigIDs)
}

func TestQueueClient_Create_ValidationError(t *testing.T) {
	cli := resty.New()
	client := NewQueueClient(cli)
	createRequest := &CreateQueueRequest{} // Missing required fields
	_, err := client.Create(context.Background(), createRequest)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'name' is required")
}

func TestQueueClient_CreateAssignment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/annotation-queues/test-queue-id/assignments", r.URL.Path)
			require.Equal(t, "POST", r.Method)
			var request AssignmentRequest
			err := json.NewDecoder(r.Body).Decode(&request)
			require.NoError(t, err)
			require.Equal(t, "user-123", request.UserID)
			// Return the assignment response
			response := CreateAssignmentResponse{
				UserID:    request.UserID,
				QueueID:   "test-queue-id",
				ProjectID: "project-456",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewQueueClient(cli)
	request := &AssignmentRequest{UserID: "user-123"}
	response, err := client.CreateAssignment(context.Background(), "test-queue-id", request)
	require.NoError(t, err)
	require.Equal(t, "user-123", response.UserID)
	require.Equal(t, "test-queue-id", response.QueueID)
	require.Equal(t, "project-456", response.ProjectID)
}

func TestQueueClient_CreateAssignment_MissingQueueID(t *testing.T) {
	cli := resty.New()
	client := NewQueueClient(cli)
	request := &AssignmentRequest{UserID: "user-123"}
	_, err := client.CreateAssignment(context.Background(), "", request)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'queueID' is required")
}

func TestQueueClient_DeleteAssignment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/annotation-queues/test-queue-id/assignments", r.URL.Path)
			require.Equal(t, "DELETE", r.Method)
			var request AssignmentRequest
			err := json.NewDecoder(r.Body).Decode(&request)
			require.NoError(t, err)
			require.Equal(t, "user-123", request.UserID)
			// Return success response
			response := DeleteAssignmentResponse{Success: true}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewQueueClient(cli)
	request := &AssignmentRequest{UserID: "user-123"}
	response, err := client.DeleteAssignment(context.Background(), "test-queue-id", request)
	require.NoError(t, err)
	require.True(t, response.Success)
}

func TestQueueClient_DeleteAssignment_ValidationError(t *testing.T) {
	cli := resty.New()
	client := NewQueueClient(cli)
	request := &AssignmentRequest{} // Missing required fields
	_, err := client.DeleteAssignment(context.Background(), "test-queue-id", request)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'userID' is required")
}
