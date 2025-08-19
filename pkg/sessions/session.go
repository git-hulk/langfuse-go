package sessions

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/git-hulk/langfuse-go/pkg/common"
	"github.com/git-hulk/langfuse-go/pkg/traces"

	"github.com/go-resty/resty/v2"
)

// Session represents a Langfuse session.
type Session struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	ProjectID   string    `json:"projectId"`
	Environment string    `json:"environment,omitempty"`
}

// SessionWithTraces represents a session with its associated traces.
type SessionWithTraces struct {
	Session
	Traces []traces.TraceEntry `json:"traces"`
}

// ListParams defines the query parameters for listing sessions.
type ListParams struct {
	Page          int
	Limit         int
	FromTimestamp time.Time
	ToTimestamp   time.Time
	Environment   []string
}

// ToQueryString converts the ListParams to a URL query string.
func (p *ListParams) ToQueryString() string {
	parts := make([]string, 0)

	if p.Page != 0 {
		parts = append(parts, "page="+strconv.Itoa(p.Page))
	}
	if p.Limit != 0 {
		parts = append(parts, "limit="+strconv.Itoa(p.Limit))
	}
	if !p.FromTimestamp.IsZero() {
		parts = append(parts, "fromTimestamp="+url.QueryEscape(p.FromTimestamp.Format(time.RFC3339)))
	}
	if !p.ToTimestamp.IsZero() {
		parts = append(parts, "toTimestamp="+url.QueryEscape(p.ToTimestamp.Format(time.RFC3339)))
	}
	if len(p.Environment) > 0 {
		for _, env := range p.Environment {
			if env != "" {
				parts = append(parts, "environment="+url.QueryEscape(env))
			}
		}
	}

	return strings.Join(parts, "&")
}

// ListSessions represents the response from listing sessions.
type ListSessions struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []Session           `json:"data"`
}

// Client represents the sessions API client.
type Client struct {
	restyCli *resty.Client
}

// NewClient creates a new sessions API client.
func NewClient(cli *resty.Client) *Client {
	return &Client{restyCli: cli}
}

// List retrieves a list of sessions based on the provided parameters.
func (c *Client) List(ctx context.Context, params ListParams) (*ListSessions, error) {
	var listResponse ListSessions
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&listResponse).
		SetQueryString(params.ToQueryString()).
		Get("/sessions")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("list sessions failed with status code %d", rsp.StatusCode())
	}
	return &listResponse, nil
}

// Get retrieves a specific session by ID with its traces.
func (c *Client) Get(ctx context.Context, sessionID string) (*SessionWithTraces, error) {
	if sessionID == "" {
		return nil, errors.New("'sessionID' is required")
	}

	var session SessionWithTraces
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&session).
		SetPathParam("sessionId", sessionID)

	rsp, err := req.Get("/sessions/{sessionId}")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("get session failed with status code %d", rsp.StatusCode())
	}
	return &session, nil
}
