// Package traces provides functionality for distributed tracing in Langfuse.
//
// This package implements hierarchical tracing with traces containing observations (spans),
// using OpenTelemetry for trace export. Traces represent execution flows in your application
// and can contain metadata, inputs, outputs, and nested observations.
package traces

import (
	"context"

	oteltrace "go.opentelemetry.io/otel/trace"
)

type tracer = oteltrace.Tracer

// TraceEntry represents the core data structure for a trace in Langfuse.
//
// A trace captures a single execution flow in your application with
// input/output data, user context, and metadata. Traces can be associated
// with sessions and contain nested observations (spans).
type TraceEntry struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Input       any      `json:"input,omitempty"`
	Output      any      `json:"output,omitempty"`
	SessionID   string   `json:"sessionId,omitempty"`
	Release     string   `json:"release,omitempty"`
	Version     string   `json:"version,omitempty"`
	UserID      string   `json:"userId,omitempty"`
	Metadata    any      `json:"metadata,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Environment string   `json:"environment,omitempty"`
}

// Trace represents an active trace that can be used to create observations and manage execution flow.
//
// A Trace embeds TraceEntry and provides methods to create child observations (spans),
// end the trace, and submit the trace for export via OpenTelemetry.
// Traces are automatically assigned unique IDs when created.
type Trace struct {
	TraceEntry
	oteltrace.Span

	tracer          tracer
	lastObservation *Observation
	otelCtx         context.Context
}

// End finalizes the trace by exporting the OTel span with trace attributes.
func (t *Trace) End() {
	if t.Span == nil {
		return
	}
	t.SetAttributes(traceAttributes(t)...)
	t.Span.End()
}

func (t *Trace) getParentContext() context.Context {
	if t.lastObservation == nil {
		return t.otelCtx
	}
	if t.lastObservation.IsRecording() {
		return t.lastObservation.otelCtx
	}
	return t.lastObservation.parentOtelCtx
}

// StartSpan creates a new child observation (span) within this trace.
//
// The span is automatically assigned a unique ID, set to span type, and linked
// to this trace via OpenTelemetry context propagation.
// Returns an Observation that can be used to add data and end the span.
func (t *Trace) StartSpan(name string) *Observation {
	return t.StartObservation(name, ObservationTypeSpan)
}

// StartObservation creates a new child observation of the specified type within this trace.
//
// The observation is automatically assigned a unique ID and linked to this trace
// via OpenTelemetry context propagation.
// Returns an Observation that can be used to add data and end the observation.
func (t *Trace) StartObservation(name string, typ ObservationType) *Observation {
	parentCtx := t.getParentContext()
	ctx, span := t.tracer.Start(parentCtx, name)
	observation := &Observation{
		Span:          span,
		Name:          name,
		Type:          typ,
		otelCtx:       ctx,
		parentOtelCtx: parentCtx,
	}
	t.lastObservation = observation
	return observation
}

// StartGeneration creates a new child observation (generation) within this trace.
//
// The generation is automatically assigned a unique ID, set to generation type, and linked
// to this trace via OpenTelemetry context propagation.
// Returns an Observation that can be used to add data and end the generation.
func (t *Trace) StartGeneration(name string) *Observation {
	return t.StartObservation(name, ObservationTypeGeneration)
}
