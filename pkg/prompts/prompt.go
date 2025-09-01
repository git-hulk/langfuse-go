// Package prompts provides functionality for managing prompt templates and versions in Langfuse.
//
// This package allows you to create, retrieve, list, and manage prompt templates
// for your AI applications. Prompts can contain placeholders and support versioning
// for iterative development and A/B testing of prompt variations.
package prompts

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

// ChatMessageWithPlaceHolder represents a chat message that can include placeholders for dynamic content.
//
// Placeholders in the content can be replaced with actual values when using the prompt.
// The Role field specifies the message role (e.g., "system", "user", "assistant"),
// Type specifies the content type, and Content contains the message text with optional placeholders.
type ChatMessageWithPlaceHolder struct {
	Role    string `json:"role"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func (c *ChatMessageWithPlaceHolder) validate() error {
	if c.Role == "" {
		return errors.New("'role' is required")
	}
	if c.Content == "" {
		return errors.New("'content' is required")
	}
	return nil
}

// PromptEntry represents a complete prompt template with its configuration and messages.
//
// A prompt entry contains the prompt name, an array of chat messages with placeholders,
// type information, version number, and optional metadata like tags and labels.
// The Config field can contain model-specific configuration parameters.
type PromptEntry struct {
	Name    string                       `json:"name"`
	Prompt  []ChatMessageWithPlaceHolder `json:"prompt"`
	Type    string                       `json:"type"`
	Version int                          `json:"version,omitempty"`
	Tags    []string                     `json:"tags,omitempty"`
	Labels  []string                     `json:"labels,omitempty"`
	Config  any                          `json:"config,omitempty"`
}

func (p *PromptEntry) validate() error {
	if p.Name == "" {
		return errors.New("'name' is required")
	}
	if len(p.Prompt) == 0 {
		return errors.New("'prompts' cannot be empty")
	}
	for _, msg := range p.Prompt {
		if err := msg.validate(); err != nil {
			return fmt.Errorf("invalid prompts message: %w", err)
		}
	}
	return nil
}

// ListParams defines the query parameters for filtering and paginating prompt listings.
//
// Use these parameters to filter prompts by name, labels, tags, and update timestamps,
// as well as to control pagination with Page and Limit fields.
type ListParams struct {
	Name          string
	Label         string
	Tag           string
	Page          int
	Limit         int
	FromUpdatedAt time.Time
	ToUpdatedAt   time.Time
}

// ToQueryString converts the ListParams to a URL query string.
func (query *ListParams) ToQueryString() string {
	parts := make([]string, 0)
	if query.Name != "" {
		parts = append(parts, "name="+query.Name+"")
	}
	if query.Label != "" {
		parts = append(parts, "label="+query.Label)
	}
	if query.Tag != "" {
		parts = append(parts, "tag="+query.Tag)
	}
	if query.Page != 0 {
		parts = append(parts, "page="+strconv.Itoa(query.Page))
	}
	if query.Limit != 0 {
		parts = append(parts, "limit="+strconv.Itoa(query.Limit))
	}
	if !query.FromUpdatedAt.IsZero() {
		// format with ios8601
		parts = append(parts, "from_updated_at="+query.FromUpdatedAt.Format(time.RFC3339))
	}
	if !query.ToUpdatedAt.IsZero() {
		parts = append(parts, "to_updated_at="+query.ToUpdatedAt.Format(time.RFC3339))
	}
	return strings.Join(parts, "&")
}

// GetParams defines the parameters for retrieving a specific prompt.
//
// Use Name to specify the prompt name, Label for a specific label,
// and Version for a specific version. If Version is 0, the latest version is returned.
type GetParams struct {
	Name    string
	Label   string
	Version int
}

// ListPrompts represents the response structure for prompt listing operations.
//
// It contains pagination metadata and an array of prompt entries matching the query parameters.
type ListPrompts struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []PromptEntry       `json:"data"`
}

// Client provides methods for interacting with the Langfuse prompts API.
//
// The client handles HTTP communication with the Langfuse API for prompt management
// operations including creating, retrieving, and listing prompt templates.
type Client struct {
	restyCli *resty.Client
}

// NewClient creates a new prompts client with the provided HTTP client.
//
// The resty client should be pre-configured with authentication and base URL.
func NewClient(cli *resty.Client) *Client {
	return &Client{restyCli: cli}
}

// Get retrieves a specific prompt by name, version, and label.
func (c *Client) Get(ctx context.Context, params GetParams) (*PromptEntry, error) {
	if params.Name == "" {
		return nil, errors.New("'name' is required")
	}

	var prompt PromptEntry
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&prompt)
	if params.Version > 0 {
		req.SetQueryParam("version", strconv.Itoa(params.Version))
	}
	if params.Label != "" {
		req.SetQueryParam("label", params.Label)
	}
	req.SetPathParam("name", params.Name)

	rsp, err := req.Get("/v2/prompts/{name}")
	if err != nil {
		return nil, err
	}
	if rsp.IsError() {
		return nil, fmt.Errorf("get prompt failed: %s, got status code: %d", rsp.String(), rsp.StatusCode())
	}
	return &prompt, nil
}

// List retrieves a list of prompts based on the provided parameters.
func (c Client) List(ctx context.Context, params ListParams) (*ListPrompts, error) {
	var listResponse ListPrompts
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&listResponse).
		SetQueryString(params.ToQueryString()).
		Get("/v2/prompts")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("list prompts failed: %s, got status code: %d", rsp.String(), rsp.StatusCode())
	}
	return &listResponse, nil
}

// Create creates a new prompt.
func (c *Client) Create(ctx context.Context, createPrompt *PromptEntry) (*PromptEntry, error) {
	if err := createPrompt.validate(); err != nil {
		return nil, err
	}

	// For reset the prompt version because it's not supported in the creating API.
	createPrompt.Version = 0

	var createdPrompt PromptEntry
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(createPrompt).
		SetResult(&createdPrompt).
		Post("/v2/prompts")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to create prompt: %s, got status code: %d", rsp.String(), rsp.StatusCode())
	}
	return &createdPrompt, nil
}
