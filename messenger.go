package tower

import "context"

type Messenger interface {
	// Returns the name of the Messenger.
	Name() string
	// Sends notification.
	SendMessage(ctx context.Context, msg MessageContext)

	// Waits until all message in the queue or until given channel is received.
	//
	// Implementer must exit the function as soon as possible when this ctx is canceled.
	Wait(ctx context.Context) error
}

type Messengers map[string]Messenger

func (c Messengers) Clone() Messengers {
	clone := make(Messengers, len(c))

	for k, v := range c {
		clone[k] = v
	}

	return clone
}
