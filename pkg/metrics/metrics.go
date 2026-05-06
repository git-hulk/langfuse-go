// Package metrics provides access to the Langfuse public metrics API.
//
// The metrics API allows callers to run analytics queries across traces,
// observations, and scores by submitting a structured query object as a
// URL-encoded JSON string.
package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// View represents the metrics dataset to query.
type View string

const (
	ViewTraces            View = "traces"
	ViewObservations      View = "observations"
	ViewScoresNumeric     View = "scores-numeric"
	ViewScoresCategorical View = "scores-categorical"
)

// Granularity represents the time bucket size used when grouping metrics.
type Granularity string

const (
	GranularityMinute Granularity = "minute"
	GranularityHour   Granularity = "hour"
	GranularityDay    Granularity = "day"
	GranularityWeek   Granularity = "week"
	GranularityMonth  Granularity = "month"
	GranularityAuto   Granularity = "auto"
)

// OrderDirection represents the sort direction for result rows.
type OrderDirection string

const (
	OrderDirectionAsc  OrderDirection = "asc"
	OrderDirectionDesc OrderDirection = "desc"
)

// Metric measure constants for use in Query.Metrics.
const (
	MeasureCount             = "count"
	MeasureObservationsCount = "observationsCount"
	MeasureScoresCount       = "scoresCount"
	MeasureInputTokens       = "inputTokens"
	MeasureLatency           = "latency"
	MeasureOutputTokens      = "outputTokens"
	MeasureTotalTokens       = "totalTokens"
	MeasureTotalCost         = "totalCost"
	MeasureTimeToFirstToken  = "timeToFirstToken"
	MeasureCountScores       = "countScores"
	MeasureValue             = "value"
)

var validViews = map[View]struct{}{
	ViewTraces:            {},
	ViewObservations:      {},
	ViewScoresNumeric:     {},
	ViewScoresCategorical: {},
}

var validGranularities = map[Granularity]struct{}{
	GranularityMinute: {},
	GranularityHour:   {},
	GranularityDay:    {},
	GranularityWeek:   {},
	GranularityMonth:  {},
	GranularityAuto:   {},
}

var validOrderDirections = map[OrderDirection]struct{}{
	OrderDirectionAsc:  {},
	OrderDirectionDesc: {},
}

var validMeasuresByView = map[View]map[string]struct{}{
	ViewTraces: {
		MeasureCount:             {},
		MeasureObservationsCount: {},
		MeasureScoresCount:       {},
		MeasureLatency:           {},
		MeasureTotalTokens:       {},
		MeasureTotalCost:         {},
	},
	ViewObservations: {
		MeasureCount:            {},
		MeasureInputTokens:      {},
		MeasureLatency:          {},
		MeasureOutputTokens:     {},
		MeasureTotalTokens:      {},
		MeasureTotalCost:        {},
		MeasureTimeToFirstToken: {},
		MeasureCountScores:      {},
	},
	ViewScoresNumeric: {
		MeasureCount: {},
		MeasureValue: {},
	},
	ViewScoresCategorical: {
		MeasureCount: {},
	},
}

// Dimension groups result rows by the provided field.
type Dimension struct {
	Field string `json:"field"`
}

// Metric selects which measure to aggregate and how to aggregate it.
type Metric struct {
	Measure     string `json:"measure"`
	Aggregation string `json:"aggregation"`
}

