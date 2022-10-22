package tower

import (
	"context"
	"strings"
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
	Apply(OptionBuilder)
}

type OptionBuilder interface {
	// Sender Asks the messages to be send, ignoring any delays and cooldowns.
	SetSkipVerification(b bool)
	// Sender asks the messengers to respect private marshaling.
	// If the Messenger does not support it, revert to default behaviour.
	//
	// Tower already warns the user to implement tower.PrivateMarshalJSON, so the Messenger should check for that implementation.
	SetPrivate(b bool)
	// Also send Message to these Messengers. Implementers should extends tower's already registered Messengers.
	AddMessengers(...Messenger)
	// Senders Asks only to send to Messenger with this name. If found, SpecificMessenger must return this Messenger, otherwise that Error return nil.
	Only(name string)
	// Sender asks to send very specifically to this Messenger.
	OnlyThisMessenger(m Messenger)
	// Only sends to Messenger with the following prefix in its name.
	Prefix(prefix string)
	// Only sends to Messenger with the following suffix.
	Suffix(suffix string)
	// Only sends to Messenger that contains the following string.
	Contains(contains string)
	// Sender asks the cooldown for this message to be this duration.
	SetCooldown(time.Duration)
}

type Option interface {
	SkipVerification() bool
	// When Tower sees that this Value is not nil, Tower will only triggers this Messenger.
	SpecificMessenger() Messenger
	// When Tower sees the returned value is higher than 0, Tower will only send Messages to these Messengers.
	Messengers() Messengers
	Private() bool
	Cooldown() time.Duration
}

type OptionGenerator interface {
	GenerateOption(t *Tower, opts ...MessageOption) Option
}

var _ OptionGenerator = (OptionGeneratorFunc)(nil)

type OptionGeneratorFunc func(t *Tower, opts ...MessageOption) Option

func (o OptionGeneratorFunc) GenerateOption(t *Tower, opts ...MessageOption) Option {
	return o(t, opts...)
}

func generateOption(t *Tower, opts ...MessageOption) Option {
	o := &option{tower: t}
	for _, v := range opts {
		v.Apply(o)
	}
	return o
}

var (
	_ Option        = (*option)(nil)
	_ OptionBuilder = (*option)(nil)
)

type option struct {
	skipVerification  bool
	specificMessenger Messenger
	messengers        Messengers
	private           bool
	cooldown          time.Duration
	tower             *Tower
}

// Sender Asks the messages to be send, ignoring any delays and cooldowns.
func (o *option) SetSkipVerification(b bool) {
	o.skipVerification = true
}

// Sender asks the messengers to respect private marshaling.
func (o *option) SetPrivate(b bool) {
	o.private = b
}

// Sender asks to send very specifically to this Messenger.
func (o *option) OnlyThisMessenger(m Messenger) {
	o.specificMessenger = m
}

// Only sends to Messenger with the following prefix.
func (o *option) Prefix(prefix string) {
	msg := make(Messengers, len(o.tower.messengers))
	for k, v := range o.tower.messengers {
		if strings.HasPrefix(k, prefix) {
			msg[k] = v
		}
	}
}

// Only sends to Messenger with the following suffix.
func (o *option) Suffix(suffix string) {
	msg := make(Messengers, len(o.tower.messengers))
	for k, v := range o.tower.messengers {
		if strings.HasSuffix(k, suffix) {
			msg[k] = v
		}
	}
}

// Only sends to Messenger that contains the following string.
func (o *option) Contains(contains string) {
	msg := make(Messengers, len(o.tower.messengers))
	for k, v := range o.tower.messengers {
		if strings.Contains(k, contains) {
			msg[k] = v
		}
	}
}

// Sender asks to send very specifically to this Messenger.
func (o *option) SetSpecificMessenger(m Messenger) {
	o.specificMessenger = m
}

// Also send Message to these Messengers.
func (o *option) AddMessengers(m ...Messenger) {
	for _, v := range m {
		o.messengers[v.Name()] = v
	}
}

// Senders Asks only to send to Messenger with this name.
func (o *option) Only(name string) {
	o.specificMessenger = o.tower.GetMessengerByName(name)
}

// Sender asks the cooldown for this message to be this duration.
func (o *option) SetCooldown(d time.Duration) {
	o.cooldown = d
}

func (o option) SkipVerification() bool {
	return o.skipVerification
}

func (o option) SpecificMessenger() Messenger {
	return o.specificMessenger
}

func (o option) Messengers() Messengers {
	return o.messengers
}

func (o option) Private() bool {
	return o.private
}

func (o option) Cooldown() time.Duration {
	return o.cooldown
}

type MessageOptionFunc func(OptionBuilder)

func (f MessageOptionFunc) Apply(opt OptionBuilder) {
	f(opt)
}

// Asks the Messengers to Skip cooldown verifications and just send the message.
func SkipVerification(b bool) MessageOption {
	return MessageOptionFunc(func(ob OptionBuilder) {
		ob.SetSkipVerification(b)
	})
}

func Private(b bool) MessageOption {
	return MessageOptionFunc(func(ob OptionBuilder) {
		ob.SetPrivate(b)
	})
}
