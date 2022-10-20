package tower

import (
	"go.uber.org/zap/zapcore"
)

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
