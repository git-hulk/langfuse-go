package scores

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

func TestCreateScoreConfigRequest_validate(t *testing.T) {
	tests := []struct {
		name    string
		request CreateScoreConfigRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid numeric config",
			request: CreateScoreConfigRequest{
				Name:     "accuracy",
				DataType: ScoreDataTypeNumeric,
				MinValue: 0.0,
				MaxValue: 1.0,
			},
			wantErr: false,
		},
		{
			name: "valid boolean config",
			request: CreateScoreConfigRequest{
				Name:     "is_correct",
				DataType: ScoreDataTypeBoolean,
			},
			wantErr: false,
		},
		{
			name: "valid categorical config",
			request: CreateScoreConfigRequest{
				Name:     "quality",
				DataType: ScoreDataTypeCategorical,
				Categories: []ConfigCategory{
					{Value: 1, Label: "Poor"},
					{Value: 2, Label: "Fair"},
					{Value: 3, Label: "Good"},
					{Value: 4, Label: "Excellent"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			request: CreateScoreConfigRequest{
				DataType: ScoreDataTypeNumeric,
			},
			wantErr: true,
			errMsg:  "'name' is required",
		},
		{
			name: "missing dataType",
			request: CreateScoreConfigRequest{
				Name: "accuracy",
			},
			wantErr: true,
			errMsg:  "'dataType' is required",
		},
		{
			name: "invalid dataType",
			request: CreateScoreConfigRequest{
				Name:     "accuracy",
				DataType: "INVALID",
			},
			wantErr: true,
			errMsg:  "invalid 'dataType': INVALID",
		},
		{
			name: "categorical without categories",
			request: CreateScoreConfigRequest{
				Name:     "quality",
				DataType: ScoreDataTypeCategorical,
			},
			wantErr: true,
			errMsg:  "'categories' is required for categorical score configs",
		},
		{
			name: "boolean with categories",
			request: CreateScoreConfigRequest{
				Name:     "is_correct",
				DataType: ScoreDataTypeBoolean,
				Categories: []ConfigCategory{
					{Value: 0, Label: "False"},
					{Value: 1, Label: "True"},
				},
			},
			wantErr: true,
			errMsg:  "'categories' cannot be set for boolean score configs",
		},
		{
			name: "category with empty label",
			request: CreateScoreConfigRequest{
				Name:     "quality",
				DataType: ScoreDataTypeCategorical,
				Categories: []ConfigCategory{
					{Value: 1, Label: "Poor"},
					{Value: 2, Label: ""},
				},
			},
			wantErr: true,
			errMsg:  "category[1].label is required",
		},
		{
			name: "invalid min/max values",
			request: CreateScoreConfigRequest{
				Name:     "accuracy",
				DataType: ScoreDataTypeNumeric,
				MinValue: 1.0,
				MaxValue: 0.0,
			},
			wantErr: true,
			errMsg:  "'minValue' must be less than 'maxValue'",
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

func TestConfigListParams_ToQueryString(t *testing.T) {
	tests := []struct {
		name   string
		params ConfigListParams
		want   string
	}{
		{
			name:   "empty params",
			params: ConfigListParams{},
			want:   "",
		},
		{
			name:   "page and limit only",
			params: ConfigListParams{Page: 2, Limit: 25},
			want:   "page=2&limit=25",
		},
		{
			name:   "page only",
			params: ConfigListParams{Page: 1},
			want:   "page=1",
		},
		{
			name:   "limit only",
			params: ConfigListParams{Limit: 10},
			want:   "limit=10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.params.ToQueryString()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestClient_CreateConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("successful create numeric config", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/score-configs", r.URL.Path)
			require.Equal(t, "POST", r.Method)

			var createReq CreateScoreConfigRequest
			err := json.NewDecoder(r.Body).Decode(&createReq)
			require.NoError(t, err)
			require.Equal(t, "accuracy", createReq.Name)
			require.Equal(t, ScoreDataTypeNumeric, createReq.DataType)
			require.Equal(t, 0.0, createReq.MinValue)
			require.Equal(t, 1.0, createReq.MaxValue)

			response := ScoreConfig{
				ID:          "config-123",
				Name:        createReq.Name,
				DataType:    createReq.DataType,
				MinValue:    createReq.MinValue,
				MaxValue:    createReq.MaxValue,
				CreatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
				UpdatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
				ProjectID:   "project-456",
				IsArchived:  false,
				Description: createReq.Description,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		scoreClient := NewClient(client)

		createReq := &CreateScoreConfigRequest{
			Name:        "accuracy",
			DataType:    ScoreDataTypeNumeric,
			MinValue:    0.0,
			MaxValue:    1.0,
			Description: "Accuracy score configuration",
		}

		result, err := scoreClient.CreateConfig(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "config-123", result.ID)
		require.Equal(t, "accuracy", result.Name)
		require.Equal(t, ScoreDataTypeNumeric, result.DataType)
		require.Equal(t, 0.0, result.MinValue)
		require.Equal(t, 1.0, result.MaxValue)
		require.False(t, result.IsArchived)
	})

	t.Run("create with validation error", func(t *testing.T) {
		client := resty.New()
		scoreClient := NewClient(client)

		createReq := &CreateScoreConfigRequest{} // Missing required fields
		result, err := scoreClient.CreateConfig(ctx, createReq)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "'name' is required")
	})
}

func TestClient_ListConfigs(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list configs", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/score-configs", r.URL.Path)
			require.Equal(t, "GET", r.Method)

			configs := ListScoreConfigs{
				Metadata: common.ListMetadata{
					Page:       1,
					Limit:      50,
					TotalItems: 2,
					TotalPages: 1,
				},
				Data: []ScoreConfig{
					{
						ID:         "config-1",
						Name:       "accuracy",
						DataType:   ScoreDataTypeNumeric,
						CreatedAt:  mustParseTime("2023-01-01T10:00:00Z"),
						UpdatedAt:  mustParseTime("2023-01-01T10:00:00Z"),
						ProjectID:  "project-456",
						IsArchived: false,
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(configs)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		scoreClient := NewClient(client)

		result, err := scoreClient.ListConfigs(ctx, ConfigListParams{})
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 1, len(result.Data))
		require.Equal(t, 2, result.Metadata.TotalItems)
	})
}

func TestClient_GetConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get config", func(t *testing.T) {
		configID := "config-123"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/score-configs/"+configID, r.URL.Path)
			require.Equal(t, "GET", r.Method)

			config := ScoreConfig{
				ID:          configID,
				Name:        "quality",
				DataType:    ScoreDataTypeCategorical,
				CreatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
				UpdatedAt:   mustParseTime("2023-01-01T10:00:00Z"),
				ProjectID:   "project-456",
				IsArchived:  false,
				Description: "Quality assessment configuration",
				Categories: []ConfigCategory{
					{Value: 1, Label: "Poor"},
					{Value: 2, Label: "Good"},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(config)
			require.NoError(t, err)
		}))
		defer server.Close()

		client := resty.New().SetBaseURL(server.URL)
		scoreClient := NewClient(client)

		result, err := scoreClient.GetConfig(ctx, configID)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, configID, result.ID)
		require.Equal(t, "quality", result.Name)
		require.Equal(t, ScoreDataTypeCategorical, result.DataType)
		require.Len(t, result.Categories, 2)
	})

	t.Run("get with empty config ID", func(t *testing.T) {
		client := resty.New()
		scoreClient := NewClient(client)

		result, err := scoreClient.GetConfig(ctx, "")
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "'configID' is required", err.Error())
	})
}
