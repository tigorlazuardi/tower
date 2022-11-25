package tower

import "time"

type MessageContextBuilder interface {
	BuildMessageContext(entry Entry, param MessageParameter) MessageContext
}

var _ MessageContextBuilder = (MessageContextBuilderFunc)(nil)

type MessageContextBuilderFunc func(entry Entry, param MessageParameter) MessageContext

func (f MessageContextBuilderFunc) BuildMessageContext(entry Entry, param MessageParameter) MessageContext {
	return f(entry, param)
}

func defaultMessageContextBuilder(entry Entry, param MessageParameter) MessageContext {
	return &messageContext{inner: entry, param: param}
}

type messageContext struct {
	inner Entry
	param MessageParameter
}

// HTTPCode Gets HTTP Status Code for the type.
func (m messageContext) HTTPCode() int {
	return m.inner.HTTPCode()
}

// Code Gets the original code of the type.
func (m messageContext) Code() int {
	return m.inner.Code()
}

// Message Gets the Message of the type.
func (m messageContext) Message() string {
	return m.inner.Message()
}

// Caller Gets the caller of this type.
func (m messageContext) Caller() Caller {
	return m.inner.Caller()
}

// Key Gets the key for this message.
func (m messageContext) Key() string {
	return m.inner.Key()
}

// Level Gets the level of this message.
func (m messageContext) Level() Level {
	return m.inner.Level()
}

// Service Gets the service information.
func (m messageContext) Service() Service {
	return m.inner.Service()
}

// Context Gets the context of this this type.
func (m messageContext) Context() []any {
	return m.inner.Context()
}

// Time returns the time when this message was created.
func (m messageContext) Time() time.Time {
	return m.inner.Time()
}

// Err returns the error of this message, if set by the sender.
func (m messageContext) Err() error {
	return nil
}

// SkipVerification If true, Sender asks for this message to always be send.
func (m messageContext) SkipVerification() bool {
	return m.param.SkipVerification()
}

// Cooldown Gets the cooldown for this message.
func (m messageContext) Cooldown() time.Duration {
	return m.param.Cooldown()
}

// Tower Gets the tower instance that created this MessageContext.
func (m messageContext) Tower() *Tower {
	return m.param.Tower()
}
