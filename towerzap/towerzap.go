package towerzap

import (
	"context"

	"github.com/tigorlazuardi/tower"
	"go.uber.org/zap"
)

var _ tower.Logger = (*Logger)(nil)

type TraceCapturer interface {
	CaptureTrace(ctx context.Context) []zap.Field
}

type TraceCapturerFunc func(ctx context.Context) []zap.Field

func (f TraceCapturerFunc) CaptureTrace(ctx context.Context) []zap.Field {
	return f(ctx)
}

type Logger struct {
	*zap.Logger
	tracer TraceCapturer
}

func NewLogger(l *zap.Logger) *Logger {
	return &Logger{
		Logger: l,
		tracer: TraceCapturerFunc(func(ctx context.Context) []zap.Field { return nil }),
	}
}

func (l *Logger) SetTraceCapturer(capturer TraceCapturer) {
	l.tracer = capturer
}

func (l Logger) Log(ctx context.Context, entry tower.Entry) {
	elements := make([]zap.Field, 0, 16)
	elements = append(elements, zap.Time("time", entry.Time()))
	elements = append(elements, l.tracer.CaptureTrace(ctx)...)
	elements = append(elements, zap.Object("service", service(entry.Service())))
	if key := entry.Key(); key != "" {
		elements = append(elements, zap.String("key", key))
	}
	code := entry.Code()
	if code != 0 {
		elements = append(elements, zap.Int("code", code))
	}
	elements = append(elements, zap.Stringer("caller", entry.Caller()))

	data := entry.Context()
	if len(data) == 1 {
		elements = append(elements, toField("context", data[0]))
	} else if len(data) > 1 {
		elements = append(elements, zap.Array("context", encodeContextArray(entry.Context())))
	}

	l.Logger.Log(translateLevel(entry.Level()), entry.Message(), elements...)
}

func (l Logger) LogError(ctx context.Context, err tower.Error) {
	elements := make([]zap.Field, 0, 7)
	elements = append(elements, zap.Time("time", err.Time()))
	elements = append(elements, l.tracer.CaptureTrace(ctx)...)
	elements = append(elements, zap.Object("service", service(err.Service())))
	elements = append(elements, zap.Int("code", err.Code()))
	elements = append(elements, zap.Stringer("caller", err.Caller()))
	if key := err.Key(); key != "" {
		elements = append(elements, zap.String("key", key))
	}
	data := err.Context()
	if len(data) == 1 {
		elements = append(elements, toField("context", data[0]))
	} else if len(data) > 1 {
		elements = append(elements, zap.Array("context", encodeContextArray(err.Context())))
	}
	origin := err.Unwrap()
	elements = append(elements, toField("error", origin))
	l.Logger.Log(translateLevel(err.Level()), err.Message(), elements...)
}
