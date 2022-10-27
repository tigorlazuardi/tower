package discord

import (
	"context"

	"github.com/tigorlazuardi/tower"
)

type Discord struct {
	name  string
	token string //nolint wip
}

// Returns the name of the Messenger.
func (d Discord) Name() string {
	if d.name == "" {
		return "discord"
	}
	return d.name
}

// Sends notification.
func (d Discord) SendMessage(ctx tower.MessageContext) error {
	panic("not implemented") // TODO: Implement
}

// Waits until all message in the queue or until given channel is received.
//
// Implementer must exit the function as soon as possible when this context receivied the cancel context.
func (d Discord) Wait(ctx context.Context) error {
	return nil
}
