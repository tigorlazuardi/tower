package towerzap

import "go.uber.org/zap"

type Logger struct {
	*zap.Logger
}
