package datasets

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/git-hulk/langfuse-go/pkg/common"
)

// DatasetItem represents an individual item within a dataset.
//
// Dataset items contain input/output pairs for training or evaluation,
// along with metadata and optional links to source traces or observations.
// Items can have various statuses and are timestamped for audit purposes.
type DatasetItem struct {
	ID                  string    `json:"id,omitempty"`
	DatasetID           string    `json:"datasetId,omitempty"`
	DatasetName         string    `json:"datasetName,omitempty"`
	Input               any       `json:"input,omitempty"`
	ExpectedOutput      any       `json:"expectedOutput,omitempty"`
	Metadata            any       `json:"metadata,omitempty"`
	SourceTraceID       string    `json:"sourceTraceId,omitempty"`
	SourceObservationID string    `json:"sourceObservationId,omitempty"`
	Status              string    `json:"status,omitempty"`
	CreatedAt           time.Time `json:"createdAt,omitempty"`
	UpdatedAt           time.Time `json:"updatedAt,omitempty"`
}

// CreateDatasetItemRequest represents the parameters for creating a new dataset item.
//
// The DatasetName is required to specify which dataset the item belongs to.
// Input and ExpectedOutput contain the training/evaluation data pairs.
// Optional fields include metadata and source trace/observation references.
type CreateDatasetItemRequest struct {
	DatasetName         string `json:"datasetName"`
	Input               any    `json:"input,omitempty"`
	ExpectedOutput      any    `json:"expectedOutput,omitempty"`
	Metadata            any    `json:"metadata,omitempty"`
	SourceTraceID       string `json:"sourceTraceId,omitempty"`
	SourceObservationID string `json:"sourceObservationId,omitempty"`
	Status              string `json:"status,omitempty"`
	ID                  string `json:"id,omitempty"`
}

func (c *CreateDatasetItemRequest) validate() error {
	if c.DatasetName == "" {
		return errors.New("'datasetName' is required")
	}
	return nil
}

// UpdateDatasetItemRequest represents the parameters for updating an existing dataset item.
//
// All fields are optional and only non-zero values will be updated.
// Use this to modify input/output data, metadata, or source references.
type UpdateDatasetItemRequest struct {
	Input               any    `json:"input,omitempty"`
	ExpectedOutput      any    `json:"expectedOutput,omitempty"`
	Metadata            any    `json:"metadata,omitempty"`
	SourceTraceID       string `json:"sourceTraceId,omitempty"`
	SourceObservationID string `json:"sourceObservationId,omitempty"`
	Status              string `json:"status,omitempty"`
}

// ListDatasetItemParams defines the query parameters for filtering and paginating dataset items.
//
// Use DatasetName to specify which dataset to list items from (required).
// Optional filters include SourceTraceID and SourceObservationID for items linked to specific traces.
// Page and Limit control pagination.
type ListDatasetItemParams struct {
	DatasetName         string
	SourceTraceID       string
	SourceObservationID string
	Page                int
	Limit               int
}

// ToQueryString converts the ListDatasetItemParams to a URL query string.
func (query *ListDatasetItemParams) ToQueryString() string {
	parts := make([]string, 0)
	if query.DatasetName != "" {
		parts = append(parts, "datasetName="+query.DatasetName)
	}
	if query.SourceTraceID != "" {
		parts = append(parts, "sourceTraceId="+query.SourceTraceID)
	}
	if query.SourceObservationID != "" {
		parts = append(parts, "sourceObservationId="+query.SourceObservationID)
	}
	if query.Page != 0 {
		parts = append(parts, "page="+strconv.Itoa(query.Page))
	}
	if query.Limit != 0 {
		parts = append(parts, "limit="+strconv.Itoa(query.Limit))
	}
	return strings.Join(parts, "&")
}

// ListDatasetItems represents the paginated response from the list dataset items API.
//
// It contains pagination metadata and an array of dataset items matching the query criteria.
type ListDatasetItems struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []DatasetItem       `json:"data"`
}

// GetDatasetItem retrieves a specific dataset item by ID.
func (c *Client) GetDatasetItem(ctx context.Context, id string) (*DatasetItem, error) {
	if id == "" {
		return nil, errors.New("'id' is required")
	}

	var datasetItem DatasetItem
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&datasetItem).
		SetPathParam("id", id)

	rsp, err := req.Get("/dataset-items/{id}")
	if err != nil {
		return nil, err
	}
	if rsp.IsError() {
		return nil, fmt.Errorf("get dataset item failed with status code %d", rsp.StatusCode())
	}
	return &datasetItem, nil
}

// ListDatasetItems retrieves a list of dataset items based on the provided parameters.
func (c *Client) ListDatasetItems(ctx context.Context, params ListDatasetItemParams) (*ListDatasetItems, error) {
	var listResponse ListDatasetItems
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&listResponse).
		SetQueryString(params.ToQueryString()).
		Get("/dataset-items")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("list dataset items failed with status code %d", rsp.StatusCode())
	}
	return &listResponse, nil
}

// CreateDatasetItem creates a new dataset item.
func (c *Client) CreateDatasetItem(ctx context.Context, createDatasetItem *CreateDatasetItemRequest) (*DatasetItem, error) {
	if err := createDatasetItem.validate(); err != nil {
		return nil, err
	}

	var createdDatasetItem DatasetItem
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(createDatasetItem).
		SetResult(&createdDatasetItem).
		Post("/dataset-items")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to create dataset item, got status code: %d",
			rsp.StatusCode())
	}
	return &createdDatasetItem, nil
}

// DeleteDatasetItem deletes a dataset item by ID.
func (c *Client) DeleteDatasetItem(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("'id' is required")
	}

	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetPathParam("id", id).
		Delete("/dataset-items/{id}")
	if err != nil {
		return err
	}

	if rsp.IsError() {
		return fmt.Errorf("failed to delete dataset item with status code %d", rsp.StatusCode())
	}
	return nil
}
