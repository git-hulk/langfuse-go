package scores

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/git-hulk/langfuse-go/pkg/common"
)

// ConfigCategory represents a category configuration for categorical scores.
type ConfigCategory struct {
	Value float64 `json:"value"`
	Label string  `json:"label"`
}

// ScoreConfig represents a score configuration.
type ScoreConfig struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	ProjectID   string           `json:"projectId"`
	DataType    ScoreDataType    `json:"dataType"`
	IsArchived  bool             `json:"isArchived"`
	MinValue    float64          `json:"minValue,omitempty"`
	MaxValue    float64          `json:"maxValue,omitempty"`
	Categories  []ConfigCategory `json:"categories,omitempty"`
	Description string           `json:"description,omitempty"`
}

// CreateScoreConfigRequest represents the request body for creating a score config.
type CreateScoreConfigRequest struct {
	Name        string           `json:"name"`
	DataType    ScoreDataType    `json:"dataType"`
	Categories  []ConfigCategory `json:"categories,omitempty"`
	MinValue    float64          `json:"minValue,omitempty"`
	MaxValue    float64          `json:"maxValue,omitempty"`
	Description string           `json:"description,omitempty"`
}

func (r *CreateScoreConfigRequest) validate() error {
	if r.Name == "" {
		return errors.New("'name' is required")
	}
	if r.DataType == "" {
		return errors.New("'dataType' is required")
	}
	// Validate that dataType is a valid value
	validDataTypes := []ScoreDataType{ScoreDataTypeNumeric, ScoreDataTypeBoolean, ScoreDataTypeCategorical}
	isValid := false
	for _, dt := range validDataTypes {
		if r.DataType == dt {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid 'dataType': %s, must be one of NUMERIC, BOOLEAN, CATEGORICAL", r.DataType)
	}

	// Validate categories for categorical scores
	if r.DataType == ScoreDataTypeCategorical && len(r.Categories) == 0 {
		return errors.New("'categories' is required for categorical score configs")
	}
	if r.DataType == ScoreDataTypeBoolean && len(r.Categories) > 0 {
		return errors.New("'categories' cannot be set for boolean score configs")
	}

	// Validate category structure
	for i, category := range r.Categories {
		if category.Label == "" {
			return fmt.Errorf("category[%d].label is required", i)
		}
	}

	// Validate min/max values for numeric scores
	if r.MinValue != 0 || r.MaxValue != 0 {
		if r.MinValue >= r.MaxValue {
			return errors.New("'minValue' must be less than 'maxValue'")
		}
	}

	return nil
}

// ConfigListParams defines the query parameters for listing score configs.
type ConfigListParams struct {
	Page  int
	Limit int
}

// ToQueryString converts the ConfigListParams to a URL query string.
func (p *ConfigListParams) ToQueryString() string {
	parts := make([]string, 0)
	if p.Page != 0 {
		parts = append(parts, "page="+strconv.Itoa(p.Page))
	}
	if p.Limit != 0 {
		parts = append(parts, "limit="+strconv.Itoa(p.Limit))
	}
	return strings.Join(parts, "&")
}

// ListScoreConfigs represents the response from listing score configs.
type ListScoreConfigs struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []ScoreConfig       `json:"data"`
}

// CreateConfig creates a new score config.
func (c *Client) CreateConfig(ctx context.Context, createConfig *CreateScoreConfigRequest) (*ScoreConfig, error) {
	if err := createConfig.validate(); err != nil {
		return nil, err
	}

	var createdConfig ScoreConfig
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(createConfig).
		SetResult(&createdConfig).
		Post("/api/public/score-configs")
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create score config: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &createdConfig, nil
}

// ListConfigs retrieves a list of score configs based on the provided parameters.
func (c *Client) ListConfigs(ctx context.Context, params ConfigListParams) (*ListScoreConfigs, error) {
	var listResponse ListScoreConfigs
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&listResponse).
		SetQueryString(params.ToQueryString()).
		Get("/api/public/score-configs")
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("list score configs failed with status code %d", rsp.StatusCode())
	}
	return &listResponse, nil
}

// GetConfig retrieves a specific score config by ID.
func (c *Client) GetConfig(ctx context.Context, configID string) (*ScoreConfig, error) {
	if configID == "" {
		return nil, errors.New("'configID' is required")
	}

	var config ScoreConfig
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&config).
		SetPathParam("configId", configID)

	rsp, err := req.Get("/api/public/score-configs/{configId}")
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get score config failed with status code %d", rsp.StatusCode())
	}
	return &config, nil
}
