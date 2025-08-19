package annotations

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

// QueueStatus represents the status of an annotation queue item.
type QueueStatus string

const (
	StatusPending   QueueStatus = "PENDING"
	StatusCompleted QueueStatus = "COMPLETED"
)

// QueueObjectType represents the type of object in an annotation queue.
type QueueObjectType string

const (
	ObjectTypeTrace       QueueObjectType = "TRACE"
	ObjectTypeObservation QueueObjectType = "OBSERVATION"
)

// Item represents an annotation queue item.
type Item struct {
	ID          string          `json:"id"`
	QueueID     string          `json:"queueId"`
	ObjectID    string          `json:"objectId"`
	ObjectType  QueueObjectType `json:"objectType"`
	Status      QueueStatus     `json:"status"`
	CompletedAt time.Time       `json:"completedAt,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// CreateItemRequest represents the request payload for creating an annotation queue item.
type CreateItemRequest struct {
	ObjectID   string          `json:"objectId"`
	ObjectType QueueObjectType `json:"objectType"`
	Status     QueueStatus     `json:"status,omitempty"`
}

func (r *CreateItemRequest) validate() error {
	if r.ObjectID == "" {
		return errors.New("'objectId' is required")
	}
	if r.ObjectType == "" {
		return errors.New("'objectType' is required")
	}
	if r.ObjectType != ObjectTypeTrace && r.ObjectType != ObjectTypeObservation {
		return fmt.Errorf("invalid 'objectType': %s, must be one of [TRACE, OBSERVATION]", r.ObjectType)
	}
	if r.Status != "" && r.Status != StatusPending && r.Status != StatusCompleted {
		return fmt.Errorf("invalid 'status': %s, must be one of [PENDING, COMPLETED]", r.Status)
	}
	return nil
}

// UpdateItemRequest represents the request payload for updating an annotation queue item.
type UpdateItemRequest struct {
	Status QueueStatus `json:"status,omitempty"`
}

func (r *UpdateItemRequest) validate() error {
	if r.Status != "" && r.Status != StatusPending && r.Status != StatusCompleted {
		return fmt.Errorf("invalid 'status': %s, must be one of [PENDING, COMPLETED]", r.Status)
	}
	return nil
}

// ItemListParams defines the query parameters for listing annotation queue items.
type ItemListParams struct {
	Status QueueStatus
	Page   int
	Limit  int
}

// ToQueryString converts the ItemListParams to a URL query string.
func (query *ItemListParams) ToQueryString() string {
	parts := make([]string, 0)
	if query.Status != "" {
		parts = append(parts, "status="+string(query.Status))
	}
	if query.Page != 0 {
		parts = append(parts, "page="+strconv.Itoa(query.Page))
	}
	if query.Limit != 0 {
		parts = append(parts, "limit="+strconv.Itoa(query.Limit))
	}
	return strings.Join(parts, "&")
}

// ListItems represents the response from listing annotation queue items.
type ListItems struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []Item              `json:"data"`
}

// DeleteItemResponse represents the response for deleting an annotation queue item.
type DeleteItemResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ItemClient represents the annotation queue items API client.
type ItemClient struct {
	restyCli *resty.Client
}

// NewItemClient creates a new annotation queue items API client.
func NewItemClient(cli *resty.Client) *ItemClient {
	return &ItemClient{restyCli: cli}
}

// Get retrieves a specific item from an annotation queue.
func (c *ItemClient) Get(ctx context.Context, queueID, itemID string) (*Item, error) {
	if queueID == "" {
		return nil, errors.New("'queueID' is required")
	}
	if itemID == "" {
		return nil, errors.New("'itemID' is required")
	}

	var item Item
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&item).
		SetPathParam("queueID", queueID).
		SetPathParam("itemID", itemID)

	rsp, err := req.Get("/annotation-queues/{queueID}/items/{itemID}")
	if err != nil {
		return nil, err
	}
	if rsp.IsError() {
		return nil, fmt.Errorf("get annotation queue item failed with status code %d", rsp.StatusCode())
	}
	return &item, nil
}

// List retrieves items for a specific annotation queue.
func (c *ItemClient) List(ctx context.Context, queueID string, params ItemListParams) (*ListItems, error) {
	if queueID == "" {
		return nil, errors.New("'queueID' is required")
	}

	var listResponse ListItems
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&listResponse).
		SetPathParam("queueID", queueID).
		SetQueryString(params.ToQueryString()).
		Get("/annotation-queues/{queueID}/items")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("list annotation queue items failed with status code %d", rsp.StatusCode())
	}
	return &listResponse, nil
}

// Create adds an item to an annotation queue.
func (c *ItemClient) Create(ctx context.Context, queueID string, createRequest *CreateItemRequest) (*Item, error) {
	if queueID == "" {
		return nil, errors.New("'queueID' is required")
	}
	if err := createRequest.validate(); err != nil {
		return nil, err
	}

	var createdItem Item
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(createRequest).
		SetResult(&createdItem).
		SetPathParam("queueID", queueID).
		Post("/annotation-queues/{queueID}/items")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to create annotation queue item: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &createdItem, nil
}

// Update updates an annotation queue item.
func (c *ItemClient) Update(ctx context.Context, queueID, itemID string, updateRequest *UpdateItemRequest) (*Item, error) {
	if queueID == "" {
		return nil, errors.New("'queueID' is required")
	}
	if itemID == "" {
		return nil, errors.New("'itemID' is required")
	}
	if err := updateRequest.validate(); err != nil {
		return nil, err
	}

	var updatedItem Item
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(updateRequest).
		SetResult(&updatedItem).
		SetPathParam("queueID", queueID).
		SetPathParam("itemID", itemID).
		Patch("/annotation-queues/{queueID}/items/{itemID}")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to update annotation queue item: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &updatedItem, nil
}

// Delete removes an item from an annotation queue.
func (c *ItemClient) Delete(ctx context.Context, queueID, itemID string) (*DeleteItemResponse, error) {
	if queueID == "" {
		return nil, errors.New("'queueID' is required")
	}
	if itemID == "" {
		return nil, errors.New("'itemID' is required")
	}

	var deleteResponse DeleteItemResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&deleteResponse).
		SetPathParam("queueID", queueID).
		SetPathParam("itemID", itemID).
		Delete("/annotation-queues/{queueID}/items/{itemID}")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to delete annotation queue item: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &deleteResponse, nil
}
