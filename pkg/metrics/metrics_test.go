package metrics

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestQuery_ToQueryString(t *testing.T) {
	query := &Query{
		View:       ViewTraces,
		Dimensions: []Dimension{{Field: "name"}},
		Metrics: []Metric{
			{Measure: MeasureCount, Aggregation: "count"},
		},
		Filters: []Filter{
			{Column: "name", Operator: "=", Value: "assistant"},
		},
		TimeDimension: &TimeDimension{Granularity: GranularityDay},
		FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
		ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
		OrderBy: []OrderBy{
			{Field: "count", Direction: OrderDirectionDesc},
		},
		Config: &QueryConfig{Bins: 10, RowLimit: 100},
	}

	queryString, err := query.ToQueryString()
	require.NoError(t, err)

	values, err := url.ParseQuery(queryString)
	require.NoError(t, err)
	require.Contains(t, values, "query")

	var decoded Query
	err = json.Unmarshal([]byte(values.Get("query")), &decoded)
	require.NoError(t, err)

	require.Equal(t, query.View, decoded.View)
	require.Equal(t, query.Dimensions, decoded.Dimensions)
	require.Equal(t, query.Metrics, decoded.Metrics)
	require.Equal(t, query.TimeDimension, decoded.TimeDimension)
	require.Equal(t, query.FromTimestamp, decoded.FromTimestamp)
	require.Equal(t, query.ToTimestamp, decoded.ToTimestamp)
	require.Equal(t, query.OrderBy, decoded.OrderBy)
	require.NotNil(t, decoded.Config)
	require.Equal(t, query.Config.RowLimit, decoded.Config.RowLimit)
	require.Equal(t, "assistant", decoded.Filters[0].Value)
}

