package tower

import "context"

type noopTrace struct{}

func (noopTrace) GetTraceID() string {
	return ""
}

func (noopTrace) GetTransactionID() string {
	return ""
}

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

func noopCapturer(ctx context.Context) Trace {
	return noopTrace{}
}
