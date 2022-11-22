package towerdiscord

import (
	"context"
	"fmt"
	"github.com/tigorlazuardi/tower"
	"strings"
)

func (d Discord) send(ctx context.Context, msg tower.MessageContext) {
	var intro string
	service := msg.Service()
	err := msg.Err()
	if err != nil {
		intro = fmt.Sprintf("<!here> an error has occurred on service %s of type %s on environment %s", service.Name, service.Type, service.Environment)
	} else {
		intro = fmt.Sprintf("<!here> message from service %s of type %s on environment %s", service.Name, service.Type, service.Environment)
	}

	embeds, files := d.builder.BuildEmbed(ctx, msg)
	payload := &WebhookPayload{
		Wait:     true,
		ThreadID: 0,
		Content:  intro,
		Username: fmt.Sprintf("%s Bot", service.Name),
		Embeds:   embeds,
	}

	if d.bucket != nil && len(files) > 0 {
		results := d.bucket.Upload(ctx, files)
		_ = results
	}

	PostWebhook(ctx, d.webhook, payload)

}

func (d Discord) buildKey(msg tower.MessageContext) string {
	builder := strings.Builder{}
	builder.WriteString(d.Name())
	builder.WriteString(d.cache.Separator())
	service := msg.Service()
	builder.WriteString(service.Environment)
	builder.WriteString(d.cache.Separator())
	builder.WriteString(service.Name)
	builder.WriteString(d.cache.Separator())
	builder.WriteString(service.Type)
	builder.WriteString(d.cache.Separator())

	key := msg.Key()
	if key == "" {
		key = msg.Caller().FormatAsKey()
	}
	builder.WriteString(key)
	return builder.String()
}
