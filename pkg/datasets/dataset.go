package datasets

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

// Dataset represents a Langfuse dataset.
type Dataset struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Metadata    any       `json:"metadata,omitempty"`
	ProjectID   string    `json:"projectId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateDatasetRequest represents the request body for creating a dataset.
type CreateDatasetRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Metadata    any    `json:"metadata,omitempty"`
}

func (r *CreateDatasetRequest) validate() error {
	if r.Name == "" {
		return errors.New("'name' is required")
	}
	return nil
}

// ListParams defines the query parameters for listing datasets.
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

// ListDatasets represents the response from listing datasets.
type ListDatasets struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []Dataset           `json:"data"`
}

// Client represents the datasets API client.
type Client struct {
	restyCli *resty.Client
}

// NewClient creates a new datasets API client.
func NewClient(cli *resty.Client) *Client {
	return &Client{restyCli: cli}
}

// V2 Datasets API methods

// Get retrieves a specific dataset by name.
func (c *Client) Get(ctx context.Context, datasetName string) (*Dataset, error) {
	if datasetName == "" {
		return nil, errors.New("'datasetName' is required")
	}

	var dataset Dataset
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&dataset).
		SetPathParam("datasetName", datasetName)

	rsp, err := req.Get("/v2/datasets/{datasetName}")
	if err != nil {
		return nil, err
	}
	if rsp.IsError() {
		return nil, fmt.Errorf("get dataset failed with status code %d", rsp.StatusCode())
	}
	return &dataset, nil
}

// List retrieves a list of datasets based on the provided parameters.
func (c *Client) List(ctx context.Context, params ListParams) (*ListDatasets, error) {
	var listResponse ListDatasets
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&listResponse).
		SetQueryString(params.ToQueryString()).
		Get("/v2/datasets")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("list datasets failed with status code %d", rsp.StatusCode())
	}
	return &listResponse, nil
}

// Create creates a new dataset.
func (c *Client) Create(ctx context.Context, createDataset *CreateDatasetRequest) (*Dataset, error) {
	if err := createDataset.validate(); err != nil {
		return nil, err
	}

	var createdDataset Dataset
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(createDataset).
		SetResult(&createdDataset).
		Post("/v2/datasets")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to create dataset, got status code: %d",
			rsp.StatusCode())
	}
	return &createdDataset, nil
}
