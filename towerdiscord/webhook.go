package towerdiscord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower/bucket"
	"io"
	"net/http"
)

type WebhookPayload struct {
	Wait            bool             `json:"-"`
	ThreadID        snowflake.ID     `json:"-"`
	Content         string           `json:"content,omitempty"`
	Username        string           `json:"username,omitempty"`
	AvatarURL       string           `json:"avatarURL,omitempty"`
	TTS             bool             `json:"TTS,omitempty"`
	Embeds          []*Embed         `json:"embeds,omitempty"`
	Files           []*WebhookFile   `json:"files,omitempty"`
	AllowedMentions *AllowedMentions `json:"allowedMentions,omitempty"`
	PayloadJSON     string           `json:"payloadJSON,omitempty"`
	Attachments     []*Attachment    `json:"attachments,omitempty"`
}

type DiscordErrorResponse struct {
	Code       int             `json:"code"`
	Message    string          `json:"message"`
	StatusCode int             `json:"status_code"`
	Raw        json.RawMessage `json:"raw"`
}

func newDiscordErrorResponse(statusCode int, body []byte) (*DiscordErrorResponse, error) {
	var errResp DiscordErrorResponse
	errResp.StatusCode = statusCode
	if err := json.Unmarshal(body, &errResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal discord error response: %w", err)
	}
	errResp.Raw = body
	return &errResp, nil
}

func (d DiscordErrorResponse) PrintJSON() {
	f, _ := json.Marshal(d)
	fmt.Println(string(f))
}

func (d DiscordErrorResponse) String() string {
	return fmt.Sprintf("discord error: [%d] %s", d.Code, d.Message)
}

func (d DiscordErrorResponse) Error() string {
	return d.String()
}

type WebhookFile struct {
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

type WebhookContext struct {
	Message tower.MessageContext
	Files   []bucket.File
	Payload *WebhookPayload
	Extra   *ExtraInformation
}

func (d Discord) PostWebhookJSON(ctx context.Context, web *WebhookContext) error {
	ctx = d.hook.PreMessageHook(ctx, web)
	b, err := json.Marshal(web.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, d.webhook, bytes.NewReader(b))
	if web.Payload.Wait {
		req.URL.Query().Add("wait", "true")
	}
	req.URL.Query().Add("thread_id", web.Payload.ThreadID.String())
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute webhook: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.hook.PostMessageHook(ctx, web, err)
		return fmt.Errorf("failed to read webhook response body: %w", err)
	}
	if resp.StatusCode >= 400 {
		errResp, err := newDiscordErrorResponse(resp.StatusCode, body)
		if err != nil {
			d.hook.PostMessageHook(ctx, web, err)
			return fmt.Errorf("failed to parse discord error response: %w", err)
		}
		errResp.PrintJSON()
		d.hook.PostMessageHook(ctx, web, errResp)
		return errResp
	}
	d.hook.PostMessageHook(ctx, web, err)

	return nil
}

func (d Discord) PostWebhookMultipart(ctx context.Context, payload *WebhookPayload, files []bucket.File) error {
	panic("not implemented") // TODO: Implement
	return nil
}
