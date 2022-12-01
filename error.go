package tower

import (
	"context"
	"time"
)

type ErrorConstructorContext struct {
	Err    error
	Caller Caller
	Tower  *Tower
}

type ErrorConstructor interface {
	ConstructError(*ErrorConstructorContext) ErrorBuilder
}

type ErrorConstructorFunc func(*ErrorConstructorContext) ErrorBuilder

func (f ErrorConstructorFunc) ConstructError(ctx *ErrorConstructorContext) ErrorBuilder {
	return f(ctx)
}

func defaultErrorGenerator(ctx *ErrorConstructorContext) ErrorBuilder {
	var message string
	if msg := Query.GetMessage(ctx.Err); msg != "" {
		message = msg
	} else {
		message = ctx.Err.Error()
	}
	return &errorBuilder{
		code:    Query.GetCodeHint(ctx.Err),
		message: message,
		caller:  ctx.Caller,
		context: []any{},
		level:   ErrorLevel,
		origin:  ctx.Err,
		tower:   ctx.Tower,
		time:    time.Now(),
	}
}

/*
Error is an interface providing read only values to the error, and because it's read only, this is safe for multithreaded use.
*/
type Error interface {
	error
	CallerHint
	CodeHint
	ContextHint
	ErrorUnwrapper
	ErrorWriter
	HTTPCodeHint
	KeyHint
	LevelHint
	MessageHint
	TimeHint
	ServiceHint

	/*
		Logs this error.
	*/
	Log(ctx context.Context) Error
	/*
		Notifies this error to Messengers.
	*/
	Notify(ctx context.Context, opts ...MessageOption) Error
}

type ErrorUnwrapper interface {
	// Unwrap Returns the error that is wrapped by this error. To be used by errors.Is and errors.As functions from errors library.
	Unwrap() error
}