// Filter narrows the query results.
type Filter struct {
	Column   string `json:"column"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
	Type     string `json:"type,omitempty"`
	Key      string `json:"key,omitempty"`
}

// TimeDimension groups results into time buckets.
type TimeDimension struct {
	Granularity Granularity `json:"granularity"`
}

// OrderBy sorts the returned result rows.
type OrderBy struct {
	Field     string         `json:"field"`
	Direction OrderDirection `json:"direction"`
}

// QueryConfig controls optional query execution settings.
type QueryConfig struct {
	Bins     int `json:"bins,omitempty"`
	RowLimit int `json:"row_limit,omitempty"`
}

// Query represents a metrics query sent to GET /api/public/metrics.
type Query struct {
	View          View           `json:"view"`
	Dimensions    []Dimension    `json:"dimensions,omitempty"`
	Metrics       []Metric       `json:"metrics"`
	Filters       []Filter       `json:"filters,omitempty"`
	TimeDimension *TimeDimension `json:"timeDimension,omitempty"`
	FromTimestamp time.Time      `json:"fromTimestamp"`
	ToTimestamp   time.Time      `json:"toTimestamp"`
	OrderBy       []OrderBy      `json:"orderBy,omitempty"`
	Config        *QueryConfig   `json:"config,omitempty"`
}

func (q *Query) validate() error {
	if q == nil {
		return errors.New("'query' is required")
	}
	if q.View == "" {
		return errors.New("'view' is required")
	}
	if _, ok := validViews[q.View]; !ok {
		return fmt.Errorf("invalid 'view': %s", q.View)
	}
	if len(q.Metrics) == 0 {
		return errors.New("at least one 'metric' is required")
	}
	if q.FromTimestamp.IsZero() {
		return errors.New("'fromTimestamp' is required")
	}
	if q.ToTimestamp.IsZero() {
		return errors.New("'toTimestamp' is required")
	}
	if q.ToTimestamp.Before(q.FromTimestamp) {
		return errors.New("'toTimestamp' must be greater than or equal to 'fromTimestamp'")
	}

	for i, dimension := range q.Dimensions {
		if dimension.Field == "" {
			return fmt.Errorf("'dimensions[%d].field' is required", i)
		}
	}

	for i, metric := range q.Metrics {
		if metric.Measure == "" {
			return fmt.Errorf("'metrics[%d].measure' is required", i)
		}
		if _, ok := validMeasuresByView[q.View][metric.Measure]; !ok {
			return fmt.Errorf("invalid 'metrics[%d].measure' %q for view %q", i, metric.Measure, q.View)
		}
		if metric.Aggregation == "" {
			return fmt.Errorf("'metrics[%d].aggregation' is required", i)
		}
	}

	for i, filter := range q.Filters {
		if filter.Column == "" {
			return fmt.Errorf("'filters[%d].column' is required", i)
		}
		if filter.Operator == "" {
			return fmt.Errorf("'filters[%d].operator' is required", i)
		}
		if strings.EqualFold(filter.Column, "metadata") && filter.Key == "" {
			return fmt.Errorf("'filters[%d].key' is required when filtering on metadata", i)
		}
	}

	if q.TimeDimension != nil {
		if q.TimeDimension.Granularity == "" {
			return errors.New("'timeDimension.granularity' is required")
		}
		if _, ok := validGranularities[q.TimeDimension.Granularity]; !ok {
			return fmt.Errorf("invalid 'timeDimension.granularity': %s", q.TimeDimension.Granularity)
		}
	}

	for i, order := range q.OrderBy {
		if order.Field == "" {
			return fmt.Errorf("'orderBy[%d].field' is required", i)
		}
		if order.Direction == "" {
			return fmt.Errorf("'orderBy[%d].direction' is required", i)
		}
		if _, ok := validOrderDirections[order.Direction]; !ok {
			return fmt.Errorf("invalid 'orderBy[%d].direction': %s", i, order.Direction)
		}
	}

	if q.Config != nil {
		if q.Config.Bins != 0 && (q.Config.Bins < 1 || q.Config.Bins > 100) {
			return errors.New("'config.bins' must be between 1 and 100")
		}
		if q.Config.RowLimit != 0 && (q.Config.RowLimit < 1 || q.Config.RowLimit > 1000) {
			return errors.New("'config.row_limit' must be between 1 and 1000")
		}
	}

	return nil
}

// ToQueryString converts the metrics query to the API's URL-encoded query parameter format.
func (q *Query) ToQueryString() (string, error) {
	if err := q.validate(); err != nil {
		return "", err
	}

	payload, err := json.Marshal(q)
	if err != nil {
		return "", fmt.Errorf("marshal query failed: %w", err)
	}

	query := url.Values{}
	query.Add("query", string(payload))
	return query.Encode(), nil
}

// RawRow is a metrics response row with raw JSON preserved per field.
//
// Callers can either decode the entire row into a custom struct or retrieve
// individual fields as raw JSON.
type RawRow map[string]json.RawMessage

// Decode unmarshals the row into the provided destination.
func (r RawRow) Decode(dst any) error {
	if dst == nil {
		return errors.New("'dst' is required")
	}

	payload, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshal row failed: %w", err)
	}

	if err := json.Unmarshal(payload, dst); err != nil {
		return fmt.Errorf("decode row failed: %w", err)
	}
	return nil
}

// Raw returns the raw JSON value for a field, if present.
func (r RawRow) Raw(key string) (json.RawMessage, bool) {
	value, ok := r[key]
	return value, ok
}

// DecodeRow decodes a row into the requested Go type.
func DecodeRow[T any](row RawRow) (T, error) {
	var decoded T
	if err := row.Decode(&decoded); err != nil {
		return decoded, err
	}
	return decoded, nil
}

// DecodeRows decodes multiple rows into the requested Go type.
func DecodeRows[T any](rows []RawRow) ([]T, error) {
	decoded := make([]T, 0, len(rows))
	for i, row := range rows {
		item, err := DecodeRow[T](row)
		if err != nil {
			return nil, fmt.Errorf("decode row %d failed: %w", i, err)
		}
		decoded = append(decoded, item)
	}
	return decoded, nil
}

// MetricsResponse is the response from the public metrics API.
//
// Each item in Data preserves raw JSON per field because the row shape depends
// on the requested dimensions and aggregations.
type MetricsResponse struct {
	Data []RawRow `json:"data"`
}

// Client provides access to the public metrics endpoint.
type Client struct {
	restyCli *resty.Client
}

// NewClient creates a new metrics client.
func NewClient(cli *resty.Client) *Client {
	return &Client{restyCli: cli}
}

// Get runs a metrics query against the Langfuse public metrics API.
func (c *Client) Get(ctx context.Context, query *Query) (*MetricsResponse, error) {
	queryString, err := query.ToQueryString()
	if err != nil {
		return nil, err
	}

	var response MetricsResponse
	rsp, err := c.restyCli.R().
		SetContext(ctx).
		SetResult(&response).
		SetQueryString(queryString).
		Get("/metrics")
	if err != nil {
		return nil, err
	}

	if rsp.IsError() {
		return nil, fmt.Errorf("get metrics failed: %s, got status code: %d", rsp.String(), rsp.StatusCode())
	}

	return &response, nil
}
