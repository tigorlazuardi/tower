package tower

type Messenger interface {
	// Returns the name of the Messenger.
	Name() string
	// Sends notification.
	SendMessage(ctx MessageContext) error

	// Waits until all message in the queue or until given channel is received.
	//
	// Implementer must exit the function as soon as possible when this cancel receives the item.
	Wait(cancel <-chan struct{}) error
}
