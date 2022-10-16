package tower

import "go.uber.org/zap/zapcore"

type Fields map[string]interface{}

func (f Fields) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for k, v := range f {
		switch t := v.(type) {
		case zapcore.ObjectMarshaler:
			err := enc.AddObject(k, t)
			if err != nil {
				return err
			}
		case zapcore.ArrayMarshaler:
			err := enc.AddArray(k, t)
			if err != nil {
				return err
			}
		default:
			err := enc.AddReflected(k, v)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
