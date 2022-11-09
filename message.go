package tower

import (
	"strings"
	"time"
)

type MessageContextBuilder interface {
	BuildMessageContext(entry Entry, param MessageParameter) MessageContext
}

type ErrorMessageContextBuilder interface {
	BuildErrorMessageContext(err Error, param MessageParameter) MessageContext
}

type ErrorMessageContextBuilderFunc func(err Error, param MessageParameter) MessageContext

func (f ErrorMessageContextBuilderFunc) BuildErrorMessageContext(err Error, param MessageParameter) MessageContext {
	return f(err, param)
}

func defaultErrorMessageContextBuilder(err Error, param MessageParameter) MessageContext {
	return &errorMessageContext{inner: err, param: param}
}

type MessageContext interface {
	BodyCodeHint
	HTTPCodeHint
	CodeHint
	MessageHint
	CallerHint
	KeyHint
	LevelHint
	ServiceHint
	ContextHint
	TimeHint
	// Error item. May be nil if message contains no error.
	Err() error
	// If true, Sender asks for this message to always be send.
	SkipVerification() bool
	Cooldown() time.Duration
	// Gets the tower instance that created this MessageContext.
	Tower() *Tower
}

type errorMessageContext struct {
	inner Error
	param MessageParameter
}

// Gets the Body Code for the type.
func (e errorMessageContext) BodyCode() int {
	return e.inner.BodyCode()
}

// Gets HTTP Status Code for the type.
func (e errorMessageContext) HTTPCode() int {
	return e.inner.HTTPCode()
}

// Gets the original code of the type.
func (e errorMessageContext) Code() int {
	return e.inner.Code()
}

// Gets the Message of the type.
func (e errorMessageContext) Message() string {
	return e.inner.Message()
}

// Gets the caller of this type.
func (e errorMessageContext) Caller() Caller {
	return e.inner.Caller()
}

// Gets the key for this message.
func (e errorMessageContext) Key() string {
	return e.inner.Key()
}

// Gets the level of this message.
func (e errorMessageContext) Level() Level {
	return e.inner.Level()
}

// Gets the service information.
func (e errorMessageContext) Service() Service {
	return e.param.Tower().GetService()
}

// Gets the context of this this type.
func (e errorMessageContext) Context() []any {
	return e.inner.Context()
}

func (e errorMessageContext) Time() time.Time {
	return e.inner.Time()
}

// Error item. May be nil if message contains no error.
func (e errorMessageContext) Err() error {
	return e.inner
}

// If true, Sender asks for this message to always be send.
func (e errorMessageContext) SkipVerification() bool {
	return e.param.SkipVerification()
}

// Gets the tower instance that created this MessageContext.
func (e errorMessageContext) Tower() *Tower {
	return e.param.Tower()
}

func (e errorMessageContext) Cooldown() time.Duration {
	return e.param.Cooldown()
}

type MessageOption interface {
	apply(*option)
}

type MessageParameter interface {
	SkipVerification() bool
	Cooldown() time.Duration
	Tower() *Tower
}

type option struct {
	skipVerification  bool
	specificMessenger Messenger
	messengers        Messengers
	cooldown          time.Duration
	tower             *Tower
}

func (o option) SkipVerification() bool {
	return o.skipVerification
}

func (o option) SpecificMessenger() Messenger {
	return o.specificMessenger
}

func (o option) Messengers() Messengers {
	return o.messengers
}

func (o option) Cooldown() time.Duration {
	return o.cooldown
}

func (o option) Tower() *Tower {
	return o.tower
}

type MessageOptionFunc func(*option)

func (f MessageOptionFunc) apply(opt *option) {
	f(opt)
}

// Asks the Messengers to Skip cooldown verifications and just send the message.
func SkipMessageVerification(b bool) MessageOption {
	return MessageOptionFunc(func(ob *option) {
		ob.skipVerification = b
	})
}

/*
Asks Tower to only send only to the Messenger with this name.
If name is not found, Tower returns to default behaviour.

Note: OnlyMessengerWithName option will conflict with other Messenger setters option, and thus only the latest option will be set.
*/
func OnlyMessengerWithName(name string) MessageOption {
	return MessageOptionFunc(func(ob *option) {
		ob.specificMessenger = ob.tower.GetMessengerByName(name)
		ob.messengers = nil
	})
}

/*
Asks Tower to only send only to this Messenger.

Note: OnlyThisMessenger option will conflict with other Messenger setters option, and thus only the latest option will be set.
*/
func OnlyThisMessenger(m Messenger) MessageOption {
	return MessageOptionFunc(func(ob *option) {
		ob.messengers = nil
		ob.specificMessenger = m
	})
}

func OnlyTheseMessengers(m ...Messenger) MessageOption {
	return MessageOptionFunc(func(ob *option) {
		ob.specificMessenger = nil
		mm := make(Messengers, len(m))
		for _, v := range m {
			mm[v.Name()] = v
		}
		ob.messengers = mm
	})
}

/*
Asks Tower to only send messages to Messengers whose name begins with given s.

Note: MessengerPrefix option will conflict with other Messenger setters option, and thus only the latest option will be set.
*/
func MessengerPrefix(s string) MessageOption {
	return MessageOptionFunc(func(ob *option) {
		ob.specificMessenger = nil
		messengers := ob.tower.GetMessengers()
		mm := make(Messengers, len(messengers))
		for k, v := range messengers {
			if strings.HasPrefix(k, s) {
				mm[k] = v
			}
		}
		ob.messengers = mm
	})
}

/*
Asks Tower to only send messages to Messengers whose name ends with given s.

Note: MessengerSuffix option will conflict with other Messenger setters option, and thus only the latest option will be set.
*/
func MessengerSuffix(s string) MessageOption {
	return MessageOptionFunc(func(ob *option) {
		ob.specificMessenger = nil
		messengers := ob.tower.GetMessengers()
		mm := make(Messengers, len(messengers))
		for k, v := range messengers {
			if strings.HasSuffix(k, s) {
				mm[k] = v
			}
		}
		ob.messengers = mm
	})
}

/*
Asks Tower to only send messages to Messengers whose name contains given s.

Note: MessengerNameContains option will conflict with other Messenger setters option, and thus only the latest option will be set.
*/
func MessengerNameContains(s string) MessageOption {
	return MessageOptionFunc(func(ob *option) {
		ob.specificMessenger = nil
		messengers := ob.tower.GetMessengers()
		mm := make(Messengers, len(messengers))
		for k, v := range messengers {
			if strings.Contains(k, s) {
				mm[k] = v
			}
		}
		ob.messengers = mm
	})
}

/*
Sets the Cooldown for this Message.
*/
func MessageCooldown(dur time.Duration) MessageOption {
	return MessageOptionFunc(func(ob *option) {
		ob.cooldown = dur
	})
}

/*
Asks Tower to send messages to currenty registered and also send those messeges to these Messengers.

Note: MessengerNameContains option will conflict with other Messenger setters option, and thus only the latest option will be set.
*/
func ExtraMessengers(messengers ...Messenger) MessageOption {
	return MessageOptionFunc(func(ob *option) {
		ob.specificMessenger = nil
		ob.messengers = ob.tower.GetMessengers()
		for _, v := range messengers {
			ob.messengers[v.Name()] = v
		}
	})
}
