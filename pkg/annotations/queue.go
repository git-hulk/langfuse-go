package annotations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/git-hulk/langfuse-go/pkg/common"

	"github.com/go-resty/resty/v2"
)

// Queue represents an annotation queue.
type Queue struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	ScoreConfigIDs []string  `json:"scoreConfigIDs"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// CreateQueueRequest represents the request payload for creating an annotation queue.
type CreateQueueRequest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	ScoreConfigIDs []string `json:"scoreConfigIDs"`
}

func (r *CreateQueueRequest) validate() error {
	if r.Name == "" {
		return errors.New("'name' is required")
	}
	if len(r.ScoreConfigIDs) == 0 {
		return errors.New("'scoreConfigIds' is required and cannot be empty")
	}
	return nil
}

// QueueListParams defines the query parameters for listing annotation queues.
type QueueListParams struct {
	Page  int
	Limit int
}

// ToQueryString converts the QueueListParams to a URL query string.
func (query *QueueListParams) ToQueryString() string {
	parts := make([]string, 0)
	if query.Page != 0 {
		parts = append(parts, "page="+strconv.Itoa(query.Page))
	}
	if query.Limit != 0 {
		parts = append(parts, "limit="+strconv.Itoa(query.Limit))
	}
	return strings.Join(parts, "&")
}

// ListQueues represents the response from listing annotation queues.
type ListQueues struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []Queue             `json:"data"`
}

// AssignmentRequest represents the request for creating/deleting queue assignments.
type AssignmentRequest struct {
	UserID string `json:"userId"`
}

func (r *AssignmentRequest) validate() error {
	if r.UserID == "" {
		return errors.New("'userId' is required")
	}
	return nil
}

// CreateAssignmentResponse represents the response for creating a queue assignment.
type CreateAssignmentResponse struct {
	UserID    string `json:"userId"`
	QueueID   string `json:"queueId"`
	ProjectID string `json:"projectId"`
}

// DeleteAssignmentResponse represents the response for deleting a queue assignment.
type DeleteAssignmentResponse struct {
	Success bool `json:"success"`
}

// QueueClient represents the annotation queues API client.
type QueueClient struct {
	restyCli *resty.Client
}

// NewQueueClient creates a new annotation queues API client.
func NewQueueClient(cli *resty.Client) *QueueClient {
	return &QueueClient{restyCli: cli}
}

// Get retrieves a specific annotation queue by ID.
func (c *QueueClient) Get(ctx context.Context, queueID string) (*Queue, error) {
	if queueID == "" {
		return nil, errors.New("'queueID' is required")
	}

	var queue Queue
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&queue).
		SetPathParam("queueID", queueID)

	rsp, err := req.Get("/annotation-queues/{queueID}")
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get annotation queue failed with status code %d", rsp.StatusCode())
	}
	return &queue, nil
}

// List retrieves a list of annotation queues based on the provided parameters.
func (c *QueueClient) List(ctx context.Context, params QueueListParams) (*ListQueues, error) {
	var listResponse ListQueues
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&listResponse).
		SetQueryString(params.ToQueryString()).
		Get("/annotation-queues")
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("list annotation queues failed with status code %d", rsp.StatusCode())
	}
	return &listResponse, nil
}

// Create creates a new annotation queue.
func (c *QueueClient) Create(ctx context.Context, createRequest *CreateQueueRequest) (*Queue, error) {
	if err := createRequest.validate(); err != nil {
		return nil, err
	}

	var createdQueue Queue
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(createRequest).
		SetResult(&createdQueue).
		Post("/annotation-queues")
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create annotation queue: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &createdQueue, nil
}

// CreateAssignment creates an assignment for a user to an annotation queue.
func (c *QueueClient) CreateAssignment(ctx context.Context, queueID string, request *AssignmentRequest) (*CreateAssignmentResponse, error) {
	if queueID == "" {
		return nil, errors.New("'queueID' is required")
	}
	if err := request.validate(); err != nil {
		return nil, err
	}

	var assignmentResponse CreateAssignmentResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(request).
		SetResult(&assignmentResponse).
		SetPathParam("queueID", queueID).
		Post("/annotation-queues/{queueID}/assignments")
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create assignment: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &assignmentResponse, nil
}

// DeleteAssignment deletes an assignment for a user to an annotation queue.
func (c *QueueClient) DeleteAssignment(ctx context.Context, queueID string, request *AssignmentRequest) (*DeleteAssignmentResponse, error) {
	if queueID == "" {
		return nil, errors.New("'queueID' is required")
	}
	if err := request.validate(); err != nil {
		return nil, err
	}

	var deleteResponse DeleteAssignmentResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(request).
		SetResult(&deleteResponse).
		SetPathParam("queueID", queueID).
		Delete("/annotation-queues/{queueID}/assignments")
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to delete assignment: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &deleteResponse, nil
}
