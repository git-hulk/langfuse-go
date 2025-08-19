package datasets

import (
	"context"
	"errors"
	"fmt"
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
		return nil, fmt.Errorf("get dataset runs failed with status code %d", rsp.StatusCode())
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
		return nil, fmt.Errorf("get dataset run failed with status code %d", rsp.StatusCode())
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
		return nil, fmt.Errorf("delete dataset run failed with status code %d", rsp.StatusCode())
	}
	return &deleteResponse, nil
}
