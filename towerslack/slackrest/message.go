package slackrest

import (
	"context"

	"github.com/francoispqt/gojay"
	"github.com/tigorlazuardi/tower-go/towerslack/block"
)

type MessageParseOption string

const (
	MessageParseNone MessageParseOption = "none"
	MessageParseFull MessageParseOption = "full"
)

// See https://api.slack.com/methods/chat.postMessage for details.
type MessagePayload struct {
	Text           string
	Blocks         block.Blocks
	Attachments    []gojay.MarshalerJSONObject
	AsUser         bool
	IconEmoji      string
	IconURL        string
	LinkNames      bool
	Metadata       gojay.MarshalerJSONArray
	Mrkdwn         bool
	Parse          MessageParseOption
	ReplyBroadcast bool
	ThreadTS       string
	UnfurlLinks    bool
	UnfurlMedia    bool
	Username       string
}

func (m MessagePayload) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddStringKeyOmitEmpty("text", m.Text)
	enc.AddArrayKeyOmitEmpty("blocks", m.Blocks)
	enc.AddArrayKeyOmitEmpty("attachments", gojay.EncodeArrayFunc(func(e *gojay.Encoder) {
		for _, v := range m.Attachments {
			e.AddObject(v)
		}
	}))
	enc.AddBoolKeyOmitEmpty("as_user", m.AsUser)
	enc.AddStringKeyOmitEmpty("icon_emoji", m.IconEmoji)
	enc.AddStringKeyOmitEmpty("icon_url", m.IconURL)
	enc.AddBoolKeyOmitEmpty("link_names", m.LinkNames)
	enc.AddArrayKeyOmitEmpty("metadata", m.Metadata)
	enc.AddBoolKeyOmitEmpty("mrkdwn", m.Mrkdwn)
	enc.AddStringKeyOmitEmpty("parse", string(m.Parse))
	enc.AddBoolKeyOmitEmpty("reply_broadcast", m.ReplyBroadcast)
	enc.AddStringKeyOmitEmpty("thread_ts", m.ThreadTS)
	enc.AddBoolKeyOmitEmpty("unfurl_links", m.UnfurlLinks)
	enc.AddBoolKeyOmitEmpty("unfurl_media", m.UnfurlMedia)
	enc.AddStringKeyOmitEmpty("username", m.Username)
}

func (m MessagePayload) IsNil() bool {
	return len(m.Text) == 0 && len(m.Blocks) == 0 && len(m.Attachments) == 0
}

func PostMessage(ctx context.Context, payload MessagePayload) {}
