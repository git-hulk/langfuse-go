package datasets

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

// DatasetItem represents a Langfuse dataset item.
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

// CreateDatasetItemRequest represents the request to create a dataset item.
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

// UpdateDatasetItemRequest represents the request to update a dataset item.
type UpdateDatasetItemRequest struct {
	Input               any    `json:"input,omitempty"`
	ExpectedOutput      any    `json:"expectedOutput,omitempty"`
	Metadata            any    `json:"metadata,omitempty"`
	SourceTraceID       string `json:"sourceTraceId,omitempty"`
	SourceObservationID string `json:"sourceObservationId,omitempty"`
	Status              string `json:"status,omitempty"`
}

// ListParams defines the query parameters for listing dataset items.
type ListParams struct {
	DatasetName         string
	SourceTraceID       string
	SourceObservationID string
	Page                int
	Limit               int
}

// ToQueryString converts the ListParams to a URL query string.
func (query *ListParams) ToQueryString() string {
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

// ListDatasetItems represents the response from listing dataset items.
type ListDatasetItems struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []DatasetItem       `json:"data"`
}

// Client represents the dataset items API client.
type Client struct {
	restyCli *resty.Client
}

// NewClient creates a new dataset items API client.
func NewClient(cli *resty.Client) *Client {
	return &Client{restyCli: cli}
}

// Get retrieves a specific dataset item by ID.
func (c *Client) Get(ctx context.Context, id string) (*DatasetItem, error) {
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
	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get dataset item failed with status code %d", rsp.StatusCode())
	}
	return &datasetItem, nil
}

// List retrieves a list of dataset items based on the provided parameters.
func (c *Client) List(ctx context.Context, params ListParams) (*ListDatasetItems, error) {
	var listResponse ListDatasetItems
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&listResponse).
		SetQueryString(params.ToQueryString()).
		Get("/dataset-items")
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("list dataset items failed with status code %d", rsp.StatusCode())
	}
	return &listResponse, nil
}

// Create creates a new dataset item.
func (c *Client) Create(ctx context.Context, createDatasetItem *CreateDatasetItemRequest) (*DatasetItem, error) {
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

	if rsp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to create dataset item: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &createdDatasetItem, nil
}

// Delete deletes a dataset item by ID.
func (c *Client) Delete(ctx context.Context, id string) error {
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

	if rsp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete dataset item with status code %d", rsp.StatusCode())
	}
	return nil
}
