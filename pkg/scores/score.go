// Package scores provides functionality for managing evaluation scores and score configurations in Langfuse.
//
// This package allows you to create, retrieve, and manage scores for your traces
// and observations, including score configurations for different data types.
// Scores can be numeric, boolean, or categorical and come from various sources
// including manual annotations, API calls, or automated evaluations.
package scores

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/git-hulk/langfuse-go/pkg/common"
	"github.com/git-hulk/langfuse-go/pkg/traces"

	"github.com/go-resty/resty/v2"
)

// ScoreSource represents the origin or source of a score.
//
// Scores can originate from different sources such as manual annotations,
// API calls, or automated evaluations, which helps track how scores were generated.
type ScoreSource string

const (
	ScoreSourceAnnotation ScoreSource = "ANNOTATION"
	ScoreSourceAPI        ScoreSource = "API"
	ScoreSourceEval       ScoreSource = "EVAL"
)

// ScoreDataType represents the data type and format of a score value.
//
// Scores can be numeric (float values), boolean (true/false), or categorical
// (predefined categories with associated values).
type ScoreDataType string

const (
	ScoreDataTypeNumeric     ScoreDataType = "NUMERIC"
	ScoreDataTypeBoolean     ScoreDataType = "BOOLEAN"
	ScoreDataTypeCategorical ScoreDataType = "CATEGORICAL"
)