func TestQuery_ToQueryStringValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		query   *Query
		wantErr string
	}{
		{
			name:    "nil query",
			query:   nil,
			wantErr: "'query' is required",
		},
		{
			name: "missing view",
			query: &Query{
				Metrics:       []Metric{{Measure: "count", Aggregation: "count"}},
				FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
				ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
			},
			wantErr: "'view' is required",
		},
		{
			name: "invalid metadata filter",
			query: &Query{
				View:          ViewTraces,
				Metrics:       []Metric{{Measure: MeasureCount, Aggregation: "count"}},
				Filters:       []Filter{{Column: "metadata", Operator: "=", Value: "x"}},
				FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
				ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
			},
			wantErr: "'filters[0].key' is required when filtering on metadata",
		},
		{
			name: "invalid time dimension",
			query: &Query{
				View:          ViewTraces,
				Metrics:       []Metric{{Measure: MeasureCount, Aggregation: "count"}},
				TimeDimension: &TimeDimension{Granularity: Granularity("year")},
				FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
				ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
			},
			wantErr: "invalid 'timeDimension.granularity'",
		},
		{
			name: "invalid order direction",
			query: &Query{
				View:          ViewTraces,
				Metrics:       []Metric{{Measure: MeasureCount, Aggregation: "count"}},
				OrderBy:       []OrderBy{{Field: "count", Direction: OrderDirection("down")}},
				FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
				ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
			},
			wantErr: "invalid 'orderBy[0].direction'",
		},
		{
			name: "invalid config bounds",
			query: &Query{
				View:          ViewTraces,
				Metrics:       []Metric{{Measure: MeasureCount, Aggregation: "count"}},
				Config:        &QueryConfig{Bins: 101},
				FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
				ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
			},
			wantErr: "'config.bins' must be between 1 and 100",
		},
		{
			name: "invalid measure for view",
			query: &Query{
				View:          ViewScoresCategorical,
				Metrics:       []Metric{{Measure: MeasureValue, Aggregation: "avg"}},
				FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
				ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
			},
			wantErr: `invalid 'metrics[0].measure' "value" for view "scores-categorical"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.query.ToQueryString()
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestQuery_ToQueryStringViewSpecificMeasures(t *testing.T) {
	tests := []struct {
		name    string
		query   *Query
		wantErr string
	}{
		{
			name: "traces supports totalCost",
			query: &Query{
				View:          ViewTraces,
				Metrics:       []Metric{{Measure: MeasureTotalCost, Aggregation: "sum"}},
				FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
				ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
			},
		},
		{
			name: "observations supports timeToFirstToken",
			query: &Query{
				View:          ViewObservations,
				Metrics:       []Metric{{Measure: MeasureTimeToFirstToken, Aggregation: "p95"}},
				FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
				ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
			},
		},
		{
			name: "observations supports input and output tokens",
			query: &Query{
				View: ViewObservations,
				Metrics: []Metric{
					{Measure: MeasureInputTokens, Aggregation: "sum"},
					{Measure: MeasureOutputTokens, Aggregation: "sum"},
				},
				FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
				ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
			},
		},
		{
			name: "scores numeric supports value",
			query: &Query{
				View:          ViewScoresNumeric,
				Metrics:       []Metric{{Measure: MeasureValue, Aggregation: "avg"}},
				FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
				ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
			},
		},
		{
			name: "scores categorical rejects latency",
			query: &Query{
				View:          ViewScoresCategorical,
				Metrics:       []Metric{{Measure: MeasureLatency, Aggregation: "avg"}},
				FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
				ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
			},
			wantErr: `invalid 'metrics[0].measure' "latency" for view "scores-categorical"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.query.ToQueryString()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestQuery_ToQueryStringMarshalError(t *testing.T) {
	query := &Query{
		View:          ViewTraces,
		Metrics:       []Metric{{Measure: MeasureCount, Aggregation: "count"}},
		Filters:       []Filter{{Column: "name", Operator: "=", Value: func() {}}},
		FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
		ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
	}

	_, err := query.ToQueryString()
	require.Error(t, err)
	require.Contains(t, err.Error(), "marshal query failed")
}

func TestRawRowHelpers(t *testing.T) {
	row := RawRow{
		"name":        json.RawMessage(`"assistant"`),
		"count_count": json.RawMessage(`"10"`),
		"histogram":   json.RawMessage(`[[0,10,4]]`),
	}

	rawHistogram, ok := row.Raw("histogram")
	require.True(t, ok)
	require.JSONEq(t, `[[0,10,4]]`, string(rawHistogram))

	var decoded struct {
		Name  string `json:"name"`
		Count string `json:"count_count"`
	}
	err := row.Decode(&decoded)
	require.NoError(t, err)
	require.Equal(t, "assistant", decoded.Name)
	require.Equal(t, "10", decoded.Count)

	typedRows, err := DecodeRows[struct {
		Name  string `json:"name"`
		Count string `json:"count_count"`
	}]([]RawRow{row})
	require.NoError(t, err)
	require.Len(t, typedRows, 1)
	require.Equal(t, "assistant", typedRows[0].Name)
	require.Equal(t, "10", typedRows[0].Count)
}

func TestRawRowHelpersErrors(t *testing.T) {
	row := RawRow{
		"name": json.RawMessage(`123`),
	}

	err := row.Decode(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "'dst' is required")

	_, ok := row.Raw("missing")
	require.False(t, ok)
}

func TestClient_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get metrics", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/metrics", r.URL.Path)
			require.Equal(t, "GET", r.Method)

			var query Query
			err := json.Unmarshal([]byte(r.URL.Query().Get("query")), &query)
			require.NoError(t, err)
			require.Equal(t, ViewTraces, query.View)
			require.Len(t, query.Metrics, 1)
			require.Equal(t, MeasureCount, query.Metrics[0].Measure)

			w.Header().Set("Content-Type", "application/json")
			_, err = w.Write([]byte(`{"data":[{"name":"assistant","count_count":"42"},{"histogram":[[0,10,4]]}]}`))
			require.NoError(t, err)
		}))
		defer server.Close()

		client := NewClient(resty.New().SetBaseURL(server.URL))
		response, err := client.Get(ctx, &Query{
			View:          ViewTraces,
			Metrics:       []Metric{{Measure: MeasureCount, Aggregation: "count"}},
			Dimensions:    []Dimension{{Field: "name"}},
			FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
			ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
		})
		require.NoError(t, err)
		require.Len(t, response.Data, 2)

		rawName, ok := response.Data[0].Raw("name")
		require.True(t, ok)
		var name string
		err = json.Unmarshal(rawName, &name)
		require.NoError(t, err)
		require.Equal(t, "assistant", name)

		rawCount, ok := response.Data[0].Raw("count_count")
		require.True(t, ok)
		var count string
		err = json.Unmarshal(rawCount, &count)
		require.NoError(t, err)
		require.Equal(t, "42", count)

		rows, err := DecodeRows[struct {
			Name  string `json:"name"`
			Count string `json:"count_count"`
		}](response.Data)
		require.NoError(t, err)
		require.Equal(t, "assistant", rows[0].Name)
		require.Equal(t, "42", rows[0].Count)
	})

	t.Run("validation error does not send request", func(t *testing.T) {
		called := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(resty.New().SetBaseURL(server.URL))
		_, err := client.Get(ctx, &Query{
			Metrics:       []Metric{{Measure: MeasureCount, Aggregation: "count"}},
			FromTimestamp: mustParseTime("2024-01-01T00:00:00Z"),
			ToTimestamp:   mustParseTime("2024-01-02T00:00:00Z"),
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "'view' is required")
		require.False(t, called)
	})

	t.Run("curl example observations metrics query", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/metrics", r.URL.Path)
			require.Equal(t, "GET", r.Method)

			var query Query
			err := json.Unmarshal([]byte(r.URL.Query().Get("query")), &query)
			require.NoError(t, err)
			require.Equal(t, ViewObservations, query.View)
			require.Equal(t, []Metric{
				{Measure: MeasureTotalCost, Aggregation: "sum"},
				{Measure: MeasureInputTokens, Aggregation: "sum"},
				{Measure: MeasureOutputTokens, Aggregation: "sum"},
				{Measure: MeasureCount, Aggregation: "count"},
			}, query.Metrics)
			require.Equal(t, []Dimension{
				{Field: "tags"},
				{Field: "providedModelName"},
			}, query.Dimensions)
			require.Empty(t, query.Filters)
			require.Equal(t, mustParseTime("2026-04-29T00:00:00Z"), query.FromTimestamp)
			require.Equal(t, mustParseTime("2026-04-29T23:59:59Z"), query.ToTimestamp)

			w.Header().Set("Content-Type", "application/json")
			_, err = w.Write([]byte(`{"data":[{"tags":["nutrition","meal"],"providedModelName":"gpt-4o-mini","totalCost_sum":"0.42","inputTokens_sum":"1234","outputTokens_sum":"567","count_count":"8"}]}`))
			require.NoError(t, err)
		}))
		defer server.Close()

		client := NewClient(resty.New().SetBaseURL(server.URL))
		response, err := client.Get(ctx, &Query{
			View: ViewObservations,
			Metrics: []Metric{
				{Measure: MeasureTotalCost, Aggregation: "sum"},
				{Measure: MeasureInputTokens, Aggregation: "sum"},
				{Measure: MeasureOutputTokens, Aggregation: "sum"},
				{Measure: MeasureCount, Aggregation: "count"},
			},
			Dimensions: []Dimension{
				{Field: "tags"},
				{Field: "providedModelName"},
			},
			Filters:       []Filter{},
			FromTimestamp: mustParseTime("2026-04-29T00:00:00Z"),
			ToTimestamp:   mustParseTime("2026-04-29T23:59:59Z"),
		})
		require.NoError(t, err)
		require.Len(t, response.Data, 1)

		rawModelName, ok := response.Data[0].Raw("providedModelName")
		require.True(t, ok)
		var modelName string
		err = json.Unmarshal(rawModelName, &modelName)
		require.NoError(t, err)
		require.Equal(t, "gpt-4o-mini", modelName)

		rawTags, ok := response.Data[0].Raw("tags")
		require.True(t, ok)
		require.JSONEq(t, `["nutrition","meal"]`, string(rawTags))

		rows, err := DecodeRows[struct {
			Tags              []string `json:"tags"`
			ProvidedModelName string   `json:"providedModelName"`
			TotalCostSum      string   `json:"totalCost_sum"`
			InputTokensSum    string   `json:"inputTokens_sum"`
			OutputTokensSum   string   `json:"outputTokens_sum"`
			CountCount        string   `json:"count_count"`
		}](response.Data)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.Equal(t, []string{"nutrition", "meal"}, rows[0].Tags)
		require.Equal(t, "gpt-4o-mini", rows[0].ProvidedModelName)
		require.Equal(t, "0.42", rows[0].TotalCostSum)
		require.Equal(t, "1234", rows[0].InputTokensSum)
		require.Equal(t, "567", rows[0].OutputTokensSum)
		require.Equal(t, "8", rows[0].CountCount)
	})
}

func mustParseTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}
