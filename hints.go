package tower

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"go.uber.org/zap/zapcore"
)

const sep = string(os.PathSeparator)

type Caller struct {
	PC   uintptr
	File string
	Line int
}

func (c Caller) Function() *runtime.Func {
	f := runtime.FuncForPC(c.PC)
	return f
}

func (c Caller) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("source", fmt.Sprintf("%s:%d", c.File, c.Line))
	enc.AddString("origin", c.ShortOrigin())
	return nil
}

func (c Caller) getOrigin() []string {
	f := runtime.FuncForPC(c.PC)
	return strings.Split(f.Name(), "/")
}

func (c Caller) ShortOrigin() string {
	s := c.getOrigin()

	for len(s) > 4 {
		s = s[1:]
	}

	return strings.Join(s, "/")
}

func GetCaller(depth int) (Caller, bool) {
	pc, file, line, ok := runtime.Caller(depth)
	if !ok {
		return Caller{}, false
	}

	return Caller{
		PC:   pc,
		File: file,
		Line: line,
	}, true
}

type CodeHinter interface {
	// Gets the error code for this error.
	CodeHint() int
}

type MessageHinter interface {
	// Gets the error message for this error.
	MessageHint() string
}

type CallHinter interface {
	// Gets the origin location where this error is created.
	CallHint() Caller
}

type ContextHinter interface {
	// Gets the context for this error.
	ContextHint() []zapcore.ObjectMarshaler
}

type HTTPCodeHinter interface {
	// Gets the HTTP status code for this error.
	HTTPCodeHint() int
}

type BodyCodeHinter interface {
	// Gets the body code for this error.
	BodyCodeHint() int
}

type KeyHinter interface {
	// Gets the hint for the message key.
	KeyHint() string
}
