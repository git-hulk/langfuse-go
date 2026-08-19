package traces

import (
	"encoding/json"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

const (
	AttrTraceName     = "langfuse.trace.name"
	AttrTraceInput    = "langfuse.trace.input"
	AttrTraceOutput   = "langfuse.trace.output"
	AttrTraceTags     = "langfuse.trace.tags"
	AttrTraceMetadata = "langfuse.trace.metadata"

	AttrUserID    = "user.id"
	AttrSessionID = "session.id"

	AttrObservationType            = "langfuse.observation.type"
	AttrObservationInput           = "langfuse.observation.input"
	AttrObservationOutput          = "langfuse.observation.output"
	AttrObservationMetadata        = "langfuse.observation.metadata"
	AttrObservationLevel           = "langfuse.observation.level"
	AttrObservationStatusMessage   = "langfuse.observation.status_message"
	AttrObservationModelName       = "langfuse.observation.model.name"
	AttrObservationModelParameters = "langfuse.observation.model.parameters"
	AttrObservationUsageDetails    = "langfuse.observation.usage_details"
	AttrObservationPromptName      = "langfuse.observation.prompt.name"
	AttrObservationPromptVersion   = "langfuse.observation.prompt.version"
	AttrObservationCompletionStart = "langfuse.observation.completion_start_time"

	AttrEnvironment = "langfuse.environment"
	AttrRelease     = "langfuse.release"
	AttrVersion     = "langfuse.version"
)

func serializeToJSON(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func traceAttributes(t *Trace) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 10)
	attrs = append(attrs, attribute.String(AttrTraceName, t.Name))

	if t.Input != nil {
		attrs = append(attrs, attribute.String(AttrTraceInput, serializeToJSON(t.Input)))
	}
	if t.Output != nil {
		attrs = append(attrs, attribute.String(AttrTraceOutput, serializeToJSON(t.Output)))
	}
	if t.UserID != "" {
		attrs = append(attrs, attribute.String(AttrUserID, t.UserID))
	}
	if t.SessionID != "" {
		attrs = append(attrs, attribute.String(AttrSessionID, t.SessionID))
	}
	if len(t.Tags) > 0 {
		attrs = append(attrs, attribute.String(AttrTraceTags, serializeToJSON(t.Tags)))
	}
	if t.Metadata != nil {
		attrs = append(attrs, attribute.String(AttrTraceMetadata, serializeToJSON(t.Metadata)))
	}
	if t.Release != "" {
		attrs = append(attrs, attribute.String(AttrRelease, t.Release))
	}
	if t.Version != "" {
		attrs = append(attrs, attribute.String(AttrVersion, t.Version))
	}
	if t.Environment != "" {
		attrs = append(attrs, attribute.String(AttrEnvironment, t.Environment))
	}
	return attrs
}

func observationAttributes(o *Observation) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 12)
	attrs = append(attrs, attribute.String(AttrObservationType, strings.ToLower(string(o.Type))))

	if o.Input != nil {
		attrs = append(attrs, attribute.String(AttrObservationInput, serializeToJSON(o.Input)))
	}
	if o.Output != nil {
		attrs = append(attrs, attribute.String(AttrObservationOutput, serializeToJSON(o.Output)))
	}
	if o.Metadata != nil {
		attrs = append(attrs, attribute.String(AttrObservationMetadata, serializeToJSON(o.Metadata)))
	}
	if o.Level != "" {
		attrs = append(attrs, attribute.String(AttrObservationLevel, string(o.Level)))
	}
	if o.StatusMessage != "" {
		attrs = append(attrs, attribute.String(AttrObservationStatusMessage, o.StatusMessage))
	}
	if o.Model != "" {
		attrs = append(attrs, attribute.String(AttrObservationModelName, o.Model))
	}
	if len(o.ModelParameters) > 0 {
		attrs = append(attrs, attribute.String(AttrObservationModelParameters, serializeToJSON(o.ModelParameters)))
	}
	if o.Usage != (Usage{}) {
		attrs = append(attrs, attribute.String(AttrObservationUsageDetails, serializeToJSON(o.Usage)))
	}
	if o.PromptName != "" {
		attrs = append(attrs, attribute.String(AttrObservationPromptName, o.PromptName))
	}
	if o.PromptVersion != 0 {
		attrs = append(attrs, attribute.Int(AttrObservationPromptVersion, o.PromptVersion))
	}
	if o.CompletionStartTime != nil {
		attrs = append(attrs, attribute.String(AttrObservationCompletionStart, serializeToJSON(o.CompletionStartTime)))
	}
	if o.Environment != "" {
		attrs = append(attrs, attribute.String(AttrEnvironment, o.Environment))
	}
	return attrs
}
