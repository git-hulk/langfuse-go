package traces

import (
	"time"
)

type ObservationType string

const (
	ObservationTypeSpan       ObservationType = "SPAN"
	ObservationTypeGeneration ObservationType = "GENERATION"
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
	ID                  string                 `json:"id,omitempty"`
	TraceID             string                 `json:"traceId,omitempty"`
	Type                ObservationType        `json:"type"`
	Name                string                 `json:"name,omitempty"`
	StartTime           time.Time              `json:"startTime,omitempty"`
	EndTime             time.Time              `json:"endTime,omitempty"`
	CompletionStartTime time.Time              `json:"completionStartTime,omitempty"`
	Model               string                 `json:"model,omitempty"`
	ModelParameters     map[string]interface{} `json:"modelParameters,omitempty"`
	Input               any                    `json:"input,omitempty"`
	Version             string                 `json:"version,omitempty"`
	Metadata            any                    `json:"metadata,omitempty"`
	Output              any                    `json:"output,omitempty"`
	Usage               *Usage                 `json:"usage,omitempty"`
	Level               ObservationLevel       `json:"level,omitempty"`
	StatusMessage       string                 `json:"statusMessage,omitempty"`
	ParentObservationID string                 `json:"parentObservationId,omitempty"`
	Environment         string                 `json:"environment,omitempty"`
}

func (o *Observation) End() {
	o.EndTime = time.Now()
}
