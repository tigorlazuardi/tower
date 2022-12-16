package towerzap

import (
	"github.com/tigorlazuardi/tower"
	"go.uber.org/zap/zapcore"
)

type fields tower.Fields

func (f fields) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for k, v := range f {
		if err := encodeObject(enc, k, v); err != nil {
			enc.AddString(k, err.Error())
		}
	}
	return nil
}
