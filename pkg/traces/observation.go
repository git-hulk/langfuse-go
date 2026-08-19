package traces

import (
	"context"
	"time"

	oteltrace "go.opentelemetry.io/otel/trace"
)

type ObservationType string

const (
	ObservationTypeSpan       ObservationType = "SPAN"
	ObservationTypeGeneration ObservationType = "GENERATION"
	ObservationTypeEvent      ObservationType = "EVENT"
	ObservationTypeAgent      ObservationType = "AGENT"
	ObservationTypeTool       ObservationType = "TOOL"
	ObservationTypeChain      ObservationType = "CHAIN"
	ObservationTypeRetriever  ObservationType = "RETRIEVER"
	ObservationTypeEvaluator  ObservationType = "EVALUATOR"
	ObservationTypeEmbedding  ObservationType = "EMBEDDING"
	ObservationTypeGuardrail  ObservationType = "GUARDRAIL"
)

type UnitType string

const (
	UnitCharacters   UnitType = "CHARACTERS"
	UnitTokens       UnitType = "TOKENS"
	UnitMilliseconds UnitType = "MILLISECONDS"
	UnitSeconds      UnitType = "SECONDS"
	UnitImages       UnitType = "IMAGES"
	UnitRequests     UnitType = "REQUESTS"
)

type ObservationLevel string

const (
	ObservationLevelDebug   ObservationLevel = "DEBUG"
	ObservationLevelDefault ObservationLevel = "DEFAULT"
	ObservationLevelWarning ObservationLevel = "WARNING"
	ObservationLevelError   ObservationLevel = "ERROR"
)

type Usage struct {
	Input  int      `json:"input,omitempty"`
	Output int      `json:"output,omitempty"`
	Total  int      `json:"total,omitempty"`
	Unit   UnitType `json:"unit,omitempty"`
}

type Observation struct {
	ID                  string           `json:"id,omitempty"`
	Type                ObservationType  `json:"type"`
	Name                string           `json:"name,omitempty"`
	PromptName          string           `json:"promptName,omitempty"`
	PromptVersion       int              `json:"promptVersion,omitempty"`
	CompletionStartTime *time.Time       `json:"completionStartTime,omitempty"`
	Model               string           `json:"model,omitempty"`
	ModelParameters     map[string]any   `json:"modelParameters,omitempty"`
	Input               any              `json:"input,omitempty"`
	Metadata            any              `json:"metadata,omitempty"`
	Output              any              `json:"output,omitempty"`
	Usage               Usage            `json:"usage,omitempty"`
	Level               ObservationLevel `json:"level,omitempty"`
	StatusMessage       string           `json:"statusMessage,omitempty"`
	Environment         string           `json:"environment,omitempty"`

	otelCtx       context.Context
	parentOtelCtx context.Context
}

// End finalizes the observation by exporting the OTel span.
func (o *Observation) End() {
	if o.otelCtx == nil {
		return
	}
	span := oteltrace.SpanFromContext(o.otelCtx)
	span.SetAttributes(observationAttributes(o)...)
	span.End()
}
