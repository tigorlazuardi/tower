package slackrest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tigorlazuardi/tower"
	"io"
	"net/http"
	"sync"

	"github.com/francoispqt/gojay"
	"github.com/tigorlazuardi/tower-go/towerslack/block"
)

func init() {
	MessagePayloadPool = &sync.Pool{
		New: func() any {
			return &MessagePayload{}
		},
	}
}

var MessagePayloadPool *sync.Pool

type MessageParseOption string

// See https://api.slack.com/methods/chat.postMessage for details.
type MessagePayload struct {
	Channel        string                      `json:"channel"`
	Text           string                      `json:"text"`
	Blocks         block.Blocks                `json:"blocks"`
	Attachments    []gojay.MarshalerJSONObject `json:"attachments"`
	AsUser         bool                        `json:"as_user"`
	IconEmoji      string                      `json:"icon_emoji"`
	IconURL        string                      `json:"icon_url"`
	LinkNames      bool                        `json:"link_names"`
	Metadata       gojay.MarshalerJSONArray    `json:"metadata"`
	Mrkdwn         bool                        `json:"mrkdwn"`
	Parse          MessageParseOption          `json:"parse"`
	ReplyBroadcast bool                        `json:"reply_broadcast"`
	ThreadTS       string                      `json:"thread_ts"`
	UnfurlLinks    bool                        `json:"unfurl_links"`
	UnfurlMedia    bool                        `json:"unfurl_media"`
	Username       string                      `json:"username"`
}

func (m *MessagePayload) Reset() {
	m.Text = ""
	if m.Blocks != nil {
		m.Blocks = m.Blocks[:0]
	}
	if m.Attachments != nil {
		m.Attachments = m.Attachments[:0]
	}
	m.AsUser = false
	m.IconEmoji = ""
	m.IconURL = ""
	m.LinkNames = false
	m.Metadata = nil
	m.Mrkdwn = false
	m.Parse = ""
	m.ReplyBroadcast = false
	m.ThreadTS = ""
	m.UnfurlLinks = false
	m.UnfurlMedia = false
	m.Username = ""
	m.Channel = ""
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

type MessageResponse struct {
	Ok      bool                `json:"ok"`
	Channel string              `json:"channel"`
	Ts      string              `json:"ts"`
	Message MessageResponseItem `json:"message"`
}

func (i *MessageResponse) Release() {
	i.Reset()
	MessageResponsePool.Put(i)
}

type MessageResponseItem struct {
	Text        string          `json:"text"`
	Username    string          `json:"username"`
	BotID       string          `json:"bot_id"`
	Attachments json.RawMessage `json:"attachments"`
	Type        string          `json:"type"`
	Subtype     string          `json:"subtype"`
	Ts          string          `json:"ts"`
}

// Posts Message to Slack. This is not for messages with file attachments.
func PostMessage(ctx context.Context, client Client, token string, payload *MessagePayload) (resp *MessageResponse, err error) {
	buf := &bytes.Buffer{}
	enc := gojay.BorrowEncoder(buf)
	defer enc.Release()
	_ = enc.EncodeObject(payload)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://slack.com/api/chat.postMessage", buf)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	res, err := client.Do(req)
	if err != nil {
		return resp, fmt.Errorf("failed to receive response from slack: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			_ = tower.WrapFreeze(err, "failed to close response body").Log(ctx)
		}
	}(res.Body)
	if res.StatusCode >= 400 {
		errResp := &ErrorResponse{}
		err = gojay.NewDecoder(res.Body).DecodeObject(errResp)
		if err != nil {
			return resp, fmt.Errorf("failed to unmarshal json response body from slack: %w", err)
		}
		return resp, errResp
	}
	resp = MessageResponsePool.Get().(*MessageResponse) //nolint

	return
}
