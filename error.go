package tower

import (
	"context"
	"fmt"
	"time"
)

type ErrorConstructorContext struct {
	Err    error
	Caller Caller
	Tower  *Tower
}

type ErrorConstructor interface {
	ContructError(*ErrorConstructorContext) ErrorBuilder
}

type ErrorConstructorFunc func(*ErrorConstructorContext) ErrorBuilder

func (f ErrorConstructorFunc) ContructError(ctx *ErrorConstructorContext) ErrorBuilder {
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
	}
}

/*
ErrorBuilder is an interface to create customizable error.

ErrorBuilder by itself is not an error type. You have to call .Freeze() method to create proper Error type.
*/
type ErrorBuilder interface {
	/*
		Sets the error code for this error.
		This is used to identify the error type and how towerhttp will interact with this error.

		The default implementation for Tower handles code like this:

		The default value, if anything in the error implements CodeHint interface,
		use the outermost CodeHint implementer, otherwise fallsback to 500.

		`tower.Error` that is generated by tower implements CodeHinter interface,
		and thus if the error is wrapped again, the default code for this error will be the code of the wrapped `tower.Error`.

		Example:

			if err != nil {
				return tower.Wrap(err).Code(500).Freeze()
			}
	*/
	Code(i int) ErrorBuilder

	/*
		Overrides the error message for this error.

		In built in implementation, If args are supplied, fmt.Sprintf will be called with s as base string.
	*/
	Message(s string, args ...any) ErrorBuilder

	/*
		Sets the origin error for ErrorBuilder. Very unlikely to need to set this because tower.Wrap already wraps the error.
		But the api is available to set the origin error.
	*/
	Error(err error) ErrorBuilder

	/*
		Sets additional data that will enrich how the error will look.

		`tower.Fields` is a type that is well integrated with built in Messengers.
		Using this type as Context value will have the performance more optimized when being marshaled
		or provides additional quality of life improvements without having to implement those features
		yourself. Use `tower.F` as alias for this type.

		In built-in implementation, additional call to .Context() will make additional index, not replacing what you already set.

		Example:

			tower.Wrap(err).Code(400).Context(tower.F{"foo": "bar"}).Freeze()
	*/
	Context(ctx interface{}) ErrorBuilder

	/*
		Sets the key for this error. This is how Messenger will use to identify if an error is the same as previous or not.

		In tower's built-in implementation, by default, no key is set.

		Usually by not setting the key, The Messenger will generate their own.

		In built in implementation, If args are supplied, fmt.Sprintf will be called with key as base string.
	*/
	Key(key string, args ...any) ErrorBuilder

	/*
		Sets the caller for this error.

		In tower's built-in implementation, by default, the caller is the location where you call `tower.Wrap` or `tower.WrapFreeze`
	*/
	Caller(c Caller) ErrorBuilder

	/*
		Sets the level for this error.

		In tower's built-in implementation, this defaults to ErrorLevel if not set.
	*/
	Level(lvl Level) ErrorBuilder

	/*
		Sets the time for this error.

		In tower's built-in implementation, this is already set to when tower.Wrap is called.
	*/
	Time(t time.Time) ErrorBuilder

	/*
		Freeze this ErrorBuilder, preventing further mutations and set this ErrorBuilder into proper error.

		The returned Error is safe for multithreaded usage because of it's immutable nature.
	*/
	Freeze() Error

	/*
		Logs this error. Implicitly calls .Freeze() on this ErrorBuilder.
	*/
	Log(ctx context.Context) Error
	/*
		Notifies this error to Messengers. Implicitly calls .Freeze() on this ErrorBuilder.
	*/
	Notify(ctx context.Context, opts ...MessageOption) Error
}

type errorBuilder struct {
	code    int
	message string
	caller  Caller
	context []any
	key     string
	level   Level
	origin  error
	tower   *Tower
	time    time.Time
}

func (e *errorBuilder) Level(lvl Level) ErrorBuilder {
	e.level = lvl
	return e
}

func (e *errorBuilder) Caller(c Caller) ErrorBuilder {
	e.caller = c
	return e
}

func (e *errorBuilder) Code(i int) ErrorBuilder {
	e.code = i
	return e
}

func (e *errorBuilder) Error(err error) ErrorBuilder {
	e.origin = err
	return e
}

