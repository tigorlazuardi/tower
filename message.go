package tower

import (
	"context"
	"time"
)

type MessageContext interface {
	BodyCodeHint
	HTTPCodeHint
	CodeHint
	MessageHint
	CallerHint
	KeyHint
	LevelHint
	ServiceHint
	// Current Context.
	Ctx() context.Context
	// Current time.
	Time() time.Time
	// Data Text for Human readable text. Should not be used for structured data. May be nil.
	DataDisplay() DisplayWriter
	// Human readable Error Text. Should not be used for structured data. May be nil.
	ErrorDisplay() DisplayWriter
	// Human readable Summary. Should not be used for structured data. May be nil.
	SummaryDisplay() DisplayWriter
	// data Object
	Data() interface{}
	// Error item. May be nil if message contains no error.
	Err() error
	// If true, Sender asks for this message to always be send.
	SkipVerification() bool
}

type MessageOption interface {
	apply(*option)
}

type option struct {
	skipVerification  bool
	specificMessenger Messenger
	messengers        []Messenger
}

type messageOptionFunc func(*option)

func (f messageOptionFunc) apply(opt *option) {
	f(opt)
}
