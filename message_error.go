package tower

import "time"

type ErrorMessageContextBuilder interface {
	BuildErrorMessageContext(err Error, param MessageParameter) MessageContext
}

var _ ErrorMessageContextBuilder = (ErrorMessageContextBuilderFunc)(nil)

type ErrorMessageContextBuilderFunc func(err Error, param MessageParameter) MessageContext

func (f ErrorMessageContextBuilderFunc) BuildErrorMessageContext(err Error, param MessageParameter) MessageContext {
	return f(err, param)
}

func defaultErrorMessageContextBuilder(err Error, param MessageParameter) MessageContext {
	return &errorMessageContext{inner: err, param: param}
}

var _ MessageContext = (*errorMessageContext)(nil)

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
