package towerzap

import (
	"github.com/tigorlazuardi/tower"
	"go.uber.org/zap/zapcore"
)

type service tower.Service

func (s service) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if s.Name != "" {
		enc.AddString("name", s.Name)
	}
	if s.Type != "" {
		enc.AddString("type", s.Type)
	}
	if s.Environment != "" {
		enc.AddString("environment", s.Environment)
	}
	if s.Version != "" {
		enc.AddString("version", s.Version)
	}
	return nil
}
