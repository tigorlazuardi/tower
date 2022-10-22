package tower

import "go.uber.org/zap"

type Tower struct {
	messengers       Messengers
	logger           *zap.Logger
	errorConstructor ErrorGenerator
	service          Service
	traceCapturer    TraceCapturer
	optionGenerator  OptionGenerator
}

// Creates New Tower Instance. Using Built in Generator Engines.
func NewTower(service Service) *Tower {
	log, err := zap.NewProduction(zap.AddCallerSkip(1), zap.Fields(zap.Object("service", service)))
	if err != nil {
		panic(err)
	}
	return &Tower{
		messengers:       Messengers{},
		logger:           log,
		errorConstructor: ErrorGeneratorFunc(defaultErrorGenerator),
		service:          service,
		traceCapturer:    TraceCaptureFunc(noopCapturer),
		optionGenerator:  OptionGeneratorFunc(generateOption),
	}
}

// Returns a CLONE of the registered messengers.
func (t *Tower) GetMessengers() Messengers {
	return t.messengers.Clone()
}

// Gets the Messenger by name. Returns nil if not found.
func (t *Tower) GetMessengerByName(name string) Messenger {
	return t.messengers[name]
}

// Gets the underlying Logger.
func (t *Tower) GetLogger() *zap.Logger {
	return t.logger
}

// Gets the service metadata that Tower is running under.
func (t *Tower) GetService() Service {
	return t.service
}

// Sets the underlying Logger. This method is NOT concurrent safe.
func (t *Tower) SetLogger(log *zap.Logger) {
	t.logger = log
}
