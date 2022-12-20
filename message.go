package tower

import (
	"strings"
	"time"
)

// MessageContext is the context of a message.
//
// It holds the message and data that can be sent to the Messenger's target.
type MessageContext interface {
	HTTPCodeHint
	CodeHint
	MessageHint
	CallerHint
	KeyHint
	LevelHint
	ServiceHint
	ContextHint
	TimeHint
	// Err returns the error item. May be nil if message contains no error.
	Err() error
	// SkipVerification If true, Sender asks for this message to always be send.
	SkipVerification() bool
	// Cooldown returns non-zero value if Sender asks for this message to be sent after this duration.
	Cooldown() time.Duration
	// Tower Gets the tower instance that created this MessageContext.
	Tower() *Tower
}

type MessageOption interface {
	apply(*messageOption)
}

type MessageParameter interface {
	SkipVerification() bool
	Cooldown() time.Duration
	Tower() *Tower
}

type messageOption struct {
	skipVerification  bool
	specificMessenger Messenger
	messengers        Messengers
	cooldown          time.Duration
	tower             *Tower
}

func (o messageOption) SkipVerification() bool {
	return o.skipVerification
}

func (o messageOption) SpecificMessenger() Messenger {
	return o.specificMessenger
}

func (o messageOption) Messengers() Messengers {
	return o.messengers
}

func (o messageOption) Cooldown() time.Duration {
	return o.cooldown
}

func (o messageOption) Tower() *Tower {
	return o.tower
}

type MessageOptionFunc func(*messageOption)

func (f MessageOptionFunc) apply(opt *messageOption) {
	f(opt)
}

// SkipMessageVerification Asks the Messengers to Skip cooldown verifications and just send the message.
func SkipMessageVerification(b bool) MessageOption {
	return MessageOptionFunc(func(ob *messageOption) {
		ob.skipVerification = b
	})
}

/*
OnlyMessengerWithName Asks Tower to only send only to the Messenger with this name.
If name is not found, Tower returns to default behaviour.

Note: OnlyMessengerWithName messageOption will conflict with other Messenger setters messageOption, and thus only the latest messageOption will be set.
*/
func OnlyMessengerWithName(name string) MessageOption {
	return MessageOptionFunc(func(ob *messageOption) {
		ob.specificMessenger = ob.tower.GetMessengerByName(name)
		ob.messengers = nil
	})
}

/*
Asks Tower to only send only to this Messenger.

Note: OnlyThisMessenger messageOption will conflict with other Messenger setters messageOption, and thus only the latest messageOption will be set.
*/
func OnlyThisMessenger(m Messenger) MessageOption {
	return MessageOptionFunc(func(ob *messageOption) {
		ob.messengers = nil
		ob.specificMessenger = m
	})
}

func OnlyTheseMessengers(m ...Messenger) MessageOption {
	return MessageOptionFunc(func(ob *messageOption) {
		ob.specificMessenger = nil
		mm := make(Messengers, len(m))
		for _, v := range m {
			mm[v.Name()] = v
		}
		ob.messengers = mm
	})
}

/*
Asks Tower to only send messages to Messengers whose name begins with given s.

Note: MessengerPrefix messageOption will conflict with other Messenger setters messageOption, and thus only the latest messageOption will be set.
*/
func MessengerPrefix(s string) MessageOption {
	return MessageOptionFunc(func(ob *messageOption) {
		ob.specificMessenger = nil
		messengers := ob.tower.GetMessengers()
		mm := make(Messengers, len(messengers))
		for k, v := range messengers {
			if strings.HasPrefix(k, s) {
				mm[k] = v
			}
		}
		ob.messengers = mm
	})
}

/*
Asks Tower to only send messages to Messengers whose name ends with given s.

Note: MessengerSuffix messageOption will conflict with other Messenger setters messageOption, and thus only the latest messageOption will be set.
*/
func MessengerSuffix(s string) MessageOption {
	return MessageOptionFunc(func(ob *messageOption) {
		ob.specificMessenger = nil
		messengers := ob.tower.GetMessengers()
		mm := make(Messengers, len(messengers))
		for k, v := range messengers {
			if strings.HasSuffix(k, s) {
				mm[k] = v
			}
		}
		ob.messengers = mm
	})
}

/*
Asks Tower to only send messages to Messengers whose name contains given s.

Note: MessengerNameContains messageOption will conflict with other Messenger setters messageOption, and thus only the latest messageOption will be set.
*/
func MessengerNameContains(s string) MessageOption {
	return MessageOptionFunc(func(ob *messageOption) {
		ob.specificMessenger = nil
		messengers := ob.tower.GetMessengers()
		mm := make(Messengers, len(messengers))
		for k, v := range messengers {
			if strings.Contains(k, s) {
				mm[k] = v
			}
		}
		ob.messengers = mm
	})
}

/*
Sets the Cooldown for this Message.
*/
func MessageCooldown(dur time.Duration) MessageOption {
	return MessageOptionFunc(func(ob *messageOption) {
		ob.cooldown = dur
	})
}

/*
Asks Tower to send messages to currenty registered and also send those messeges to these Messengers.

Note: MessengerNameContains messageOption will conflict with other Messenger setters messageOption, and thus only the latest messageOption will be set.
*/
func ExtraMessengers(messengers ...Messenger) MessageOption {
	return MessageOptionFunc(func(ob *messageOption) {
		ob.specificMessenger = nil
		ob.messengers = ob.tower.GetMessengers()
		for _, v := range messengers {
			ob.messengers[v.Name()] = v
		}
	})
}
