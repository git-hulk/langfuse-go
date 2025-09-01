package llmconnections

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLLMConnectionValidation(t *testing.T) {
	tests := []struct {
		name      string
		req       UpsertLLMConnectionRequest
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid request",
			req: UpsertLLMConnectionRequest{
				Provider:  "openai",
				Adapter:   AdapterOpenAI,
				SecretKey: "sk-test123",
			},
			wantError: false,
		},
		{
			name: "missing provider",
			req: UpsertLLMConnectionRequest{
				Adapter:   AdapterOpenAI,
				SecretKey: "sk-test123",
			},
			wantError: true,
			errorMsg:  "'provider' is required",
		},
		{
			name: "missing adapter",
			req: UpsertLLMConnectionRequest{
				Provider:  "openai",
				SecretKey: "sk-test123",
			},
			wantError: true,
			errorMsg:  "'adapter' is required",
		},
		{
			name: "missing secret key",
			req: UpsertLLMConnectionRequest{
				Provider: "openai",
				Adapter:  AdapterOpenAI,
			},
			wantError: true,
			errorMsg:  "'secretKey' is required",
		},
		{
			name: "invalid adapter",
			req: UpsertLLMConnectionRequest{
				Provider:  "openai",
				Adapter:   "invalid",
				SecretKey: "sk-test123",
			},
			wantError: true,
			errorMsg:  "invalid 'adapter': invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.validate()
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestListParams_ToQueryString(t *testing.T) {
	tests := []struct {
		name     string
		params   ListParams
		expected string
	}{
		{
			name:     "empty params",
			params:   ListParams{},
			expected: "",
		},
		{
			name:     "page only",
			params:   ListParams{Page: 2},
			expected: "page=2",
		},
		{
			name:     "limit only",
			params:   ListParams{Limit: 50},
			expected: "limit=50",
		},
		{
			name:     "both page and limit",
			params:   ListParams{Page: 3, Limit: 25},
			expected: "page=3&limit=25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.ToQueryString()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClient_List(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/llm-connections", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "page=1")
		assert.Contains(t, r.URL.RawQuery, "limit=10")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"data": [
				{
					"id": "test-id",
					"provider": "openai",
					"adapter": "openai",
					"displaySecretKey": "sk-***123",
					"baseURL": "",
					"customModels": ["gpt-4"],
					"withDefaultModels": true,
					"extraHeaderKeys": ["X-Custom-Header"],
					"createdAt": "2023-01-01T00:00:00Z",
					"updatedAt": "2023-01-01T00:00:00Z"
				}
			],
			"meta": {
				"page": 1,
				"limit": 10,
				"totalItems": 1,
				"totalPages": 1
			}
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := &Client{
		restyCli: resty.New().SetBaseURL(server.URL),
	}

	params := ListParams{Page: 1, Limit: 10}
	result, err := client.List(context.Background(), params)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "test-id", result.Data[0].ID)
	assert.Equal(t, "openai", result.Data[0].Provider)
	assert.Equal(t, AdapterOpenAI, result.Data[0].Adapter)
	assert.Equal(t, "sk-***123", result.Data[0].DisplaySecretKey)
	assert.Equal(t, []string{"gpt-4"}, result.Data[0].CustomModels)
	assert.True(t, result.Data[0].WithDefaultModels)
	assert.Equal(t, []string{"X-Custom-Header"}, result.Data[0].ExtraHeaderKeys)
}

func TestClient_Upsert(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/llm-connections", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"id": "created-id",
			"provider": "openai",
			"adapter": "openai",
			"displaySecretKey": "sk-***123",
			"baseURL": "https://api.openai.com/v1",
			"customModels": ["gpt-4", "gpt-3.5-turbo"],
			"withDefaultModels": true,
			"extraHeaderKeys": [],
			"createdAt": "2023-01-01T00:00:00Z",
			"updatedAt": "2023-01-01T00:00:00Z"
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := &Client{
		restyCli: resty.New().SetBaseURL(server.URL),
	}

	withDefaultModels := true
	req := &UpsertLLMConnectionRequest{
		Provider:          "openai",
		Adapter:           AdapterOpenAI,
		SecretKey:         "sk-test123",
		BaseURL:           "https://api.openai.com/v1",
		CustomModels:      []string{"gpt-4", "gpt-3.5-turbo"},
		WithDefaultModels: withDefaultModels,
		ExtraHeaders:      map[string]string{},
	}

	result, err := client.Upsert(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "created-id", result.ID)
	assert.Equal(t, "openai", result.Provider)
	assert.Equal(t, AdapterOpenAI, result.Adapter)
	assert.Equal(t, "sk-***123", result.DisplaySecretKey)
	assert.Equal(t, "https://api.openai.com/v1", result.BaseURL)
	assert.Equal(t, []string{"gpt-4", "gpt-3.5-turbo"}, result.CustomModels)
	assert.True(t, result.WithDefaultModels)
	assert.Empty(t, result.ExtraHeaderKeys)
}

func TestClient_UpsertValidationError(t *testing.T) {
	client := &Client{
		restyCli: resty.New(),
	}

	req := &UpsertLLMConnectionRequest{
		// Missing required fields
	}

	result, err := client.Upsert(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "'provider' is required")
}

func TestClient_ListError(t *testing.T) {
	// Mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{
		restyCli: resty.New().SetBaseURL(server.URL),
	}

	params := ListParams{Page: 1, Limit: 10}
	result, err := client.List(context.Background(), params)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "list LLM connections failed with status code 500")
}

func TestClient_UpsertError(t *testing.T) {
	// Mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad request"))
	}))
	defer server.Close()

	client := &Client{
		restyCli: resty.New().SetBaseURL(server.URL),
	}

	req := &UpsertLLMConnectionRequest{
		Provider:  "openai",
		Adapter:   AdapterOpenAI,
		SecretKey: "sk-test123",
	}

	result, err := client.Upsert(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to upsert LLM connection")
	assert.Contains(t, err.Error(), "got status code: 400")
}
