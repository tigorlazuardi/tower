package tower

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ErrorGeneratorContext struct {
	Err     error
	Caller  Caller
	Service Service
	Tower   *Tower
}

type ErrorGenerator interface {
	ContructError(*ErrorGeneratorContext) ErrorBuilder
}

type ErrorGeneratorFunc func(*ErrorGeneratorContext) ErrorBuilder

func (f ErrorGeneratorFunc) ContructError(ctx *ErrorGeneratorContext) ErrorBuilder {
	return f(ctx)
}

type ErrorEntry interface {
	error
	/*
		Logs this error.
	*/
	Log(ctx context.Context, opts ...zap.Option) ErrorEntry
	/*
		Notifies this error to Messengers.
	*/
	Notify(ctx context.Context, opts ...MessageOption) ErrorEntry
}

/*
ErrorBuilder is an interface to create customizable error.

It's not recommended to cast an error to this interface, because:

 1. The implementation for the error is not guaranteed to be thread safe.
 2. This api exposes mutable references to the error, which can cause point 1.
 3. What you want most likely only read access to the error values, which this interface will not give.
*/
type ErrorBuilder interface {
	error
	ErrorEntry

	/*
		Sets the error code for this error.
		This is used to identify the error type and how towerhttp will interact with this error.

		A code between 200 and 599 will make towerhttp return this error as a HTTP response with the given status code.
		The body code will also reflect the error code.

		A code between 2200 and 4999 will make towerhttp return status code of 200,
		but the body will code will still the same as origin.

		A code between 5000 and 5599 will make towerhttp return a modulus 1000 of the code as the status code,
		but the body code will still the same as origin.

		Other values will make towerhttp return 500 http status code and the body code will be 5500.

		A default value of 5500 will be used if this method is not called,
		unless the wrapped error implements `tower.CodeHinter` interface,
		in which case the default code for this error will be whatever that interface value will return.

		`tower.Error` that is generated by tower implements CodeHinter interface,
		and thus if the error is wrapped again, the default code for this error will be the code of the wrapped error.

		Example:

			if err != nil {
				return tower.Wrap(err).Code(500)
			}
	*/
	SetCode(i int) ErrorBuilder

	/*
		Sets the error message for this error.

		If the wrapped error implements tower.MessageHinter interface,
		the default message for this error will be whatever that interface value will return.
		Otherwise, it will take the error message by caling .Error() method.

		Calling this method will override the default behaviour.
	*/
	SetMessage(s string) ErrorBuilder

	/*
		Sets the error data for this error.

		Use this to add more information to the error.

		Like for example, user id or email address.

		To customize the way the data is marshalled, implement zapcore.ObjectMarshaler interface
		for the object you want to pass in.

		tower.Fields type can be used to easily add context to the error.

		Example:

			tower.Wrap(err).Code(400).SetContext(tower.Fields{"foo": "bar"})
	*/
	SetContext(ctx interface{}) ErrorBuilder

	/*
		Sets the key for this error. This is how Messenger will use to identify if an error is the same as previous or not.

		By default, the key is the file:line from where the error is created.

		Tower Logger operation will ignore the key completely.
		No matter the circumstances, Tower will always try to log the error when the .Log(ctx) method is invoked.

		The key however, may be used by Messenger to determine how they will treat the error.
	*/
	SetKey(key string) ErrorBuilder
}

/*
Error is an interface providing read only values to the error, and because it's read only, this is safe for multithreaded use.
*/
type Error interface {
	error
	CodeHint
	HTTPCodeHint
	BodyCodeHint
	MessageHint
	CallerHint
	ContextHint
	ErrorUnwrapper

	/*
		Signals the Tower library that this error should be logged.
	*/
	LogError(ctx context.Context) ErrorBuilder

	/*
		Signals the Tower library that this error should be send to Messengers.
	*/
	NotifyError(ctx context.Context) ErrorBuilder
}

