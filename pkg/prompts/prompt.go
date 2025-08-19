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

// ChatMessageWithPlaceHolder represents a chat message that can include placeholders.
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

// PromptEntry represents a Langfuse prompt with its configuration and messages.
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

// ListParams defines the query parameters for listing prompts.
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

// GetParams defines the parameters for retrieving a single prompt.
type GetParams struct {
	Name    string
	Label   string
	Version int
}

type ListPrompts struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     struct {
		Prompts []PromptEntry `json:"prompts"`
	} `json:"data"`
}

type Client struct {
	restyCli *resty.Client
}

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
		return nil, fmt.Errorf("get prompt failed with status code %d", rsp.StatusCode())
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
		return nil, fmt.Errorf("list prompts failed with status code %d", rsp.StatusCode())
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
		return nil, fmt.Errorf("failed to create prompt, got status code: %d", rsp.StatusCode())
	}
	return &createdPrompt, nil
}