func (e *errorBuilder) Message(s string, args ...any) ErrorBuilder {
	if len(args) > 0 {
		e.message = fmt.Sprintf(s, args...)
	} else {
		e.message = s
	}
	return e
}

func (e *errorBuilder) Context(ctx interface{}) ErrorBuilder {
	e.context = append(e.context, ctx)
	return e
}

func (e *errorBuilder) Key(key string, args ...any) ErrorBuilder {
	if len(args) > 0 {
		e.key = fmt.Sprintf(key, args...)
	} else {
		e.key = key
	}
	return e
}

func (e *errorBuilder) Time(t time.Time) ErrorBuilder {
	e.time = t
	return e
}

func (e *errorBuilder) Freeze() Error {
	return implError{inner: e}
}

func (e *errorBuilder) Log(ctx context.Context) Error {
	return e.Freeze().Log(ctx)
}

func (e *errorBuilder) Notify(ctx context.Context, opts ...MessageOption) Error {
	return e.Freeze().Notify(ctx, opts...)
}

/*
Error is an interface providing read only values to the error, and because it's read only, this is safe for multithreaded use.
*/
type Error interface {
	error
	CodeHint
	HTTPCodeHint
	MessageHint
	CallerHint
	ContextHint
	LevelHint
	ErrorUnwrapper
	ErrorWriter
	TimeHint

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
	// Returns the error that is wrapped by this error. To be used by errors.Is and errors.As functions from errors library.
	Unwrap() error
}

type implError struct {
	inner *errorBuilder
}

func (e implError) Error() string {
	s := NewLineWriterBuilder().Separator(" => ").Build()
	e.WriteError(s)
	return s.String()
}

// Writes the error.Error to the writer instead of being allocated as value.
func (e implError) WriteError(w LineWriter) {
	w.WriteIndent()
	msg := e.inner.message
	if e.inner.origin == nil {
		if len(msg) > 0 {
			w.WritePrefix()
			_, _ = w.WriteString(msg)
			w.WriteSuffix()
			w.WriteSeparator()
		}
		w.WritePrefix()
		_, _ = w.WriteString("[nil]")
		w.WriteSuffix()
		return
	}

	errMsg := e.inner.origin.Error()
	if ew, ok := e.inner.origin.(ErrorWriter); ok {
		if mh, ok := e.inner.origin.(MessageHint); ok && msg != mh.Message() && msg != errMsg {
			w.WritePrefix()
			_, _ = w.WriteString(msg)
			w.WriteSuffix()
			w.WriteSeparator()
		} else if msg != errMsg {
			w.WritePrefix()
			_, _ = w.WriteString(msg)
			w.WriteSeparator()
			w.WriteSuffix()
		}
		w.WritePrefix()
		ew.WriteError(w)
		w.WriteSuffix()
		return
	}
	if msg != errMsg {
		w.WritePrefix()
		_, _ = w.WriteString(msg)
		w.WriteSuffix()
		w.WriteSeparator()
	}
	w.WritePrefix()
	_, _ = w.WriteString(errMsg)
	w.WriteSuffix()
}

// Gets the original code of the type.
func (e implError) Code() int {
	return e.inner.code
}

// Gets HTTP Status Code for the type.
func (e implError) HTTPCode() int {
	switch {
	case e.inner.code >= 200 && e.inner.code <= 599:
		return e.inner.code
	case e.inner.code > 999:
		code := e.inner.code % 1000
		if code >= 200 && code <= 599 {
			return code
		}
	}
	return 500
}

// Gets the Message of the type.
func (e implError) Message() string {
	return e.inner.message
}

// Gets the caller of this type.
func (e implError) Caller() Caller {
	return e.inner.caller
}

// Gets the context of this this type.
func (e implError) Context() []any {
	return e.inner.context
}

func (e implError) Level() Level {
	return e.inner.level
}

func (e implError) Time() time.Time {
	return e.inner.time
}

// Returns the error that is wrapped by this error. To be used by errors.Is and errors.As functions from errors library.
func (e implError) Unwrap() error {
	return e.inner.origin
}

/*
Logs this error.
*/
func (e implError) Log(ctx context.Context) Error {
	e.inner.tower.LogError(ctx, e)
	return e
}

/*
Notifies this error to Messengers.
*/
func (e implError) Notify(ctx context.Context, opts ...MessageOption) Error {
	e.inner.tower.NotifyError(ctx, e, opts...)
	return e
}
