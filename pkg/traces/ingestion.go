package traces

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gofrs/uuid/v5"

	"github.com/git-hulk/langfuse-go/pkg/batch"
)

const (
	IngestionCreateTrace = "trace-create"
	IngestionCreateSpan  = "span-create"
)

type TraceID [16]byte

func (t TraceID) String() string {
	return fmt.Sprintf("%02x", t[:])
}

func FromTraceID(s string) (TraceID, error) {
	var id TraceID
	if len(s) != 32 {
		return TraceID{}, fmt.Errorf("invalid trace ID length: expected 32 hex characters, got %d", len(s))
	}
	for i := 0; i < 16; i++ {
		_, err := fmt.Sscanf(s[i*2:i*2+2], "%02x", &id[i])
		if err != nil {
			return TraceID{}, fmt.Errorf("invalid hex character at position %d: %w", i*2, err)
		}
	}
	return id, nil
}

type SpanID [8]byte

func (s SpanID) String() string {
	return fmt.Sprintf("%02x", s[:])
}

func FromSpanID(s string) (SpanID, error) {
	var id SpanID
	if len(s) != 16 {
		return SpanID{}, fmt.Errorf("invalid span ID length: expected 16 hex characters, got %d", len(s))
	}
	for i := 0; i < 8; i++ {
		_, err := fmt.Sscanf(s[i*2:i*2+2], "%02x", &id[i])
		if err != nil {
			return SpanID{}, fmt.Errorf("invalid hex character at position %d: %w", i*2, err)
		}
	}
	return id, nil
}

type IDGenerator struct {
	sync.Mutex
	source *rand.Rand
}

func NewIDGenerator() *IDGenerator {
	var seed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &seed)
	source := rand.New(rand.NewSource(seed))
	return &IDGenerator{
		source: source,
	}
}

func (g *IDGenerator) GenerateTraceID() TraceID {
	var id TraceID
	g.Lock()
	_, _ = g.source.Read(id[:])
	g.Unlock()
	return id
}

func (g *IDGenerator) GenerateSpanID() SpanID {
	var id SpanID
	g.Lock()
	_, _ = g.source.Read(id[:])
	g.Unlock()
	return id
}

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
	restyCli    *resty.Client
	processor   *batch.Processor[*Trace]
	idGenerator *IDGenerator
}

func NewIngestor(cli *resty.Client) *Ingestor {
	collector := &Ingestor{
		restyCli:    cli,
		idGenerator: NewIDGenerator(),
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

func (ingestor *Ingestor) StartTrace(_ context.Context, name string) *Trace {
	traceID := ingestor.idGenerator.GenerateTraceID().String()
	return ingestor.withTraceID(traceID, name)
}

func (ingestor *Ingestor) withTraceID(id, name string) *Trace {
	return &Trace{
		ingestor:     ingestor,
		observations: make([]*Observation, 0),
		TraceEntry: TraceEntry{
			ID:        id,
			Name:      name,
			Timestamp: time.Now(),
		},
	}
}

func (ingestor *Ingestor) Flush() {
	ingestor.processor.Flush()
}

func (ingestor *Ingestor) Close() error {
	return ingestor.processor.Close()
}
