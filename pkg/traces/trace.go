// Package traces provides functionality for distributed tracing in Langfuse.
//
// This package implements hierarchical tracing with traces containing observations (spans),
// using efficient batch processing for ingestion. Traces represent execution flows
// in your application and can contain metadata, inputs, outputs, and nested observations.
package traces

import (
	"time"

	"go.uber.org/zap"

	"github.com/git-hulk/langfuse-go/pkg/logger"
)

// TraceEntry represents the core data structure for a trace in Langfuse.
//
// A trace captures a single execution flow in your application with timing,
// input/output data, user context, and metadata. Traces can be associated
// with sessions and contain nested observations (spans).
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

// Trace represents an active trace that can be used to create observations and manage execution flow.
//
// A Trace embeds TraceEntry and provides methods to create child observations (spans),
// end the trace with automatic latency calculation, and submit the trace for batch processing.
// Traces are automatically assigned unique IDs and timestamps when created.
type Trace struct {
	TraceEntry

	ingestor     *Ingestor
	observations []*Observation
}

// End finalizes the trace by calculating its latency and submitting it for batch processing.
//
// This method calculates the total latency from the trace's start timestamp to now,
// then submits the trace to the batch processor for efficient ingestion to Langfuse.
// If submission fails, an error is logged but the method does not return an error.
func (t *Trace) End() {
	t.Latency = time.Since(t.Timestamp).Milliseconds()
	if err := t.ingestor.processor.Submit(t); err != nil {
		logger.Get().With(
			zap.Error(err),
			zap.String("trace_name", t.Name),
		).Error("Failed to submit trace for processing")
	}
}

func (t *Trace) getParentObservationID() string {
	if len(t.observations) == 0 {
		return t.ID // If no observations, use trace ID as parent
	}

	lastObservation := t.observations[len(t.observations)-1]
	if lastObservation.EndTime == nil || lastObservation.EndTime.IsZero() {
		return lastObservation.ID // Use last observation ID if it's still active
	}
	return lastObservation.ParentObservationID // Use parent observation ID of the last observation
}

// StartSpan creates a new child observation (span) within this trace.
//
// The span is automatically assigned a unique ID, set to span type, and linked
// to this trace. The span's start time is set to the current time.
// Returns an Observation that can be used to add data and end the span.
func (t *Trace) StartSpan(name string) *Observation {
	observation := t.StartObservation(name, ObservationTypeSpan)
	return observation
}

// StartObservation creates a new child observation of the specified type within this trace.
//
// The observation is automatically assigned a unique ID and linked to this trace.
// The observation's start time is set to the current time.
// Returns an Observation that can be used to add data and end the observation.
func (t *Trace) StartObservation(name string, typ ObservationType) *Observation {
	observationID := t.ingestor.idGenerator.GenerateSpanID().String()
	observation := &Observation{
		TraceID:             t.ID,
		ID:                  observationID,
		Name:                name,
		Type:                typ,
		ParentObservationID: t.getParentObservationID(),
		StartTime:           time.Now(),
	}
	t.observations = append(t.observations, observation)
	return observation
}

// StartGeneration creates a new child observation (generation) within this trace.
//
// The generation is automatically assigned a unique ID, set to generation type, and linked
// to this trace. The generation's start time is set to the current time.
// Returns an Observation that can be used to add data and end the generation.
func (t *Trace) StartGeneration(name string) *Observation {
	observation := t.StartObservation(name, ObservationTypeGeneration)
	return observation
}
