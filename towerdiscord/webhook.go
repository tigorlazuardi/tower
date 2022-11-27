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
	"mime/multipart"
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
	AllowedMentions *AllowedMentions `json:"allowed_mentions,omitempty"`
	Attachments     []*Attachment    `json:"attachments,omitempty"`
}

func (w *WebhookPayload) BuildMultipartPayloadJSON() ([]byte, error) {
	fields := map[string]any{}
	if w.Content != "" {
		fields["content"] = w.Content
	}
	if w.Username != "" {
		fields["username"] = w.Username
	}
	if w.AvatarURL != "" {
		fields["avatar_url"] = w.AvatarURL
	}
	if w.TTS {
		fields["tts"] = w.TTS
	}
	if len(w.Embeds) > 0 {
		fields["embeds"] = w.Embeds
	}
	if len(w.Attachments) > 0 {
		fields["attachments"] = w.Attachments
	}
	if w.AllowedMentions != nil {
		fields["allowed_mentions"] = w.AllowedMentions
	}
	return json.Marshal(fields)
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
}

type AllowedMentions struct {
	Parse []string
	Roles []snowflake.ID
	Users []snowflake.ID
}

type Attachment struct {
	ID          snowflake.ID `json:"id,omitempty"`
	Filename    string       `json:"filename,omitempty"`
	Description string       `json:"description,omitempty"`
	ContentType string       `json:"content_type,omitempty"`
	Size        int          `json:"size,omitempty"`
	URL         string       `json:"url,omitempty"`
	ProxyURL    string       `json:"proxy_url,omitempty"`
	Height      int          `json:"height,omitempty"`
	Width       int          `json:"width,omitempty"`
	Ephemeral   bool         `json:"ephemeral,omitempty"`
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.webhook, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	if web.Payload.Wait {
		req.URL.Query().Add("wait", "true")
	}
	req.URL.Query().Add("thread_id", web.Payload.ThreadID.String())
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute webhook: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
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

func (d Discord) PostWebhookMultipart(ctx context.Context, web *WebhookContext) error {
	ctx = d.hook.PreMessageHook(ctx, web)
	requestBody, contentType := d.buildMultipartWebhookBody(ctx, web)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.webhook, requestBody)
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	if web.Payload.Wait {
		req.URL.Query().Add("wait", "true")
	}
	req.URL.Query().Add("thread_id", web.Payload.ThreadID.String())
	req.Header.Set("Content-Type", contentType)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute webhook: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
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

func (d Discord) buildMultipartWebhookBody(ctx context.Context, web *WebhookContext) (body io.Reader, contentType string) {
	reader, writer := io.Pipe()
	multipartWriter := multipart.NewWriter(writer)
	contentType = multipartWriter.FormDataContentType()
	go func() {
		payload := web.Payload
		var err error
		defer func() {
			if err != nil {
				err = web.Message.Tower().Wrap(err).
					Message("%s: failed to build multipart/form-data webhook payload", d.Name()).
					Caller(web.Message.Caller()).
					Freeze()
				_ = writer.CloseWithError(err)
			} else {
				_ = writer.Close()
			}
		}()
		defer func(w *multipart.Writer) {
			if errClose := w.Close(); errClose != nil {
				err = errClose
			}
		}(multipartWriter)

		id := d.snowflake.Generate()
		filename := id.String() + ".md"
		fw, err := multipartWriter.CreateFormFile("files[0]", filename)
		if err != nil {
			return
		}

		buf := &bytes.Buffer{}
		for i, file := range web.Files {
			if i > 0 {
				buf.WriteString("\n\n")
			}
			buf.WriteString(file.Pretext())
			buf.WriteString("\n\n")
			_, err = io.Copy(buf, file.Data())
			if err != nil {
				return
			}
		}
		_, err = io.Copy(fw, buf)
		if err != nil {
			return
		}

		payload.Attachments = append(payload.Attachments, &Attachment{
			ID:          id,
			Filename:    filename,
			Description: "File segments",
			ContentType: "text/markdown",
			Size:        buf.Len(),
		})

		fw, err = multipartWriter.CreateFormField("payload_json")
		if err != nil {
			return
		}
		js, err := payload.BuildMultipartPayloadJSON()
		if err != nil {
			return
		}
		_, err = io.Copy(fw, bytes.NewReader(js))
	}()
	return reader, contentType
}
