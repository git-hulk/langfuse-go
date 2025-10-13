package datasets

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/git-hulk/langfuse-go/pkg/common"
)

// DatasetRun represents an execution run against a dataset.
//
// A dataset run tracks the evaluation or processing of dataset items
// in a specific experiment or evaluation session. It contains metadata
// about the run and links to the associated dataset.
type DatasetRun struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Metadata    any       `json:"metadata,omitempty"`
	DatasetID   string    `json:"datasetId"`
	DatasetName string    `json:"datasetName"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// DatasetRunItem represents a single item within a dataset run.
//
// Each run item links a specific dataset item to a trace that was generated
// during the run, enabling tracking of how each dataset item was processed.
type DatasetRunItem struct {
	ID             string    `json:"id"`
	DatasetRunID   string    `json:"datasetRunId"`
	DatasetRunName string    `json:"datasetRunName"`
	DatasetItemID  string    `json:"datasetItemId"`
	TraceID        string    `json:"traceId"`
	ObservationID  string    `json:"observationId,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// DatasetRunWithItems represents a complete dataset run including all its items.
//
// This structure embeds the DatasetRun and includes an array of all
// DatasetRunItems that were processed during this run.
type DatasetRunWithItems struct {
	DatasetRun
	DatasetRunItems []DatasetRunItem `json:"datasetRunItems"`
}

// ListDatasetRuns represents the paginated response from the list dataset runs API.
//
// It contains pagination metadata and an array of dataset runs matching the query criteria.
type ListDatasetRuns struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []DatasetRun        `json:"data"`
}

// DeleteDatasetRunResponse represents the response from deleting a dataset run.
type DeleteDatasetRunResponse struct {
	Message string `json:"message"`
}

// CreateDatasetRunItemRequest represents the request body for creating a dataset run item.
type CreateDatasetRunItemRequest struct {
	RunName string `json:"runName"`
	// Description of the run. If run exists, description will be updated.
	RunDescription string `json:"runDescription"`
	DatasetItemID  string `json:"datasetItemId,omitempty"`
	// Metadata of the dataset run, updates run if run already exists
	Metadata      any    `json:"metadata,omitempty"`
	ObservationID string `json:"observationId,omitempty"`
	// traceId should always be provided. For compatibility with older SDK versions it can also be inferred from the provided observationId.
	TraceID string `json:"traceId"`
}

func (r *CreateDatasetRunItemRequest) validate() error {
	if r.RunName == "" {
		return errors.New("'runName' is required")
	}
	if r.TraceID == "" {
		return errors.New("'traceId' is required")
	}
	return nil
}

// ListDatasetRunItemsParams represents the paginated response from the list dataset runs API.
type ListDatasetRunItemsParams struct {
	DatasetID string `json:"datasetId"`
	RunName   string `json:"runName"`
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
}

// ToQueryString converts the ListDatasetRunItemsParams to a URL query string.
func (p *ListDatasetRunItemsParams) ToQueryString() string {
	parts := url.Values{}
	if p.Page != 0 {
		parts.Add("page", strconv.Itoa(p.Page))
	}
	if p.Limit != 0 {
		parts.Add("limit", strconv.Itoa(p.Limit))
	}
	if p.RunName != "" {
		parts.Add("runName", p.RunName)
	}
	if p.DatasetID != "" {
		parts.Add("datasetId", p.DatasetID)
	}
	return parts.Encode()
}

// ListDatasetRunItems represents the paginated response from the list dataset runs API.
//
// It contains pagination metadata and an array of datasets matching the query criteria.
type ListDatasetRunItems struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []DatasetRunItem    `json:"data"`
}

// GetDatasetRuns retrieves runs for a specific dataset.
func (c *Client) GetDatasetRuns(ctx context.Context, datasetName string, params ListParams) (*ListDatasetRuns, error) {
	if datasetName == "" {
		return nil, errors.New("'datasetName' is required")
	}

	var listResponse ListDatasetRuns
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&listResponse).
		SetPathParam("datasetName", datasetName).
		SetQueryString(params.ToQueryString())

	rsp, err := req.Get("/datasets/{datasetName}/runs")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("get dataset runs failed: %s, got status code: %d", rsp.String(), rsp.StatusCode())
	}
	return &listResponse, nil
}

// GetDatasetRun retrieves a specific dataset run and its items.
func (c *Client) GetDatasetRun(ctx context.Context, datasetName, runName string) (*DatasetRunWithItems, error) {
	if datasetName == "" {
		return nil, errors.New("'datasetName' is required")
	}
	if runName == "" {
		return nil, errors.New("'runName' is required")
	}

	var datasetRun DatasetRunWithItems
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&datasetRun).
		SetPathParam("datasetName", datasetName).
		SetPathParam("runName", runName)

	rsp, err := req.Get("/datasets/{datasetName}/runs/{runName}")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("get dataset run failed: %s, got status code: %d", rsp.String(), rsp.StatusCode())
	}
	return &datasetRun, nil
}

// DeleteDatasetRun deletes a dataset run and all its run items.
func (c *Client) DeleteDatasetRun(ctx context.Context, datasetName, runName string) (*DeleteDatasetRunResponse, error) {
	if datasetName == "" {
		return nil, errors.New("'datasetName' is required")
	}
	if runName == "" {
		return nil, errors.New("'runName' is required")
	}

	var deleteResponse DeleteDatasetRunResponse
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&deleteResponse).
		SetPathParam("datasetName", datasetName).
		SetPathParam("runName", runName)

	rsp, err := req.Delete("/datasets/{datasetName}/runs/{runName}")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("delete dataset run failed: %s, got status code: %d", rsp.String(), rsp.StatusCode())
	}
	return &deleteResponse, nil
}

// CreateDatasetRunItems create a dataset run and current run items.
func (c *Client) CreateDatasetRunItems(ctx context.Context, req CreateDatasetRunItemRequest) (*DatasetRunItem, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}
	var resp DatasetRunItem
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&resp).
		Post("/dataset-run-items")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to dataset run items: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &resp, nil
}

// ListDatasetRunItems retrieves a list of dataset run items.
func (c *Client) ListDatasetRunItems(ctx context.Context, params ListDatasetRunItemsParams) (*ListDatasetRunItems, error) {
	if params.DatasetID == "" {
		return nil, errors.New("'datasetId' is required")
	}
	if params.RunName == "" {
		return nil, errors.New("'runName' is required")
	}
	var listResponse ListDatasetRunItems
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&listResponse).
		SetQueryString(params.ToQueryString())

	rsp, err := req.Get("/dataset-run-items")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("list dataset run items failed: %s, got status code: %d", rsp.String(), rsp.StatusCode())
	}
	return &listResponse, nil
}