type ErrorUnwrapper interface {
	// Returns the error that is wrapped by this error. To be used by errors.Is and errors.As functions from errors library.
	Unwrap() error
}

var _ ErrorBuilder = (*errorEntry)(nil)

type errorEntry struct {
	// Who calls tower API.
	caller Caller
	// Channel target to post the message to if Notifier supports it.
	key     string
	message string
	// Earliest time this message with the same key may be repeated.
	service Service
	// data Object
	data []any
	// Error item.
	error error
	// Message level.
	level zapcore.Level
	code  int
	tower *Tower
}

func (e errorEntry) Error() string {
	w := &strings.Builder{}
	e.WriteError(w)
	return w.String()
}

// implements ErrorWriter.
func (e errorEntry) WriteError(w Writer) {
	if e.error == nil {
		if len(e.message) > 0 {
			_, _ = w.WriteString(e.message)
			_, _ = w.WriteString(" => ")
		}
		_, _ = w.WriteString("[nil]")
		return
	}
	if ew, ok := e.error.(ErrorWriter); ok { //nolint:errorlint
		if mh, ok := e.error.(MessageHint); ok && e.message == mh.Message() { //nolint:errorlint
			ew.WriteError(w)
			return
		}
		_, _ = w.WriteString(e.message)
		_, _ = w.WriteString(" => ")
		ew.WriteError(w)
		return
	}
	errMsg := e.error.Error()
	if mh, ok := e.error.(MessageHint); ok { //nolint:errorlint
		hint := mh.Message()
		if e.message == hint && e.message == errMsg {
			_, _ = w.WriteString(hint)
			return
		} else if e.message == hint {
			_, _ = w.WriteString(hint)
			_, _ = w.WriteString(" => ")
			_, _ = w.WriteString(errMsg)
			return
		}
	}
	if e.message != errMsg {
		_, _ = w.WriteString(e.message)
		_, _ = w.WriteString(" => ")
	}
	_, _ = w.WriteString(errMsg)
}

func (e *errorEntry) SetCode(i int) ErrorBuilder {
	e.code = i
	return e
}

/*
Sets the error message for this error.
If the wrapped error implements tower.MessageHinter interface,
the default message for this error will be whatever that interface value will return.
Otherwise, it will take the error message by caling .Error() method.
Calling this method will override the default behaviour.
*/
func (e *errorEntry) SetMessage(s string) ErrorBuilder {
	e.message = s
	return e
}

/*
Sets the error data for this error.
Use this to add more information to the error.
Like for example, user id or email address.
To customize the way the data is marshalled, implement zapcore.ObjectMarshaler interface
for the object you want to pass in.
tower.Fields type can be used to easily add context to the error.
Example:

	tower.Wrap(err).Code(400).SetContext(tower.Fields{"foo": "bar"})
*/
func (e *errorEntry) SetContext(ctx any) ErrorBuilder {
	e.data = append(e.data, ctx)
	return e
}

/*
Sets the key for this error. This is how Messenger will use to identify if an error is the same as previous or not.
By default, the key is the file:line from where the error is created.
Tower Logger operation will ignore the key completely.
No matter the circumstances, Tower will always try to log the error when the .Log(ctx) method is invoked.
The key however, may be used by Messenger to determine how they will treat the error.
*/
func (e *errorEntry) SetKey(key string) ErrorBuilder {
	e.key = key
	return e
}

/*
Signals the Tower library that this error should be logged.
You should call this method after calling the Set methods, after you have set all the other values for the error.
Because they need to be set before the error is logged.
*/
func (e *errorEntry) Log(ctx context.Context, opts ...zap.Option) ErrorEntry {
	return e
}

/*
Signals the Tower library that this error should be send to Messengers.
You should call this method after calling the Set methods, after you have set all the other values for the error.
Because they need to be set before the error is send to messengers.
*/
func (e *errorEntry) Notify(ctx context.Context, opts ...MessageOption) ErrorEntry {
	return e
}
