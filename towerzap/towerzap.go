package towerzap

import (
	"context"
	"fmt"

	"github.com/tigorlazuardi/tower"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
		elements = append(elements, zap.Array("context", encodeContext(entry.Context())))
	}

	l.Logger.Log(translateLevel(entry.Level()), entry.Message(), elements...)
}

func (l Logger) LogError(ctx context.Context, err tower.Error) {
	elements := make([]zap.Field, 0, 7)
	elements = append(elements, zap.Time("time", err.Time()))
	elements = append(elements, l.tracer.CaptureTrace(ctx)...)
	elements = append(elements, zap.Int("code", err.Code()))
	elements = append(elements, zap.Stringer("caller", err.Caller()))

	data := err.Context()
	if len(data) == 1 {
		elements = append(elements, toField("context", data[0]))
	} else if len(data) > 1 {
		elements = append(elements, zap.Array("context", encodeContext(err.Context())))
	}
	elements = append(elements, zap.Error(err))
	l.Logger.Log(translateLevel(err.Level()), err.Message(), elements...)
}

func toField(key string, value any) zap.Field {
	switch value := value.(type) {
	case tower.Fields:
		return zap.Object(key, fields(value))
	case zapcore.ObjectMarshaler:
		return zap.Object(key, value)
	case zapcore.ArrayMarshaler:
		return zap.Array(key, value)
	case map[string]any:
		return zap.Object(key, fields(value))
	default:
		return zap.Any(key, value)
	}
}

func translateLevel(lvl tower.Level) zapcore.Level {
	switch lvl {
	case tower.DebugLevel:
		return zapcore.DebugLevel
	case tower.InfoLevel:
		return zapcore.InfoLevel
	case tower.WarnLevel:
		return zapcore.WarnLevel
	case tower.ErrorLevel:
		return zapcore.ErrorLevel
	case tower.FatalLevel:
		return zapcore.FatalLevel
	case tower.PanicLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InvalidLevel
	}
}

func encodeContext(ctx []any) zapcore.ArrayMarshaler {
	return zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, value := range ctx {
			var err error
			switch value := value.(type) {
			case tower.Fields:
				err = ae.AppendObject(fields(value))
			case zapcore.ObjectMarshaler:
				err = ae.AppendObject(value)
			case zapcore.ArrayMarshaler:
				err = ae.AppendArray(value)
			case map[string]any:
				err = ae.AppendObject(fields(value))
			default:
				err = ae.AppendReflected(value)
			}
			if err != nil {
				ae.AppendString(fmt.Sprintf("failed marshal log: %v", err))
			}
		}
		return nil
	})
}
