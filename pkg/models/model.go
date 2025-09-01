// Package models provides functionality for managing model configurations and pricing in Langfuse.
//
// This package allows you to define model pricing, match patterns, and manage
// model metadata for cost tracking and analytics. Models can be configured with
// different pricing structures and tokenization settings.
package models

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/git-hulk/langfuse-go/pkg/common"

	"github.com/go-resty/resty/v2"
)

// TokenizerConfig represents configuration for tokenization behavior.
//
// This configuration defines how tokens are counted for specific operations
// like name processing and message handling, affecting cost calculations.
type TokenizerConfig struct {
	TokensPerName    int `json:"tokensPerName,omitempty"`
	TokensPerMessage int `json:"tokensPerMessage,omitempty"`
}

// ModelEntry represents a model configuration with pricing and metadata.
//
// A model entry defines how a model is identified (via name and match pattern),
// its pricing structure for input/output tokens, and tokenization configuration.
// Models are used for cost tracking and analytics in Langfuse.
type ModelEntry struct {
	ID              string          `json:"id,omitempty"`
	ModelName       string          `json:"modelName"`
	MatchPattern    string          `json:"matchPattern,omitempty"`
	StartDate       time.Time       `json:"startDate,omitempty"`
	InputPrice      float64         `json:"inputPrice,omitempty"`
	OutputPrice     float64         `json:"outputPrice,omitempty"`
	TotalPrice      float64         `json:"totalPrice,omitempty"`
	Unit            string          `json:"unit"`
	TokenizerId     string          `json:"tokenizerId,omitempty"`
	TokenizerConfig TokenizerConfig `json:"tokenizerConfig,omitempty"`
}

func (m *ModelEntry) validate() error {
	if m.ModelName == "" {
		return errors.New("'modelName' is required")
	}
	if m.MatchPattern == "" {
		return errors.New("'matchPattern' is required")
	}
	if m.Unit != "" && !common.ModelUsageUnits.Contains(m.Unit) {
		return fmt.Errorf("invalid 'unit': %s, must be one of %v", m.Unit, common.ModelUsageUnits.Slice())
	}
	return nil
}

// ListParams defines the query parameters for listing models.
type ListParams struct {
	Page  int
	Limit int
}

// ToQueryString converts the ListParams to a URL query string.
func (query *ListParams) ToQueryString() string {
	parts := make([]string, 0)
	if query.Page != 0 {
		parts = append(parts, "page="+strconv.Itoa(query.Page))
	}
	if query.Limit != 0 {
		parts = append(parts, "limit="+strconv.Itoa(query.Limit))
	}
	return strings.Join(parts, "&")
}

// ListModels represents the response from listing models.
type ListModels struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []ModelEntry        `json:"data"`
}

// Client represents the models API client.
type Client struct {
	restyCli *resty.Client
}

// NewClient creates a new models API client.
func NewClient(cli *resty.Client) *Client {
	return &Client{restyCli: cli}
}

// Get retrieves a specific model by ID.
func (c *Client) Get(ctx context.Context, id string) (*ModelEntry, error) {
	if id == "" {
		return nil, errors.New("'id' is required")
	}

	var model ModelEntry
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&model).
		SetPathParam("id", id)

	rsp, err := req.Get("/models/{id}")
	if err != nil {
		return nil, err
	}
	if rsp.IsError() {
		return nil, fmt.Errorf("get model failed: %s, got status code: %d", rsp.String(), rsp.StatusCode())
	}
	return &model, nil
}

// List retrieves a list of models based on the provided parameters.
func (c *Client) List(ctx context.Context, params ListParams) (*ListModels, error) {
	var listResponse ListModels
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&listResponse).
		SetQueryString(params.ToQueryString()).
		Get("/models")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("list models failed: %s, got status code: %d", rsp.String(), rsp.StatusCode())
	}
	return &listResponse, nil
}

// Create creates a new model.
func (c *Client) Create(ctx context.Context, createModel *ModelEntry) (*ModelEntry, error) {
	if err := createModel.validate(); err != nil {
		return nil, err
	}

	var createdModel ModelEntry
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(createModel).
		SetResult(&createdModel).
		Post("/models")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to create model: %s, got status code: %d", rsp.String(), rsp.StatusCode())
	}
	return &createdModel, nil
}

// Delete deletes a model by ID.
func (c *Client) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("'id' is required")
	}

	req := c.restyCli.R().
		SetContext(ctx).
		SetPathParam("id", id)

	rsp, err := req.Delete("/models/{id}")
	if err != nil {
		return err
	}
	if rsp.IsError() {
		return fmt.Errorf("delete model failed: %s, got status code: %d", rsp.String(), rsp.StatusCode())
	}
	return nil
}
