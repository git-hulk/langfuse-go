// Package health provides functionality to check the Langfuse API health.
//
// This package exposes a simple client to retrieve the server health status
// and version as defined by the OpenAPI spec.
package health

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// HealthResponse represents the response structure for the health endpoint.
//
// It includes the server version and a status string (e.g., "OK").
type HealthResponse struct {
	Version string `json:"version"`
	Status  string `json:"status"`
}

// Client provides methods for interacting with the health endpoint.
type Client struct {
	restyCli *resty.Client
}

// NewClient creates a new health client with the provided HTTP client.
//
// The resty client should be pre-configured with authentication and base URL.
func NewClient(cli *resty.Client) *Client {
	return &Client{restyCli: cli}
}

// Check retrieves the API health status and version.
func (c *Client) Check(ctx context.Context) (*HealthResponse, error) {
	var health HealthResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&health).
		Get("/health")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("get health failed: %s, got status code: %d", rsp.String(), rsp.StatusCode())
	}
	return &health, nil
}
