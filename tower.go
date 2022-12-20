package tower

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// Tower is an instance of Tower.
type Tower struct {
	messengers                 Messengers
	logger                     Logger
	errorConstructor           ErrorConstructor
	entryConstructor           EntryConstructor
	service                    Service
	defaultNotifyOption        []MessageOption
	errorMessageContextBuilder ErrorMessageContextBuilder
	messageContextBuilder      MessageContextBuilder
	callerDepth                int
}

// SetCallerDepth Sets the depth of the caller to be used when constructing the ErrorBuilder.
func (t *Tower) SetCallerDepth(callerDepth int) {
	t.callerDepth = callerDepth
}

var (
	_ Logger    = (*Tower)(nil)
	_ Messenger = (*Tower)(nil)
)

// NewTower Creates New Tower Instance. Using Built in Generator Engines by default.
func NewTower(service Service) *Tower {
	return &Tower{
		messengers:                 Messengers{},
		logger:                     NoopLogger{},
		errorConstructor:           ErrorConstructorFunc(defaultErrorGenerator),
		entryConstructor:           EntryConstructorFunc(defaultEntryConstructor),
		errorMessageContextBuilder: ErrorMessageContextBuilderFunc(defaultErrorMessageContextBuilder),
		messageContextBuilder:      MessageContextBuilderFunc(defaultMessageContextBuilder),
		service:                    service,
		callerDepth:                2,
	}
}

// RegisterMessenger Registers a messenger to the tower.
//
// The messenger's name should be unique. Same name will replace the previous messenger with the same name.
//
// If you wish to have multiple messengers of the same type, you should use different names for each of them.
func (t *Tower) RegisterMessenger(messenger Messenger) {
	t.messengers[messenger.Name()] = messenger
}

// RemoveMessenger Removes the Messenger by name.
func (t *Tower) RemoveMessenger(name string) {
	delete(t.messengers, name)
}

// Wrap like exported tower.Wrap, but at the scope of this Tower's instance instead.
func (t *Tower) Wrap(err error) ErrorBuilder {
	if err == nil {
		err = errors.New("<nil>")
	}
	caller := GetCaller(t.callerDepth)
	return t.errorConstructor.ConstructError(&ErrorConstructorContext{
		Err:    err,
		Caller: caller,
		Tower:  t,
	})
}

// NewEntry Creates a new EntryBuilder. The returned EntryBuilder may be appended with values.
func (t *Tower) NewEntry(msg string, args ...any) EntryBuilder {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	caller := GetCaller(t.callerDepth)
	return t.entryConstructor.ConstructEntry(&EntryConstructorContext{
		Caller:  caller,
		Tower:   t,
		Message: msg,
	})
}

// Bail creates a new ErrorBuilder from simple messages.
//
// If args is not empty, msg will be fed into fmt.Errorf along with the args.
// Otherwise, msg will be fed into `errors.New()`.
func (t *Tower) Bail(msg string, args ...any) ErrorBuilder {
	var err error
	if len(args) > 0 {
		err = fmt.Errorf(msg, args...)
	} else {
		err = errors.New(msg)
	}
	caller := GetCaller(t.callerDepth)
	return t.errorConstructor.ConstructError(&ErrorConstructorContext{
		Err:    err,
		Caller: caller,
		Tower:  t,
	})
}

// BailFreeze creates new immutable Error from simple messages.
//
// If args is not empty, msg will be fed into fmt.Errorf along with the args.
// Otherwise, msg will be fed into `errors.New()`.
func (t *Tower) BailFreeze(msg string, args ...any) Error {
	var err error
	if len(args) > 0 {
		err = fmt.Errorf(msg, args...)
	} else {
		err = errors.New(msg)
	}
	caller := GetCaller(t.callerDepth)
	return t.errorConstructor.ConstructError(&ErrorConstructorContext{
		Err:    err,
		Caller: caller,
		Tower:  t,
	}).Freeze()
}

// SetErrorConstructor Sets how the ErrorBuilder will be constructed.
func (t *Tower) SetErrorConstructor(c ErrorConstructor) {
	t.errorConstructor = c
}

