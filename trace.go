package tower

import "context"

type KeyValue[K, V any] struct {
	Key   K
	Value V
}

type Trace []KeyValue[string, string]

// Tracer that does nothing.
type NoopTracer struct{}

func (n NoopTracer) CaptureTrace(ctx context.Context) Trace {
	return nil
}

type TraceCapturer interface {
	CaptureTrace(ctx context.Context) Trace
}

type TraceCaptureFunc func(ctx context.Context) Trace

func (t TraceCaptureFunc) CaptureTrace(ctx context.Context) Trace {
	return t(ctx)
}
