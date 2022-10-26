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
	Apply(MessageOptionBuilder)
}

type MessageOptionBuilder interface {
	// Sender Asks the messages to be send, ignoring any delays and cooldowns.
	SetSkipMessageVerification(b bool)
	// Also send Message to these Messengers. Implementers should extends tower's already registered Messengers.
	ExtraMessengers(...Messenger)
	// Senders Asks only to send to Messenger with this name. If found, SpecificMessenger must return this Messenger, otherwise that Error return nil.
	OnlyMessengerWithName(name string)
	// Sender asks to send very specifically to this Messenger.
	OnlyThisMessenger(m Messenger)
	// Sender asks to send very specifically to these Messenger.
	OnlyTheseMessengers(...Messenger)
	// Only sends to Messenger with the following prefix in its name.
	MessengerPrefix(prefix string)
	// Only sends to Messenger with the following suffix.
	MessengerSuffix(suffix string)
	// Only sends to Messenger that contains the following string.
	MessengerNameContains(contains string)
	// Sender asks the cooldown for this message to be this duration.
	MessageCooldown(time.Duration)
}

type MessageParameter interface {
	SkipVerification() bool
	// When Tower sees that this Value is not nil, Tower will only triggers this Messenger.
	SpecificMessenger() Messenger
	// When Tower sees the returned value is higher than 0, Tower will only send Messages to these Messengers.
	Messengers() Messengers
	// Returns 0 if sender does not set any custom base cooldown.
	Cooldown() time.Duration
}

type OptionGenerator interface {
	GenerateOption(t *Tower, opts ...MessageOption) MessageParameter
}

var _ OptionGenerator = (OptionGeneratorFunc)(nil)

type OptionGeneratorFunc func(t *Tower, opts ...MessageOption) MessageParameter

func (o OptionGeneratorFunc) GenerateOption(t *Tower, opts ...MessageOption) MessageParameter {
	return o(t, opts...)
}

func generateOption(t *Tower, opts ...MessageOption) MessageParameter {
	o := &option{tower: t}
	for _, v := range opts {
		v.Apply(o)
	}
	return o
}

var (
	_ MessageParameter     = (*option)(nil)
	_ MessageOptionBuilder = (*option)(nil)
)

type option struct {
	skipVerification  bool
	specificMessenger Messenger
	messengers        Messengers
	cooldown          time.Duration
	tower             *Tower
}

// Sender Asks the messages to be send, ignoring any delays and cooldowns.
func (o *option) SetSkipMessageVerification(b bool) {
	o.skipVerification = b
}

// Senders Asks only to send to Messenger with this name. If found, SpecificMessenger must return this Messenger, otherwise that Error return nil.
func (o *option) OnlyMessengerWithName(name string) {
	o.specificMessenger = o.tower.GetMessengerByName(name)
	o.messengers = nil
}

// Sender asks to send very specifically to these Messenger.
func (o *option) OnlyTheseMessengers(m ...Messenger) {
	o.specificMessenger = nil
	mm := make(Messengers, len(m))
	for _, v := range m {
		mm[v.Name()] = v
	}
	o.messengers = mm
}

// Only sends to Messenger with the following prefix in its name.
func (o *option) MessengerPrefix(prefix string) {
	o.specificMessenger = nil
	messengers := o.tower.GetMessengers()
	mm := make(Messengers, len(messengers))
	for k, v := range messengers {
		if strings.HasPrefix(k, prefix) {
			mm[k] = v
		}
	}
	o.messengers = mm
}

// Only sends to Messenger with the following suffix.
func (o *option) MessengerSuffix(suffix string) {
	o.specificMessenger = nil
	messengers := o.tower.GetMessengers()
	mm := make(Messengers, len(messengers))
	for k, v := range messengers {
		if strings.HasSuffix(k, suffix) {
			mm[k] = v
		}
	}
	o.messengers = mm
}