// SetErrorMessageContextBuilder Sets how the MessageContext will be built from tower.Error.
func (t *Tower) SetErrorMessageContextBuilder(b ErrorMessageContextBuilder) {
	t.errorMessageContextBuilder = b
}

// SetMessageContextBuilder Sets how the MessageContext will be built from tower.Entry.
func (t *Tower) SetMessageContextBuilder(b MessageContextBuilder) {
	t.messageContextBuilder = b
}

// SetEntryConstructor Sets how the EntryBuilder will be constructed.
func (t *Tower) SetEntryConstructor(c EntryConstructor) {
	t.entryConstructor = c
}

// SetDefaultNotifyOption Sets the default options for Notify and NotifyError.
// When Notify or NotifyError is called, the default options will be applied first,
// then followed by the options passed in on premise by the user.
func (t *Tower) SetDefaultNotifyOption(opts ...MessageOption) {
	t.defaultNotifyOption = opts
}

// WrapFreeze is a Shorthand for tower.Wrap(err).Message(message).Freeze()
//
// Useful when just wanting to add extra simple messages to the error chain.
// If args is not empty, message will be fed into fmt.Errorf along with the args.
func (t *Tower) WrapFreeze(err error, message string, args ...any) Error {
	caller := GetCaller(t.callerDepth)
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	return t.errorConstructor.ConstructError(&ErrorConstructorContext{
		Err:    err,
		Caller: caller,
		Tower:  t,
	}).
		Message(message).
		Freeze()
}

// GetMessengers Returns a CLONE of the registered messengers.
func (t Tower) GetMessengers() Messengers {
	return t.messengers.Clone()
}

// GetMessengerByName Gets the Messenger by name. Returns nil if not found.
func (t Tower) GetMessengerByName(name string) Messenger {
	return t.messengers[name]
}

// GetLogger Gets the underlying Logger.
func (t Tower) GetLogger() Logger {
	return t.logger
}

// GetService Gets the service metadata that Tower is running under.
func (t Tower) GetService() Service {
	return t.service
}

// SetLogger Sets the underlying Logger. This method is NOT concurrent safe.
func (t *Tower) SetLogger(log Logger) {
	t.logger = log
}

// Notify Sends the Entry to Messengers.
func (t Tower) Notify(ctx context.Context, entry Entry, parameters ...MessageOption) {
	opts := t.createOption(parameters...)
	msg := t.messageContextBuilder.BuildMessageContext(entry, opts)
	t.sendNotif(ctx, msg, opts)
}

// NotifyError Sends the Error to Messengers.
func (t Tower) NotifyError(ctx context.Context, err Error, parameters ...MessageOption) {
	opts := t.createOption(parameters...)
	msg := t.errorMessageContextBuilder.BuildErrorMessageContext(err, opts)
	t.sendNotif(ctx, msg, opts)
}

func (t *Tower) createOption(parameters ...MessageOption) *messageOption {
	opts := &messageOption{tower: t}
	for _, v := range t.defaultNotifyOption {
		v.apply(opts)
	}
	for _, v := range parameters {
		v.apply(opts)
	}
	return opts
}

func (t Tower) sendNotif(ctx context.Context, msg MessageContext, opts *messageOption) {
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

// Log Implements tower.Logger interface. So The Tower instance itself may be used as Logger Engine.
func (t Tower) Log(ctx context.Context, entry Entry) {
	t.logger.Log(ctx, entry)
}

// LogError Implements tower.Logger interface. So The Tower instance itself may be used as Logger Engine.
func (t Tower) LogError(ctx context.Context, err Error) {
	t.logger.LogError(ctx, err)
}

// Name Implements tower.Messenger interface. So The Tower instance itself may be used as Messenger.
//
// Returns the service registered in the format of "tower-service_name-service_type-service_environment".
func (t Tower) Name() string {
	return "tower-" + t.service.String()
}

// SendMessage Implements tower.Messenger interface. So The Tower instance itself may be used as Messenger.
//
// Sends notification to all messengers registered in this instance.
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

// Wait Implements tower.Messenger interface. So The Tower instance itself may be used as Messenger.
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
