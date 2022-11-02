package towerzap

import (
	"github.com/tigorlazuardi/tower"
	"go.uber.org/zap/zapcore"
)

type fields tower.Fields

func (f fields) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for k, v := range f {
		switch v := v.(type) {
		case zapcore.ObjectMarshaler:
			if err := enc.AddObject(k, v); err != nil {
				return err
			}
		case zapcore.ArrayMarshaler:
			if err := enc.AddArray(k, v); err != nil {
				return err
			}
		default:
			if err := enc.AddReflected(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
