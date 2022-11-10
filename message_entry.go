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

// Gets the Body Code for the type.
func (m messageContext) BodyCode() int {
	return m.inner.BodyCode()
}

// Gets HTTP Status Code for the type.
func (m messageContext) HTTPCode() int {
	return m.inner.HTTPCode()
}

// Gets the original code of the type.
func (m messageContext) Code() int {
	return m.inner.Code()
}

// Gets the Message of the type.
func (m messageContext) Message() string {
	return m.inner.Message()
}

// Gets the caller of this type.
func (m messageContext) Caller() Caller {
	return m.inner.Caller()
}

// Gets the key for this message.
func (m messageContext) Key() string {
	return m.inner.Key()
}

// Gets the level of this message.
func (m messageContext) Level() Level {
	return m.inner.Level()
}

// Gets the service information.
func (m messageContext) Service() Service {
	return m.inner.Service()
}

// Gets the context of this this type.
func (m messageContext) Context() []any {
	return m.inner.Context()
}

func (m messageContext) Time() time.Time {
	return m.inner.Time()
}

// Error item. May be nil if message contains no error.
func (m messageContext) Err() error {
	return nil
}

// If true, Sender asks for this message to always be send.
func (m messageContext) SkipVerification() bool {
	return m.param.SkipVerification()
}

func (m messageContext) Cooldown() time.Duration {
	return m.param.Cooldown()
}

// Gets the tower instance that created this MessageContext.
func (m messageContext) Tower() *Tower {
	return m.param.Tower()
}
