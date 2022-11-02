package tower

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type Tower struct {
	messengers       Messengers
	logger           Logger
	errorConstructor ErrorGenerator
	service          Service
	optionGenerator  OptionGenerator
}

var (
	_ Logger    = (*Tower)(nil)
	_ Messenger = (*Tower)(nil)
)

// Creates New Tower Instance. Using Built in Generator Engines.
func NewTower(service Service) *Tower {
	return &Tower{
		messengers:       Messengers{},
		logger:           NoopLogger{},
		errorConstructor: ErrorGeneratorFunc(defaultErrorGenerator),
		service:          service,
		optionGenerator:  OptionGeneratorFunc(generateOption),
	}
}

// Returns a CLONE of the registered messengers.
func (t Tower) GetMessengers() Messengers {
	return t.messengers.Clone()
}

// Gets the Messenger by name. Returns nil if not found.
func (t Tower) GetMessengerByName(name string) Messenger {
	return t.messengers[name]
}

// Gets the underlying Logger.
func (t Tower) GetLogger() Logger {
	return t.logger
}

// Gets the service metadata that Tower is running under.
func (t Tower) GetService() Service {
	return t.service
}

// Sets the underlying Logger. This method is NOT concurrent safe.
func (t *Tower) SetLogger(log Logger) {
	t.logger = log
}

func (t Tower) Log(ctx context.Context, entry Entry) {
	t.logger.Log(ctx, entry)
}

func (t Tower) LogError(ctx context.Context, err Error) {
	t.logger.LogError(ctx, err)
}

// Returns the name of the Tower.
func (t Tower) Name() string {
	return fmt.Sprintf("%s-%s-%s", t.service.Name, t.service.Type, t.service.Environment)
}

// Sends notification to all messengers in Tower's known messengers.
// use GetMessengers or GetMessengerByName to get specific messenger.
func (t Tower) SendMessage(ctx MessageContext) {
	for _, v := range t.messengers {
		v.SendMessage(ctx)
	}
}

type multierror []error

func (m multierror) Error() string {
	s := strings.Builder{}
	for i, err := range m {
		s.WriteString(strconv.Itoa(i + 1))
		s.WriteString(". ")
		if i > 0 {
			s.WriteString("; ")
		}
		s.WriteString(err.Error())
	}
	return s.String()
}

// Waits until all message in the queue or until given channel is received.
//
// Implementer must exit the function as soon as possible when this ctx is canceled.
func (t Tower) Wait(ctx context.Context) error {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	wg.Add(len(t.messengers))
	errs := make(multierror, 0, len(t.messengers))
	for _, v := range t.messengers {
		go func(messenger Messenger) {
			defer wg.Done()
			err := messenger.Wait(ctx)
			if err != nil {
				mu.Lock()
				defer mu.Unlock()
				errs = append(errs, err)
			}
		}(v)
	}
	wg.Wait()
	if len(errs) > 0 {
		return errs
	}
	return nil
}
