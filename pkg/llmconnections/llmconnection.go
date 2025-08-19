// Package llmconnections provides functionality for managing LLM provider connections in Langfuse.
//
// This package allows you to configure connections to various LLM providers
// like OpenAI, Anthropic, Azure OpenAI, AWS Bedrock, and Google Vertex AI.
// Connections can be configured with custom models, base URLs, and additional headers.
package llmconnections

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-set/v3"

	"github.com/git-hulk/langfuse-go/pkg/common"

	"github.com/go-resty/resty/v2"
)

// LLMAdapter represents the type of LLM provider adapter.
//
// Each adapter corresponds to a specific LLM provider and defines
// how to interact with that provider's API endpoints.
type LLMAdapter string

const (
	AdapterAnthropic      LLMAdapter = "anthropic"
	AdapterOpenAI         LLMAdapter = "openai"
	AdapterAzure          LLMAdapter = "azure"
	AdapterBedrock        LLMAdapter = "bedrock"
	AdapterGoogleVertexAI LLMAdapter = "google-vertex-ai"
	AdapterGoogleAIStudio LLMAdapter = "google-ai-studio"
)

// LLMConnection represents a configured connection to an LLM provider.
//
// The connection contains provider-specific configuration including custom models,
// base URLs, and metadata. Sensitive information like API keys is excluded
// from this structure for security reasons.
type LLMConnection struct {
	ID                string     `json:"id"`
	Provider          string     `json:"provider"`
	Adapter           LLMAdapter `json:"adapter"`
	DisplaySecretKey  string     `json:"displaySecretKey"`
	BaseURL           string     `json:"baseURL,omitempty"`
	CustomModels      []string   `json:"customModels"`
	WithDefaultModels bool       `json:"withDefaultModels"`
	ExtraHeaderKeys   []string   `json:"extraHeaderKeys"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// UpsertLLMConnectionRequest represents the parameters for creating or updating an LLM connection.
//
// Provider, Adapter, and SecretKey are required fields. BaseURL is required for some adapters
// like Azure. CustomModels can specify additional models beyond the defaults.
// WithDefaultModels controls whether to include provider's default models.
type UpsertLLMConnectionRequest struct {
	Provider          string            `json:"provider"`
	Adapter           LLMAdapter        `json:"adapter"`
	SecretKey         string            `json:"secretKey"`
	BaseURL           string            `json:"baseURL,omitempty"`
	CustomModels      []string          `json:"customModels,omitempty"`
	WithDefaultModels *bool             `json:"withDefaultModels,omitempty"`
	ExtraHeaders      map[string]string `json:"extraHeaders,omitempty"`
}

func (r *UpsertLLMConnectionRequest) validate() error {
	if r.Provider == "" {
		return errors.New("'provider' is required")
	}
	if r.Adapter == "" {
		return errors.New("'adapter' is required")
	}
	if r.SecretKey == "" {
		return errors.New("'secretKey' is required")
	}

	validAdapters := set.From([]LLMAdapter{
		AdapterAnthropic, AdapterOpenAI, AdapterAzure,
		AdapterBedrock, AdapterGoogleVertexAI, AdapterGoogleAIStudio,
	})
	if !validAdapters.Contains(r.Adapter) {
		return fmt.Errorf("invalid 'adapter': %s, must be one of %v", r.Adapter, validAdapters)
	}

	return nil
}

// ListParams defines the query parameters for listing LLM connections.
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

// ListLLMConnections represents the paginated response from the list LLM connections API.
//
// It contains pagination metadata and an array of LLM connections matching the query criteria.
type ListLLMConnections struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []LLMConnection     `json:"data"`
}

// Client provides methods for interacting with the Langfuse LLM connections API.
//
// The client handles HTTP communication for LLM connection operations
// including creating, updating, and listing provider connections.
type Client struct {
	restyCli *resty.Client
}

// NewClient creates a new LLM connections client with the provided HTTP client.
//
// The resty client should be pre-configured with authentication and base URL.
func NewClient(cli *resty.Client) *Client {
	return &Client{restyCli: cli}
}

// List retrieves a list of LLM connections based on the provided parameters.
func (c *Client) List(ctx context.Context, params ListParams) (*ListLLMConnections, error) {
	var listResponse ListLLMConnections
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&listResponse).
		SetQueryString(params.ToQueryString()).
		Get("/llm-connections")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("list LLM connections failed with status code %d", rsp.StatusCode())
	}
	return &listResponse, nil
}

// Upsert creates or updates an LLM connection.
func (c *Client) Upsert(ctx context.Context, req *UpsertLLMConnectionRequest) (*LLMConnection, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	var connection LLMConnection
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&connection).
		Put("/llm-connections")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to upsert LLM connection, got status code: %d",
			rsp.StatusCode())
	}
	return &connection, nil
}
