package models

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestListParams_ToQueryString(t *testing.T) {
	tests := []struct {
		name   string
		params ListParams
		want   string
	}{
		{"with page and limit", ListParams{Page: 1, Limit: 10}, "page=1&limit=10"},
		{"with page only", ListParams{Page: 1}, "page=1"},
		{"with limit only", ListParams{Limit: 10}, "limit=10"},
		{"no params", ListParams{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.params.ToQueryString())
		})
	}
}

func TestModelEntry_validate(t *testing.T) {
	tests := []struct {
		name    string
		model   ModelEntry
		wantErr bool
		errMsg  string
	}{
		{
			"valid model",
			ModelEntry{MatchPattern: "gpt-4", ModelName: "gpt-4", Unit: "TOKENS"},
			false,
			"",
		},
		{
			"missing model name",
			ModelEntry{MatchPattern: "gpt-4", Unit: "TOKENS"},
			true,
			"'modelName' is required",
		},
		{
			"missing match pattern",
			ModelEntry{ModelName: "gpt-4", Unit: "TOKENS"},
			true,
			"'matchPattern' is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestModelClient_Get(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/models/test-model-id", r.URL.Path)
			model := ModelEntry{ID: "test-model-id", ModelName: "gpt-4", Unit: "TOKENS"}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(model)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewClient(cli)
	model, err := client.Get(context.Background(), "test-model-id")
	require.NoError(t, err)
	require.Equal(t, "test-model-id", model.ID)
	require.Equal(t, "gpt-4", model.ModelName)
	require.Equal(t, "TOKENS", model.Unit)
}

func TestModelClient_Get_MissingID(t *testing.T) {
	cli := resty.New()
	client := NewClient(cli)
	_, err := client.Get(context.Background(), "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "'id' is required")
}

func TestModelClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/models", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"meta":{"page":1,"limit":10,"totalItems":1,"totalPages":1},"data":[{"id":"test-model-id","modelName":"gpt-4","unit":"TOKENS"}]}`))
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewClient(cli)
	modelList, err := client.List(context.Background(), ListParams{})
	require.NoError(t, err)
	require.Len(t, modelList.Data, 1)
	require.Equal(t, "test-model-id", modelList.Data[0].ID)
	require.Equal(t, "gpt-4", modelList.Data[0].ModelName)
	require.Equal(t, "TOKENS", modelList.Data[0].Unit)
	// verify meta
	require.Equal(t, 1, modelList.Metadata.Page)
	require.Equal(t, 10, modelList.Metadata.Limit)
	require.Equal(t, 1, modelList.Metadata.TotalItems)
	require.Equal(t, 1, modelList.Metadata.TotalPages)
}

func TestModelClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "models", r.URL.Path)
			require.Equal(t, "POST", r.Method)
			var model ModelEntry
			err := json.NewDecoder(r.Body).Decode(&model)
			require.NoError(t, err)
			require.Equal(t, "gpt-4-custom", model.ModelName)
			require.Equal(t, "TOKENS", model.Unit)
			// Return the created model with an ID
			model.ID = "created-model-id"
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(model)
			require.NoError(t, err)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewClient(cli)
	createModel := &ModelEntry{
		MatchPattern: ".*gpt-4.*",
		ModelName:    "gpt-4-custom",
		Unit:         "TOKENS",
		InputPrice:   0.03,
		OutputPrice:  0.06,
	}
	model, err := client.Create(context.Background(), createModel)
	require.NoError(t, err)
	require.Equal(t, "created-model-id", model.ID)
	require.Equal(t, "gpt-4-custom", model.ModelName)
	require.Equal(t, "TOKENS", model.Unit)
}

func TestModelClient_Create_ValidationError(t *testing.T) {
	cli := resty.New()
	client := NewClient(cli)
	createModel := &ModelEntry{} // Missing required fields
	_, err := client.Create(context.Background(), createModel)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'modelName' is required")
}

func TestModelClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/models/test-model-id", r.URL.Path)
			require.Equal(t, "DELETE", r.Method)
			w.WriteHeader(http.StatusNoContent)
		}))
	defer server.Close()

	cli := resty.New().SetBaseURL(server.URL)
	client := NewClient(cli)
	err := client.Delete(context.Background(), "test-model-id")
	require.NoError(t, err)
}

func TestModelClient_Delete_MissingID(t *testing.T) {
	cli := resty.New()
	client := NewClient(cli)
	err := client.Delete(context.Background(), "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "'id' is required")
}
