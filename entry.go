package tower

import (
	"context"
	"strings"

	"go.uber.org/zap/zapcore"
)

type entry struct {
	tower   *Tower
	code    int
	message string
	err     error
	context []zapcore.ObjectMarshaler
	key     string
}

// Writes the error.Error to the writer instead of being allocated as value.
func (e entry) WriteError(w Writer) {
	if e.err == nil && e.message == "" {
		w.WriteString("[nil]")
		return
	} else if e.err == nil {
		w.WriteString(e.message)
		w.WriteString(" => [nil]")
		return
	}
	var innerMessage string
	errMsg := e.err.Error()
	msgHint, ok := e.err.(MessageHinter)
	if ok {
		innerMessage = msgHint.MessageHint()
	}
	if e.message != innerMessage && e.message != errMsg && e.message != "" {
		w.WriteString(e.message)
		w.WriteString(" => ")
	}
	if ew, ok := e.err.(ErrorWriter); ok {
		ew.WriteError(w)
		return
	}
	w.WriteString(errMsg)
}

func (e entry) Error() string {
	s := strings.Builder{}
	e.WriteError(&s)
	return s.String()
}

func (e *entry) SetCode(i int) ErrorBuilder {
	e.code = i
	return e
}

func (e *entry) SetMessage(s string) ErrorBuilder {
	e.message = s
	return e
}

func (e *entry) SetContext(ctx zapcore.ObjectMarshaler) ErrorBuilder {
	e.context = append(e.context, ctx)
	return e
}

func (e *entry) SetKey(key string) ErrorBuilder {
	e.key = key
	return e
}

/*
Signals the Tower library that this error should be logged.

You should call this method after calling the Set methods, after you have set all the other values for the error.
Because they need to be set before the error is logged.
*/
func (e entry) LogError(ctx context.Context) ErrorBuilder {
	panic("not implemented") // TODO: Implement
}

/*
Signals the Tower library that this error should be send to Messengers.

You should call this method after calling the Set methods, after you have set all the other values for the error.
Because they need to be set before the error is send to messengers.
*/
func (e entry) NotifyError(ctx context.Context) ErrorBuilder {
	panic("not implemented") // TODO: Implement
}