// Only sends to Messenger that contains the following string.
func (o *option) MessengerNameContains(contains string) {
	o.specificMessenger = nil
	messengers := o.tower.GetMessengers()
	mm := make(Messengers, len(messengers))
	for k, v := range messengers {
		if strings.Contains(k, contains) {
			mm[k] = v
		}
	}
	o.messengers = mm
}

// Sender asks the cooldown for this message to be this duration.
func (o *option) MessageCooldown(dur time.Duration) {
	o.cooldown = dur
}

// Sender asks to send very specifically to this Messenger.
func (o *option) OnlyThisMessenger(m Messenger) {
	o.specificMessenger = m
	o.messengers = nil
}

// Also send Message to these Messengers.
func (o *option) ExtraMessengers(m ...Messenger) {
	o.specificMessenger = nil
	o.messengers = o.tower.GetMessengers()
	for _, v := range m {
		o.messengers[v.Name()] = v
	}
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

func (o option) Cooldown() time.Duration {
	return o.cooldown
}

type MessageOptionFunc func(MessageOptionBuilder)

func (f MessageOptionFunc) Apply(opt MessageOptionBuilder) {
	f(opt)
}

// Asks the Messengers to Skip cooldown verifications and just send the message.
func SkipMessageVerification(b bool) MessageOption {
	return MessageOptionFunc(func(ob MessageOptionBuilder) {
		ob.SetSkipMessageVerification(b)
	})
}

/*
Asks Tower to only send only to the Messenger with this name.
If name is not found, Tower returns to default behaviour.

Note: OnlyMessengerWithName option will conflict with other Messenger setters option, and thus only the latest option will be set.
*/
func OnlyMessengerWithName(name string) MessageOption {
	return MessageOptionFunc(func(ob MessageOptionBuilder) {
		ob.OnlyMessengerWithName(name)
	})
}

/*
Asks Tower to only send only to this Messenger.

Note: OnlyThisMessenger option will conflict with other Messenger setters option, and thus only the latest option will be set.
*/
func OnlyThisMessenger(m Messenger) MessageOption {
	return MessageOptionFunc(func(ob MessageOptionBuilder) {
		ob.OnlyThisMessenger(m)
	})
}

/*
Asks Tower to only send messages to Messengers whose name begins with given s.

Note: MessengerPrefix option will conflict with other Messenger setters option, and thus only the latest option will be set.
*/
func MessengerPrefix(s string) MessageOption {
	return MessageOptionFunc(func(ob MessageOptionBuilder) {
		ob.MessengerPrefix(s)
	})
}

/*
Asks Tower to only send messages to Messengers whose name ends with given s.

Note: MessengerSuffix option will conflict with other Messenger setters option, and thus only the latest option will be set.
*/
func MessengerSuffix(s string) MessageOption {
	return MessageOptionFunc(func(ob MessageOptionBuilder) {
		ob.MessengerSuffix(s)
	})
}

/*
Asks Tower to only send messages to Messengers whose name contains given s.

Note: MessengerNameContains option will conflict with other Messenger setters option, and thus only the latest option will be set.
*/
func MessengerNameContains(s string) MessageOption {
	return MessageOptionFunc(func(ob MessageOptionBuilder) {
		ob.MessengerNameContains(s)
	})
}

/*
Sets the Cooldown for this Message.
*/
func MessageCooldown(dur time.Duration) MessageOption {
	return MessageOptionFunc(func(ob MessageOptionBuilder) {
		ob.MessageCooldown(dur)
	})
}

/*
Asks Tower to send messages to currenty registered and also send those messeges to these Messengers.

Note: MessengerNameContains option will conflict with other Messenger setters option, and thus only the latest option will be set.
*/
func ExtraMessengers(messengers ...Messenger) MessageOption {
	return MessageOptionFunc(func(ob MessageOptionBuilder) {
		ob.ExtraMessengers(messengers...)
	})
}
