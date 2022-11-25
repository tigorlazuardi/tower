package towerdiscord

import (
	"context"
	"github.com/bwmarrin/snowflake"
	"github.com/tigorlazuardi/tower/bucket"
	"io"
)

type WebhookPayload struct {
	Wait            bool             `json:"-"`
	ThreadID        snowflake.ID     `json:"-"`
	Content         string           `json:"content,omitempty"`
	Username        string           `json:"username,omitempty"`
	AvatarURL       string           `json:"avatarURL,omitempty"`
	TTS             bool             `json:"TTS,omitempty"`
	Embeds          []*Embed         `json:"embeds,omitempty"`
	Files           []*File          `json:"files,omitempty"`
	AllowedMentions *AllowedMentions `json:"allowedMentions,omitempty"`
	PayloadJSON     string           `json:"payloadJSON,omitempty"`
	Attachments     []*Attachment    `json:"attachments,omitempty"`
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

func (d Discord) PostWebhookJSON(ctx context.Context, payload *WebhookPayload) error {

	panic("not implemented") // TODO: Implement
	return nil
}

func (d Discord) PostWebhookMultipart(ctx context.Context, payload *WebhookPayload, files []*bucket.File) error {
	panic("not implemented") // TODO: Implement
	return nil
}
