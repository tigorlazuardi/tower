package towerdiscord

import (
	"context"
	"github.com/tigorlazuardi/tower/bucket"
	"io"

	"github.com/bwmarrin/snowflake"
)

type WebhookPayload struct {
	Wait            bool
	ThreadID        snowflake.ID
	Content         string
	Username        string
	AvatarURL       string
	TTS             bool
	Embeds          []*Embed
	Files           []*File
	AllowedMentions *AllowedMentions
	PayloadJSON     string
	Attachments     []*Attachment
}

type File struct {
	Name        string
	ContentType string
	Reader      io.Reader
}

type AllowedMentions struct {
	Parse []string
	Roles []snowflake.ID
	Users []snowflake.ID
}

type Attachment struct {
	ID          snowflake.ID
	Filename    string
	Description string
	ContentType string
	Size        int
	URL         string
	ProxyURL    string
	Height      int
	Width       int
	Ephemeral   bool
}

func (d Discord) PostWebhook(ctx context.Context, payload *WebhookPayload) error {
	panic("not implemented") // TODO: Implement
	return nil
}

func (d Discord) PostWebhookWithFiles(ctx context.Context, payload *WebhookPayload, files []*bucket.File) error {
	panic("not implemented") // TODO: Implement
	return nil
}
