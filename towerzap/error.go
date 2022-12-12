package towerzap

import (
	"github.com/tigorlazuardi/tower"
	"go.uber.org/zap/zapcore"
)

type Error struct {
	tower.Error
}

type richJsonError struct{ error }

func (r richJsonError) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if r.error == nil {
		_ = enc.AddReflected("error", nil)
		return nil
	}
	enc.AddString("summary", r.error.Error())
	err := enc.AddReflected("details", r.error)
	if err != nil {
		enc.AddString("details", "failed to marshal error details")
	}
	return nil
}

func (err Error) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("code", err.Code())
	enc.AddString("message", err.Message())
	enc.AddString("caller", err.Caller().String())
	if key := err.Key(); key != "" {
		enc.AddString("key", key)
	}
	_ = enc.AddObject("service", service(err.Service()))
	data := err.Context()
	if len(data) == 1 {
		err := enc.AddReflected("context", data[0])
		if err != nil {
			enc.AddString("context", "failed to marshal context")
		}
	} else if len(data) > 1 {
		err := enc.AddArray("context", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
			for _, value := range data {
				switch value := value.(type) {
				case tower.Fields:
					return ae.AppendObject(fields(value))
				case zapcore.ObjectMarshaler:
					return ae.AppendObject(value)
				case zapcore.ArrayMarshaler:
					return ae.AppendArray(value)
				default:
					return ae.AppendReflected(value)
				}
			}
			return nil
		}))
		if err != nil {
			enc.AddString("context", "failed to marshal context")
		}
	}

	origin := err.Unwrap()
	if origin == nil {
		_ = enc.AddReflected("error", nil)
		return nil
	}

	var errMarshal error
	switch err := origin.(type) {
	case tower.Error:
		errMarshal = enc.AddObject("error", Error{err})
	case zapcore.ObjectMarshaler:
		errMarshal = enc.AddObject("error", err)
	case zapcore.ArrayMarshaler:
		errMarshal = enc.AddArray("error", err)
	default:
		errMarshal = enc.AddObject("error", richJsonError{err})
	}
	if errMarshal != nil {
		enc.AddString("error", "failed to marshal error")
	}
	return nil
}
