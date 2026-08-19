package traces

import "context"

type traceHandler interface {
	startTrace(ctx context.Context, name string) *Trace
	endTrace(t *Trace)
	startObservation(t *Trace, name string, typ ObservationType) *Observation
	endObservation(o *Observation)
	flush()
	close() error
}
