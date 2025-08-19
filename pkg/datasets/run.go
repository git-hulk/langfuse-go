package datasets

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/git-hulk/langfuse-go/pkg/common"
)

// DatasetRun represents a dataset run.
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

// DatasetRunItem represents an item in a dataset run.
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

// DatasetRunWithItems represents a dataset run with its associated items.
type DatasetRunWithItems struct {
	DatasetRun
	DatasetRunItems []DatasetRunItem `json:"datasetRunItems"`
}

// ListDatasetRuns represents the response from listing dataset runs.
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
