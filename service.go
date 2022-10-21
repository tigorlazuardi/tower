package tower

import (
	"github.com/francoispqt/gojay"
	"go.uber.org/zap/zapcore"
)

type Service struct {
	Name        string
	Environment string
	Repository  string
	Branch      string
	Type        string
}

func (s Service) IsNil() bool {
	return s.Name == ""
}

func (s Service) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", s.Name)
	if len(s.Type) > 0 {
		enc.AddString("type", s.Type)
	}
	if len(s.Environment) > 0 {
		enc.AddString("environment", s.Environment)
	}

	return nil
}

func (s Service) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddStringKey("name", s.Name)
	enc.AddStringKeyOmitEmpty("type", s.Type)
	enc.AddStringKeyOmitEmpty("repository", s.Repository)
	enc.AddStringKeyOmitEmpty("branch", s.Branch)
	enc.AddStringKeyOmitEmpty("environment", s.Environment)
}
