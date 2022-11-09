package tower

import "time"

type BodyCodeHint interface {
	// Gets the Body Code for the type.
	BodyCode() int
}

type HTTPCodeHint interface {
	// Gets HTTP Status Code for the type.
	HTTPCode() int
}

type CodeHint interface {
	// Gets the original code of the type.
	Code() int
}

type CallerHint interface {
	// Gets the caller of this type.
	Caller() Caller
}

type MessageHint interface {
	// Gets the Message of the type.
	Message() string
}

type KeyHint interface {
	// Gets the key for this message.
	Key() string
}

type ContextHint interface {
	// Gets the context of this this type.
	Context() []any
}

type ServiceHint interface {
	// Gets the service information.
	Service() Service
}

type LevelHint interface {
	// Gets the level of this message.
	Level() Level
}

type TimeHint interface {
	Time() time.Time
}
