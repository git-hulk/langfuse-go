package traces

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type TraceEntry struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
	Input       any       `json:"input,omitempty"`
	Output      any       `json:"output,omitempty"`
	SessionID   string    `json:"sessionId,omitempty"`
	Release     string    `json:"release,omitempty"`
	Version     string    `json:"version,omitempty"`
	UserID      string    `json:"userId,omitempty"`
	Metadata    any       `json:"metadata,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Latency     int64     `json:"latency,omitempty"`   // in milliseconds
	TotalCost   float64   `json:"totalCost,omitempty"` // in USD
	Environment string    `json:"environment,omitempty"`
}

type Trace struct {
	TraceEntry

	ingestor     *Ingestor
	observations []*Observation
}

func (t *Trace) End() {
	t.Latency = time.Since(t.Timestamp).Milliseconds()
	if err := t.ingestor.processor.Submit(t); err != nil {
		// TODO: handle error, e.g., log it
	}
}

func (t *Trace) StartSpan(name string) *Observation {
	observation := &Observation{
		TraceID:             t.ID,
		ID:                  uuid.Must(uuid.NewV4()).String(),
		Name:                name,
		Type:                ObservationTypeSpan,
		ParentObservationID: t.ID,
		StartTime:           time.Now(),
	}
	t.observations = append(t.observations, observation)
	return observation
}
