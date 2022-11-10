package tower

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// An instance of Tower.
type Tower struct {
	messengers                 Messengers
	logger                     Logger
	errorConstructor           ErrorConstructor
	entryConstructor           EntryConstructor
	service                    Service
	defaultNotifyOption        []MessageOption
	errorMessageContextBuilder ErrorMessageContextBuilder
	messageContextBuilder      MessageContextBuilder
}

var (
	_ Logger    = (*Tower)(nil)
	_ Messenger = (*Tower)(nil)
)

// Creates New Tower Instance. Using Built in Generator Engines.
func NewTower(service Service) *Tower {
	return &Tower{
		messengers:                 Messengers{},
		logger:                     NoopLogger{},
		errorConstructor:           ErrorConstructorFunc(defaultErrorGenerator),
		entryConstructor:           EntryConstructorFunc(defaultEntryConstructor),
		errorMessageContextBuilder: ErrorMessageContextBuilderFunc(defaultErrorMessageContextBuilder),
		messageContextBuilder:      MessageContextBuilderFunc(defaultMessageContextBuilder),
		service:                    service,
	}
}

// Wraps this error. The returned ErrorBuilder may be appended with values.
func (t *Tower) Wrap(err error) ErrorBuilder {
	if err == nil {
		err = errors.New("<nil>")
	}
	caller, _ := GetCaller(2)
	return t.errorConstructor.ContructError(&ErrorConstructorContext{
		Err:    err,
		Caller: caller,
		Tower:  t,
	})
}

// Sets how the ErrorBuilder will be constructed.
func (t *Tower) SetErrorConstructor(c ErrorConstructor) {
	t.errorConstructor = c
}

// Sets how the Message Context will be built from tower.Error.
func (t *Tower) SetErrorMessageContextBuilder(b ErrorMessageContextBuilder) {
	t.errorMessageContextBuilder = b
}

// Sets how the Message Context will be built from tower.Entry.
func (t *Tower) SetMessageContextBuilder(b MessageContextBuilder) {
	t.messageContextBuilder = b
}

// Sets how the EntryBuilder will be constructed.
func (t *Tower) SetEntryConstructor(c EntryConstructor) {
	t.entryConstructor = c
}

// Sets the default options for Notify and NotifyError.
// When Notify or NotifyError is called, the default options will be applied first,
// then followed by the options passed in on premise by the user.
func (t *Tower) SetDefaultNotifyOption(opts ...MessageOption) {
	t.defaultNotifyOption = opts
}

// Shorthand for tower.Wrap(err).Message(message).Freeze()
//
// Useful when just wanting to add extra simple messages to the error chain.
func (t *Tower) WrapFreeze(err error, message string) Error {
	caller, _ := GetCaller(2)
	return t.errorConstructor.ContructError(&ErrorConstructorContext{
		Err:    err,
		Caller: caller,
		Tower:  t,
	}).
		Message(message).
		Freeze()
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

// Sends the Entry to Messengers.
func (t Tower) Notify(ctx context.Context, entry Entry, parameters ...MessageOption) {
	opts := t.createOption(parameters...)
	msg := t.messageContextBuilder.BuildMessageContext(entry, opts)
	t.sendNotif(ctx, msg, opts)
}

// Sends the Error to Messengers.
func (t Tower) NotifyError(ctx context.Context, err Error, parameters ...MessageOption) {
	opts := t.createOption(parameters...)
	msg := t.errorMessageContextBuilder.BuildErrorMessageContext(err, opts)
	t.sendNotif(ctx, msg, opts)
}

func (t Tower) createOption(parameters ...MessageOption) *option {
	opts := &option{}
	for _, v := range t.defaultNotifyOption {
		v.apply(opts)
	}
	for _, v := range parameters {
		v.apply(opts)
	}
	return opts
}

func (t Tower) sendNotif(ctx context.Context, msg MessageContext, opts *option) {
	ctx = DetachedContext(ctx)
	if opts.specificMessenger != nil {
		go opts.specificMessenger.SendMessage(ctx, msg)
		return
	}
	if len(opts.messengers) > 0 {
		for _, messenger := range opts.messengers {
			go messenger.SendMessage(ctx, msg)
		}
		return
	}
	for _, messenger := range t.messengers {
		go messenger.SendMessage(ctx, msg)
	}
}

// Implements tower.Logger interface. So The Tower instance itself may be used as Logger Engine.
func (t Tower) Log(ctx context.Context, entry Entry) {
	t.logger.Log(ctx, entry)
}

// Implements tower.Logger interface. So The Tower instance itself may be used as Logger Engine.
func (t Tower) LogError(ctx context.Context, err Error) {
	t.logger.LogError(ctx, err)
}

// Implements tower.Messenger interface. So The Tower instance itself may be used as Messenger.
//
// Returns the service registered in the format of "tower-service_name-service_type-service_environment".
func (t Tower) Name() string {
	return "tower-" + t.service.String()
}

// Implements tower.Messenger interface. So The Tower instance itself may be used as Messenger.
//
// Sends notification to all messengers in Tower's known messengers.
// use GetMessengers or GetMessengerByName to get specific messenger.
func (t Tower) SendMessage(ctx context.Context, msg MessageContext) {
	for _, v := range t.messengers {
		v.SendMessage(ctx, msg)
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

// Implements tower.Messenger interface. So The Tower instance itself may be used as Messenger.
//
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
				errs = append(errs, fmt.Errorf("failed on waiting messages to finish from '%s': %w", messenger.Name(), err))
				mu.Unlock()
			}
		}(v)
	}
	wg.Wait()
	if len(errs) > 0 {
		return errs
	}
	return nil
}
