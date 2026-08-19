package traces

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestObservation_End(t *testing.T) {
	ingestor, _ := newTestOtelIngestor(t)
	defer ingestor.Close()

	trace := ingestor.StartTrace(context.Background(), "test-trace")
	observation := trace.StartSpan("test-span")

	require.NotNil(t, observation.otelCtx)
	span := oteltrace.SpanFromContext(observation.otelCtx)
	assert.True(t, span.IsRecording())

	observation.End()

	assert.False(t, span.IsRecording())
}

func TestObservation_Fields(t *testing.T) {
	usage := &Usage{
		Input:  100,
		Output: 50,
		Total:  150,
		Unit:   UnitTokens,
	}

	observation := &Observation{
		Type:            ObservationTypeGeneration,
		Name:            "test-generation",
		Model:           "gpt-4",
		ModelParameters: map[string]any{"temperature": 0.7},
		PromptName:      "test-prompt",
		PromptVersion:   1,
		Input:           "test input",
		Metadata:        map[string]any{"key": "value"},
		Output:          "test output",
		Usage:           *usage,
		Level:           ObservationLevelDefault,
		StatusMessage:   "completed",
		Environment:     "test",
	}

	assert.Equal(t, ObservationTypeGeneration, observation.Type)
	assert.Equal(t, "test-generation", observation.Name)
	assert.Equal(t, "gpt-4", observation.Model)
	assert.Equal(t, map[string]any{"temperature": 0.7}, observation.ModelParameters)
	assert.Equal(t, "test-prompt", observation.PromptName)
	assert.Equal(t, 1, observation.PromptVersion)
	assert.Equal(t, "test input", observation.Input)
	assert.Equal(t, map[string]any{"key": "value"}, observation.Metadata)
	assert.Equal(t, "test output", observation.Output)
	assert.Equal(t, *usage, observation.Usage)
	assert.Equal(t, ObservationLevelDefault, observation.Level)
	assert.Equal(t, "completed", observation.StatusMessage)
	assert.Equal(t, "test", observation.Environment)
}

func TestObservationType_Constants(t *testing.T) {
	assert.Equal(t, ObservationType("SPAN"), ObservationTypeSpan)
	assert.Equal(t, ObservationType("GENERATION"), ObservationTypeGeneration)
}

func TestUnitType_Constants(t *testing.T) {
	assert.Equal(t, UnitType("CHARACTERS"), UnitCharacters)
	assert.Equal(t, UnitType("TOKENS"), UnitTokens)
	assert.Equal(t, UnitType("MILLISECONDS"), UnitMilliseconds)
	assert.Equal(t, UnitType("SECONDS"), UnitSeconds)
	assert.Equal(t, UnitType("IMAGES"), UnitImages)
	assert.Equal(t, UnitType("REQUESTS"), UnitRequests)
}

func TestObservationLevel_Constants(t *testing.T) {
	assert.Equal(t, ObservationLevel("DEBUG"), ObservationLevelDebug)
	assert.Equal(t, ObservationLevel("DEFAULT"), ObservationLevelDefault)
	assert.Equal(t, ObservationLevel("WARNING"), ObservationLevelWarning)
	assert.Equal(t, ObservationLevel("ERROR"), ObservationLevelError)
}

func TestUsage_Fields(t *testing.T) {
	usage := &Usage{
		Input:  100,
		Output: 50,
		Total:  150,
		Unit:   UnitTokens,
	}

	assert.Equal(t, 100, usage.Input)
	assert.Equal(t, 50, usage.Output)
	assert.Equal(t, 150, usage.Total)
	assert.Equal(t, UnitTokens, usage.Unit)
}
