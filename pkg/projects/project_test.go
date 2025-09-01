package projects

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestCreateProjectRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateProjectRequest
		wantErr bool
		errMsg  string
	}{
		{
			"valid request",
			CreateProjectRequest{Name: "test-project", Retention: 30},
			false,
			"",
		},
		{
			"missing name",
			CreateProjectRequest{Retention: 30},
			true,
			"'name' is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateProjectRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		req     UpdateProjectRequest
		wantErr bool
		errMsg  string
	}{
		{
			"valid request",
			UpdateProjectRequest{Name: "updated-project", Retention: 60},
			false,
			"",
		},
		{
			"missing name",
			UpdateProjectRequest{Retention: 60},
			true,
			"'name' is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProjectClient_Get(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/projects", r.URL.Path)
			require.Equal(t, "GET", r.Method)
			projects := ProjectsResponse{
				Data: []Project{
					{ID: "project-1", Name: "Test Project 1"},
					{ID: "project-2", Name: "Test Project 2"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(projects)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewClient(cli)
	projects, err := client.List(context.Background())
	require.NoError(t, err)
	require.Len(t, projects.Data, 2)
	require.Equal(t, "project-1", projects.Data[0].ID)
	require.Equal(t, "Test Project 1", projects.Data[0].Name)
	require.Equal(t, "project-2", projects.Data[1].ID)
	require.Equal(t, "Test Project 2", projects.Data[1].Name)
}

func TestProjectClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/projects", r.URL.Path)
			require.Equal(t, "POST", r.Method)

			var req CreateProjectRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			require.Equal(t, "test-project", req.Name)
			require.Equal(t, 30, req.Retention)

			// Return the created project
			project := Project{
				ID:            "created-project-id",
				Name:          req.Name,
				Metadata:      req.Metadata,
				RetentionDays: req.Retention,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(project)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewClient(cli)

	createReq := &CreateProjectRequest{
		Name:      "test-project",
		Metadata:  map[string]interface{}{"key": "value"},
		Retention: 30,
	}

	project, err := client.Create(context.Background(), createReq)
	require.NoError(t, err)
	require.Equal(t, "created-project-id", project.ID)
	require.Equal(t, "test-project", project.Name)
	require.Equal(t, 30, project.RetentionDays)
}

func TestProjectClient_Create_ValidationError(t *testing.T) {
	cli := resty.New()
	client := NewClient(cli)
	createReq := &CreateProjectRequest{} // Missing required name
	_, err := client.Create(context.Background(), createReq)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'name' is required")
}

func TestProjectClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/projects/test-project-id", r.URL.Path)
			require.Equal(t, "PUT", r.Method)

			var req UpdateProjectRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			require.Equal(t, "updated-project", req.Name)
			require.Equal(t, 60, req.Retention)

			// Return the updated project
			project := Project{
				ID:            "test-project-id",
				Name:          req.Name,
				Metadata:      req.Metadata,
				RetentionDays: req.Retention,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(project)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewClient(cli)

	updateReq := &UpdateProjectRequest{
		Name:      "updated-project",
		Metadata:  map[string]interface{}{"updated": "true"},
		Retention: 60,
	}

	project, err := client.Update(context.Background(), "test-project-id", updateReq)
	require.NoError(t, err)
	require.Equal(t, "test-project-id", project.ID)
	require.Equal(t, "updated-project", project.Name)
	require.Equal(t, 60, project.RetentionDays)
}

func TestProjectClient_Update_MissingProjectID(t *testing.T) {
	cli := resty.New()
	client := NewClient(cli)
	updateReq := &UpdateProjectRequest{Name: "test", Retention: 30}
	_, err := client.Update(context.Background(), "", updateReq)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'projectID' is required")
}

func TestProjectClient_Update_ValidationError(t *testing.T) {
	cli := resty.New()
	client := NewClient(cli)
	updateReq := &UpdateProjectRequest{} // Missing required name
	_, err := client.Update(context.Background(), "test-project-id", updateReq)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'name' is required")
}

func TestProjectClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/projects/test-project-id", r.URL.Path)
			require.Equal(t, "DELETE", r.Method)

			deleteResponse := ProjectDeletionResponse{
				Success: true,
				Message: "Project deletion initiated successfully",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			err := json.NewEncoder(w).Encode(deleteResponse)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewClient(cli)

	deleteResponse, err := client.Delete(context.Background(), "test-project-id")
	require.NoError(t, err)
	require.True(t, deleteResponse.Success)
	require.Equal(t, "Project deletion initiated successfully", deleteResponse.Message)
}

func TestProjectClient_Delete_MissingProjectID(t *testing.T) {
	cli := resty.New()
	client := NewClient(cli)
	_, err := client.Delete(context.Background(), "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "'projectID' is required")
}

func TestProjectClient_GetApiKeys(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/projects/test-project-id/apiKeys", r.URL.Path)
			require.Equal(t, "GET", r.Method)

			now := time.Now()
			apiKeys := APIKeyList{
				ApiKeys: []APIKeySummary{
					{
						ID:               "api-key-1",
						CreatedAt:        now,
						PublicKey:        "pk_test_123",
						DisplaySecretKey: "sk_test_***123",
					},
					{
						ID:               "api-key-2",
						CreatedAt:        now,
						PublicKey:        "pk_test_456",
						DisplaySecretKey: "sk_test_***456",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(apiKeys)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewClient(cli)

	apiKeys, err := client.GetAPIKeys(context.Background(), "test-project-id")
	require.NoError(t, err)
	require.Len(t, apiKeys.ApiKeys, 2)
	require.Equal(t, "api-key-1", apiKeys.ApiKeys[0].ID)
	require.Equal(t, "pk_test_123", apiKeys.ApiKeys[0].PublicKey)
	require.Equal(t, "sk_test_***123", apiKeys.ApiKeys[0].DisplaySecretKey)
}

func TestProjectClient_GetApiKeys_MissingProjectID(t *testing.T) {
	cli := resty.New()
	client := NewClient(cli)
	_, err := client.GetAPIKeys(context.Background(), "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "'projectID' is required")
}

func TestProjectClient_CreateApiKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/projects/test-project-id/apiKeys", r.URL.Path)
			require.Equal(t, "POST", r.Method)

			var req CreateAPIKeyRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)

			note := "Test API Key"
			if req.Note != "" {
				note = req.Note
			}

			now := time.Now()
			apiKey := APIKeyResponse{
				ID:               "created-api-key-id",
				CreatedAt:        now,
				PublicKey:        "pk_test_new",
				SecretKey:        "sk_test_secret_new",
				DisplaySecretKey: "sk_test_***new",
				Note:             note,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(apiKey)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewClient(cli)

	note := "Test API Key"
	createReq := &CreateAPIKeyRequest{Note: note}

	apiKey, err := client.CreateAPIKey(context.Background(), "test-project-id", createReq)
	require.NoError(t, err)
	require.Equal(t, "created-api-key-id", apiKey.ID)
	require.Equal(t, "pk_test_new", apiKey.PublicKey)
	require.Equal(t, "sk_test_secret_new", apiKey.SecretKey)
	require.Equal(t, "sk_test_***new", apiKey.DisplaySecretKey)
	require.Equal(t, "Test API Key", apiKey.Note)
}

func TestProjectClient_CreateApiKey_MissingProjectID(t *testing.T) {
	cli := resty.New()
	client := NewClient(cli)
	createReq := &CreateAPIKeyRequest{}
	_, err := client.CreateAPIKey(context.Background(), "", createReq)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'projectID' is required")
}

func TestProjectClient_DeleteApiKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/projects/test-project-id/apiKeys/test-api-key-id", r.URL.Path)
			require.Equal(t, "DELETE", r.Method)

			deleteResponse := APIKeyDeletionResponse{Success: true}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(deleteResponse)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewClient(cli)

	deleteResponse, err := client.DeleteAPIKey(context.Background(), "test-project-id", "test-api-key-id")
	require.NoError(t, err)
	require.True(t, deleteResponse.Success)
}

func TestProjectClient_DeleteApiKey_MissingProjectID(t *testing.T) {
	cli := resty.New()
	client := NewClient(cli)
	_, err := client.DeleteAPIKey(context.Background(), "", "test-api-key-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "'projectID' is required")
}

func TestProjectClient_DeleteApiKey_MissingApiKeyID(t *testing.T) {
	cli := resty.New()
	client := NewClient(cli)
	_, err := client.DeleteAPIKey(context.Background(), "test-project-id", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "'apiKeyID' is required")
}