// Score represents an evaluation score attached to a trace, observation, or session.
//
// Scores are used to evaluate the quality, performance, or other metrics of AI outputs.
// They can be attached to traces, observations, sessions, or dataset runs and include
// metadata about the source, author, and optional comments explaining the score.
type Score struct {
	DataType      ScoreDataType     `json:"dataType"`
	Value         float64           `json:"value"`
	ID            string            `json:"id"`
	TraceID       string            `json:"traceId,omitempty"`
	SessionID     string            `json:"sessionId,omitempty"`
	ObservationID string            `json:"observationId,omitempty"`
	DatasetRunID  string            `json:"datasetRunId,omitempty"`
	Name          string            `json:"name"`
	Source        ScoreSource       `json:"source"`
	Timestamp     time.Time         `json:"timestamp"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
	ConfigID      string            `json:"configId,omitempty"`
	Comment       string            `json:"comment,omitempty"`
	AuthorUserID  string            `json:"authorUserId,omitempty"`
	QueueID       string            `json:"queueId,omitempty"`
	Metadata      interface{}       `json:"metadata,omitempty"`
	Trace         traces.TraceEntry `json:"trace,omitempty"`
}

// CreateScoreRequest represents the parameters for creating a new score.
//
// At least one of TraceID, SessionID, or ObservationID must be provided to specify
// what the score is attached to. The Value field can be a float64 for numeric scores
// or a string for categorical/boolean scores.
type CreateScoreRequest struct {
	ID            string        `json:"id,omitempty"`
	TraceID       string        `json:"traceId,omitempty"`
	SessionID     string        `json:"sessionId,omitempty"`
	ObservationID string        `json:"observationId,omitempty"`
	DatasetRunID  string        `json:"datasetRunId,omitempty"`
	DataType      ScoreDataType `json:"dataType,omitempty"`
	Name          string        `json:"name"`
	Value         interface{}   `json:"value"` // Can be numeric (float64) or string
	Comment       string        `json:"comment,omitempty"`
	ConfigID      string        `json:"configId,omitempty"`
	Environment   string        `json:"environment,omitempty"`
	Metadata      any           `json:"metadata,omitempty"`
}

func (r *CreateScoreRequest) validate() error {
	if r.Name == "" {
		return errors.New("'name' is required")
	}
	if r.Value == nil {
		return errors.New("'value' is required")
	}
	// At least one of TraceID, SessionID, or ObservationID must be provided
	if r.TraceID == "" && r.SessionID == "" && r.ObservationID == "" {
		return errors.New("at least one of 'traceId', 'sessionId', or 'observationId' is required")
	}
	return nil
}

// CreateScoreResponse represents the response from creating a score.
//
// It contains the ID of the newly created score for reference.
type CreateScoreResponse struct {
	ID string `json:"id"`
}

// ListParams defines the query parameters for filtering and paginating score listings.
//
// Use Name to filter scores by name, UserID to filter by author, and timestamp fields
// to filter by creation time. Source and DataType can filter by score characteristics.
// Page and Limit control pagination.
type ListParams struct {
	Page          int
	Limit         int
	UserID        string
	Name          string
	FromTimestamp time.Time
	ToTimestamp   time.Time
	Environment   []string
	Source        ScoreSource
	Operator      string
	Value         float64
	ScoreIDs      []string
	ConfigID      string
	QueueID       string
	DataType      ScoreDataType
	TraceTags     []string
}

// ToQueryString converts the ListParams to a URL query string.
func (p *ListParams) ToQueryString() string {
	parts := make([]string, 0)

	if p.Page != 0 {
		parts = append(parts, "page="+strconv.Itoa(p.Page))
	}
	if p.Limit != 0 {
		parts = append(parts, "limit="+strconv.Itoa(p.Limit))
	}
	if p.UserID != "" {
		parts = append(parts, "userId="+url.QueryEscape(p.UserID))
	}
	if p.Name != "" {
		parts = append(parts, "name="+url.QueryEscape(p.Name))
	}
	if !p.FromTimestamp.IsZero() {
		parts = append(parts, "fromTimestamp="+url.QueryEscape(p.FromTimestamp.Format(time.RFC3339)))
	}
	if !p.ToTimestamp.IsZero() {
		parts = append(parts, "toTimestamp="+url.QueryEscape(p.ToTimestamp.Format(time.RFC3339)))
	}
	if len(p.Environment) > 0 {
		for _, env := range p.Environment {
			if env != "" {
				parts = append(parts, "environment="+url.QueryEscape(env))
			}
		}
	}
	if p.Source != "" {
		parts = append(parts, "source="+url.QueryEscape(string(p.Source)))
	}
	if p.Operator != "" {
		parts = append(parts, "operator="+url.QueryEscape(p.Operator))
	}
	if p.Value != 0 {
		parts = append(parts, "value="+strconv.FormatFloat(p.Value, 'f', -1, 64))
	}
	if len(p.ScoreIDs) > 0 {
		parts = append(parts, "scoreIds="+url.QueryEscape(strings.Join(p.ScoreIDs, ",")))
	}
	if p.ConfigID != "" {
		parts = append(parts, "configId="+url.QueryEscape(p.ConfigID))
	}
	if p.QueueID != "" {
		parts = append(parts, "queueId="+url.QueryEscape(p.QueueID))
	}
	if p.DataType != "" {
		parts = append(parts, "dataType="+url.QueryEscape(string(p.DataType)))
	}
	if len(p.TraceTags) > 0 {
		for _, tag := range p.TraceTags {
			if tag != "" {
				parts = append(parts, "traceTags="+url.QueryEscape(tag))
			}
		}
	}

	return strings.Join(parts, "&")
}

// ListScores represents the paginated response from the list scores API.
//
// It contains pagination metadata and an array of scores matching the query criteria.
type ListScores struct {
	Metadata common.ListMetadata `json:"meta"`
	Data     []Score             `json:"data"`
}

// Client provides methods for interacting with the Langfuse scores API.
//
// The client handles HTTP communication for score-related operations
// including creating, retrieving, listing, and deleting scores, as well as
// managing score configurations.
type Client struct {
	restyCli *resty.Client
}

// NewClient creates a new scores client with the provided HTTP client.
//
// The resty client should be pre-configured with authentication and base URL.
func NewClient(cli *resty.Client) *Client {
	return &Client{restyCli: cli}
}

// List retrieves a list of scores based on the provided parameters (v2 API).
func (c *Client) List(ctx context.Context, params ListParams) (*ListScores, error) {
	var listResponse ListScores
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&listResponse).
		SetQueryString(params.ToQueryString()).
		Get("/v2/scores")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("list scores failed with status code %d", rsp.StatusCode())
	}
	return &listResponse, nil
}

// Get retrieves a specific score by ID (v2 API).
func (c *Client) Get(ctx context.Context, scoreID string) (*Score, error) {
	if scoreID == "" {
		return nil, errors.New("'scoreID' is required")
	}

	var score Score
	req := c.restyCli.R().
		SetContext(ctx).
		SetResult(&score).
		SetPathParam("scoreID", scoreID)

	rsp, err := req.Get("/v2/scores/{scoreID}")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("get score failed with status code %d", rsp.StatusCode())
	}
	return &score, nil
}

// Create creates a new score (v1 API).
func (c *Client) Create(ctx context.Context, createScore *CreateScoreRequest) (*CreateScoreResponse, error) {
	if err := createScore.validate(); err != nil {
		return nil, err
	}

	var createdScore CreateScoreResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetBody(createScore).
		SetResult(&createdScore).
		Post("/scores")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("failed to create score: %s, got status code: %d",
			rsp.String(), rsp.StatusCode())
	}
	return &createdScore, nil
}

// Delete deletes a score by ID (v1 API).
func (c *Client) Delete(ctx context.Context, scoreID string) error {
	if scoreID == "" {
		return errors.New("'scoreID' is required")
	}

	req := c.restyCli.R().
		SetContext(ctx).
		SetPathParam("scoreID", scoreID)

	rsp, err := req.Delete("/scores/{scoreID}")
	if err != nil {
		return err
	}
	if rsp.IsError() {
		return fmt.Errorf("delete score failed with status code %d", rsp.StatusCode())
	}
	return nil
}
