package traces

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/git-hulk/langfuse-go/pkg/batch"

	"github.com/go-resty/resty/v2"
	"github.com/gofrs/uuid/v5"
)

const (
	IngestionCreateTrace = "traces-create"
	IngestionCreateSpan  = "span-create"
)

type IngestionEvent struct {
	ID        string    `json:"id,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Type      string    `json:"type,omitempty"`
	Body      any       `json:"body,omitempty"`
}

type IngestionError struct {
	ID      string `json:"id,omitempty"`
	Status  int    `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Error   any    `json:"error,omitempty"`
}

type Ingestor struct {
	restyCli  *resty.Client
	processor *batch.Processor[*Trace]
}

func NewIngestor(cli *resty.Client) *Ingestor {
	collector := &Ingestor{
		restyCli: cli,
	}
	collector.processor = batch.NewProcessor[*Trace](collector)
	return collector
}

func (ingestor *Ingestor) TracesToEvents(traces []*Trace) []IngestionEvent {
	events := make([]IngestionEvent, 0, len(traces))
	for _, trace := range traces {
		events = append(events, IngestionEvent{
			ID:        uuid.Must(uuid.NewV4()).String(),
			Timestamp: trace.Timestamp,
			Type:      IngestionCreateTrace,
			Body:      trace,
		})
		for _, observation := range trace.observations {
			events = append(events, IngestionEvent{
				ID:        uuid.Must(uuid.NewV4()).String(),
				Timestamp: observation.StartTime,
				Type:      IngestionCreateSpan,
				Body:      observation,
			})
		}
	}
	return events
}

func (ingestor *Ingestor) Send(ctx context.Context, traces []*Trace) error {
	if len(traces) == 0 {
		return nil
	}
	events := ingestor.TracesToEvents(traces)
	rsp, err := ingestor.restyCli.R().
		SetContext(ctx).
		SetBody(map[string]interface{}{"batch": events}).
		Post("/ingestion")
	if err != nil {
		return err
	}

	var ingestResponse struct {
		Errors []IngestionError `json:"errors"`
	}
	if err := json.Unmarshal(rsp.Body(), &ingestResponse); err != nil {
		return fmt.Errorf("failed to unmarshal ingestion response: %w", err)
	}
	if len(ingestResponse.Errors) > 0 {
		return fmt.Errorf("ingestion errors: %v", ingestResponse.Errors)
	}
	if rsp.IsError() {
		return fmt.Errorf("send traces got unxpected status code: %d", rsp.StatusCode())
	}
	return nil
}

func (ingestor *Ingestor) StartTrace(Name string) *Trace {
	return &Trace{
		ingestor:     ingestor,
		observations: make([]*Observation, 0),
		TraceEntry: TraceEntry{
			ID:        uuid.Must(uuid.NewV4()).String(),
			Name:      Name,
			Timestamp: time.Now(),
		},
	}
}

func (ingestor *Ingestor) Close() error {
	return ingestor.processor.Close()
}
