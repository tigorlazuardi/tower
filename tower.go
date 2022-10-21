package tower

import "go.uber.org/zap"

type Tower struct {
	messengers       map[string]Messenger
	logger           *zap.Logger
	errorConstructor ErrorGenerator
	service          Service
}
