package projects

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Project represents a Langfuse project.
type Project struct {
	ID            string                 `json:"id,omitempty"`
	Name          string                 `json:"name"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	RetentionDays *int                   `json:"retentionDays,omitempty"`
}

// CreateProjectRequest represents the request payload for creating a project.
type CreateProjectRequest struct {
	Name      string                 `json:"name"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Retention int                    `json:"retention"`
}

// UpdateProjectRequest represents the request payload for updating a project.
type UpdateProjectRequest struct {
	Name      string                 `json:"name"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Retention int                    `json:"retention"`
}

// ProjectDeletionResponse represents the response from deleting a project.
type ProjectDeletionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ProjectsResponse represents the response from listing projects.
type ProjectsResponse struct {
	Data []Project `json:"data"`
}

// APIKeySummary represents summary information about an API key.
type APIKeySummary struct {
	ID               string     `json:"id"`
	CreatedAt        time.Time  `json:"createdAt"`
	ExpiresAt        *time.Time `json:"expiresAt,omitempty"`
	LastUsedAt       *time.Time `json:"lastUsedAt,omitempty"`
	Note             *string    `json:"note,omitempty"`
	PublicKey        string     `json:"publicKey"`
	DisplaySecretKey string     `json:"displaySecretKey"`
}

// APIKeyList represents a list of API keys for a project.
type APIKeyList struct {
	ApiKeys []APIKeySummary `json:"apiKeys"`
}

// APIKeyResponse represents the response from creating an API key.
type APIKeyResponse struct {
	ID               string    `json:"id"`
	CreatedAt        time.Time `json:"createdAt"`
	PublicKey        string    `json:"publicKey"`
	SecretKey        string    `json:"secretKey"`
	DisplaySecretKey string    `json:"displaySecretKey"`
	Note             *string   `json:"note,omitempty"`
}

// CreateAPIKeyRequest represents the request payload for creating an API key.
type CreateAPIKeyRequest struct {
	Note *string `json:"note,omitempty"`
}

// APIKeyDeletionResponse represents the response from deleting an API key.
type APIKeyDeletionResponse struct {
	Success bool `json:"success"`
}

func (req *CreateProjectRequest) validate() error {
	if req.Name == "" {
		return errors.New("'name' is required")
	}
	return nil
}

func (req *UpdateProjectRequest) validate() error {
	if req.Name == "" {
		return errors.New("'name' is required")
	}
	return nil
}

// Client represents the projects API client.
type Client struct {
	restyCli *resty.Client
}

// NewClient creates a new projects API client.
func NewClient(cli *resty.Client) *Client {
	return &Client{restyCli: cli}
}

// Get retrieves the project associated with the API key.
func (c *Client) Get(ctx context.Context) (*ProjectsResponse, error) {
	var projects ProjectsResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&projects).
		Get("/projects")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("get projects failed with status code %d", rsp.StatusCode())
	}
	return &projects, nil
}

// Create creates a new project (requires organization-scoped API key).
func (c *Client) Create(ctx context.Context, createReq *CreateProjectRequest) (*Project, error) {
	if err := createReq.validate(); err != nil {
		return nil, err
	}

	var createdProject Project
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(createReq).
		SetResult(&createdProject).
		Post("/projects")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to create project: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &createdProject, nil
}

// Update updates a project by ID (requires organization-scoped API key).
func (c *Client) Update(ctx context.Context, projectID string, updateReq *UpdateProjectRequest) (*Project, error) {
	if projectID == "" {
		return nil, errors.New("'projectID' is required")
	}
	if err := updateReq.validate(); err != nil {
		return nil, err
	}

	var updatedProject Project
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(updateReq).
		SetResult(&updatedProject).
		SetPathParam("projectID", projectID).
		Put("/projects/{projectID}")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to update project, got status code: %d", rsp.StatusCode())
	}
	return &updatedProject, nil
}

// Delete deletes a project by ID (requires organization-scoped API key).
// Project deletion is processed asynchronously.
func (c *Client) Delete(ctx context.Context, projectID string) (*ProjectDeletionResponse, error) {
	if projectID == "" {
		return nil, errors.New("'projectID' is required")
	}

	var deleteResponse ProjectDeletionResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&deleteResponse).
		SetPathParam("projectID", projectID).
		Delete("/projects/{projectID}")
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode() != http.StatusAccepted {
		return nil, fmt.Errorf("delete project failed with status code %d", rsp.StatusCode())
	}
	return &deleteResponse, nil
}

// GetAPIKeys retrieves all API keys for a project (requires organization-scoped API key).
func (c *Client) GetAPIKeys(ctx context.Context, projectID string) (*APIKeyList, error) {
	if projectID == "" {
		return nil, errors.New("'projectID' is required")
	}

	var apiKeys APIKeyList
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&apiKeys).
		SetPathParam("projectID", projectID).
		Get("/projects/{projectID}/apiKeys")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("get project API keys failed with status code %d", rsp.StatusCode())
	}
	return &apiKeys, nil
}

// CreateAPIKey creates a new API key for a project (requires organization-scoped API key).
func (c *Client) CreateAPIKey(ctx context.Context, projectID string, createReq *CreateAPIKeyRequest) (*APIKeyResponse, error) {
	if projectID == "" {
		return nil, errors.New("'projectID' is required")
	}

	var createdAPIKey APIKeyResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(createReq).
		SetResult(&createdAPIKey).
		SetPathParam("projectID", projectID).
		Post("/projects/{projectID}/apiKeys")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to create API key: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &createdAPIKey, nil
}

// DeleteAPIKey deletes an API key for a project (requires organization-scoped API key).
func (c *Client) DeleteAPIKey(ctx context.Context, projectID, apiKeyID string) (*APIKeyDeletionResponse, error) {
	if projectID == "" {
		return nil, errors.New("'projectID' is required")
	}
	if apiKeyID == "" {
		return nil, errors.New("'apiKeyID' is required")
	}

	var deleteResponse APIKeyDeletionResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&deleteResponse).
		SetPathParam("projectID", projectID).
		SetPathParam("apiKeyID", apiKeyID).
		Delete("/projects/{projectID}/apiKeys/{apiKeyID}")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("delete API key failed with status code %d", rsp.StatusCode())
	}
	return &deleteResponse, nil
}
