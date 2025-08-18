package llmconnections

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-set/v3"

	"github.com/git-hulk/langfuse-go/pkg/common"

	"github.com/go-resty/resty/v2"
)

// LLMAdapter represents the supported LLM adapters.
type LLMAdapter string

const (
	AdapterAnthropic      LLMAdapter = "anthropic"
	AdapterOpenAI         LLMAdapter = "openai"
	AdapterAzure          LLMAdapter = "azure"
	AdapterBedrock        LLMAdapter = "bedrock"
	AdapterGoogleVertexAI LLMAdapter = "google-vertex-ai"
	AdapterGoogleAIStudio LLMAdapter = "google-ai-studio"
)

// LLMConnection represents an LLM API connection configuration (secrets excluded).
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

// UpsertLLMConnectionRequest represents a request to create or update an LLM connection.
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

// ListLLMConnections represents the response from listing LLM connections.
type ListLLMConnections struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []LLMConnection     `json:"data"`
}

// Client represents the LLM connections API client.
type Client struct {
	restyCli *resty.Client
}

// NewClient creates a new LLM connections API client.
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

	if rsp.StatusCode() != http.StatusOK {
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

	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to upsert LLM connection: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &connection, nil
}
