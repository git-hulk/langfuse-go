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

	tracer tracer
}

// End finalizes the trace by exporting the OTel span with trace attributes.
func (t *Trace) End() {
	if t.Span == nil {
		return
	}
	t.SetAttributes(traceAttributes(t)...)
	t.Span.End()
}

// StartSpan creates a new observation (span) as a child of the span carried by ctx.
//
// The span is set to span type and linked to its parent via OpenTelemetry context
// propagation, so pass the context returned by StartTrace to nest it under the trace,
// or the context returned by another StartXXX call to nest it under that observation.
// Returns the context carrying the new span and an Observation that can be used to
// add data and end the span.
func (t *Trace) StartSpan(ctx context.Context, name string) (context.Context, *Observation) {
	return t.StartObservation(ctx, name, ObservationTypeSpan)
}

// StartObservation creates a new observation of the specified type as a child of the
// span carried by ctx.
//
// The observation is linked to its parent via OpenTelemetry context propagation, so
// pass the context returned by StartTrace to nest it under the trace, or the context
// returned by another StartXXX call to nest it under that observation.
// Returns the context carrying the new span and an Observation that can be used to
// add data and end the observation.
func (t *Trace) StartObservation(ctx context.Context, name string, typ ObservationType) (context.Context, *Observation) {
	ctx, span := t.tracer.Start(ctx, name)
	return ctx, &Observation{
		Span: span,
		Name: name,
		Type: typ,
	}
}

// StartGeneration creates a new observation (generation) as a child of the span carried by ctx.
//
// The generation is set to generation type and linked to its parent via OpenTelemetry
// context propagation, so pass the context returned by StartTrace to nest it under the
// trace, or the context returned by another StartXXX call to nest it under that observation.
// Returns the context carrying the new span and an Observation that can be used to
// add data and end the generation.
func (t *Trace) StartGeneration(ctx context.Context, name string) (context.Context, *Observation) {
	return t.StartObservation(ctx, name, ObservationTypeGeneration)
}
