package traces

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObservation_End(t *testing.T) {
	startTime := time.Now()
	observation := &Observation{
		ID:        "test-observation-id",
		StartTime: startTime,
	}

	time.Sleep(10 * time.Millisecond)
	observation.End()

	require.NotNil(t, observation.EndTime)
	assert.True(t, observation.EndTime.After(startTime))
}

func TestObservation_Fields(t *testing.T) {
	usage := &Usage{
		Input:  100,
		Output: 50,
		Total:  150,
		Unit:   UnitTokens,
	}

	observation := &Observation{
		ID:                  "obs-123",
		TraceID:             "traces-456",
		Type:                ObservationTypeGeneration,
		Name:                "test-generation",
		Model:               "gpt-4",
		ModelParameters:     map[string]any{"temperature": 0.7},
		PromptName:          "test-prompt",
		PromptVersion:       1,
		Input:               "test input",
		Version:             "1.0",
		Metadata:            map[string]any{"key": "value"},
		Output:              "test output",
		Usage:               *usage,
		Level:               ObservationLevelDefault,
		StatusMessage:       "completed",
		ParentObservationID: "parent-789",
		Environment:         "test",
	}

	assert.Equal(t, "obs-123", observation.ID)
	assert.Equal(t, "traces-456", observation.TraceID)
	assert.Equal(t, ObservationTypeGeneration, observation.Type)
	assert.Equal(t, "test-generation", observation.Name)
	assert.Equal(t, "gpt-4", observation.Model)
	assert.Equal(t, map[string]any{"temperature": 0.7}, observation.ModelParameters)
	assert.Equal(t, "test-prompt", observation.PromptName)
	assert.Equal(t, 1, observation.PromptVersion)
	assert.Equal(t, "test input", observation.Input)
	assert.Equal(t, "1.0", observation.Version)
	assert.Equal(t, map[string]any{"key": "value"}, observation.Metadata)
	assert.Equal(t, "test output", observation.Output)
	assert.Equal(t, *usage, observation.Usage)
	assert.Equal(t, ObservationLevelDefault, observation.Level)
	assert.Equal(t, "completed", observation.StatusMessage)
	assert.Equal(t, "parent-789", observation.ParentObservationID)
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
