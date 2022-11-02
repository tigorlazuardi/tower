package tower

import "context"

// Tracer that does nothing.
type NoopTracer struct{}

func (NoopTracer) GetTraceID() string {
	return ""
}

func (NoopTracer) GetTransactionID() string {
	return ""
}

// A common captured interface of Trace identifications.
type Trace interface {
	GetTraceID() string
	GetTransactionID() string
}

type TraceCapturer interface {
	CaptureTrace(ctx context.Context) Trace
}

type TraceCaptureFunc func(ctx context.Context) Trace

func (t TraceCaptureFunc) CaptureTrace(ctx context.Context) Trace {
	return t(ctx)
}
