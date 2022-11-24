package towerdiscord

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tigorlazuardi/tower"
)

func (d Discord) send(ctx context.Context, msg tower.MessageContext) {
	key := d.buildKey(msg)
	ticker := time.NewTicker(time.Millisecond * 300)
	for d.cache.Exist(ctx, d.globalKey) {
		<-ticker.C
	}
	if err := d.cache.Set(ctx, d.globalKey, []byte("locked"), time.Second*30); err != nil {
		_ = msg.Tower().Wrap(err).Message("%s: failed to set global lock to cache", d.Name()).Log(ctx)
	}
	if msg.SkipVerification() {
		_ = d.postMessage(ctx, msg)
		return
	}
	if d.cache.Exist(ctx, key) {
		d.cache.Delete(ctx, d.globalKey)
		return
	}
	ticker.Stop()
	iterKey := key + d.cache.Separator() + "iter"
	iter := d.getAndSetIter(ctx, iterKey)
	err := d.postMessage(ctx, msg)
	if err == nil {
		message := msg.Message()
		if msg.Err() != nil {
			message = msg.Err().Error()
		}
		if err := d.cache.Set(ctx, key, []byte(message), d.countCooldown(msg, iter)); err != nil {
			_ = msg.Tower().
				Wrap(err).
				Message("%s: failed to set message key to cache", d.Name()).
				Context(tower.F{"key": key, "payload": message}).
				Log(ctx)
		}
	}
}

func (d Discord) postMessage(ctx context.Context, msg tower.MessageContext) error {
	id := d.snowflake.Generate()
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
		ThreadID: id,
		Content:  intro,
		Username: fmt.Sprintf("%s Bot", service.Name),
		Embeds:   embeds,
	}

	if d.bucket != nil && len(files) > 0 {
		results := d.bucket.Upload(ctx, files)
		_ = results
	}

	_ = PostWebhook(ctx, d.webhook, payload)
	return nil
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

func (d Discord) countCooldown(msg tower.MessageContext, iter int) time.Duration {
	mult := (iter * iter) >> 1
	if mult < 1 {
		mult = 1
	}
	cooldown := msg.Cooldown()
	if cooldown == 0 {
		cooldown = d.cooldown
	}
	cooldown *= time.Duration(mult)
	if cooldown > time.Hour*24 {
		cooldown = time.Hour * 24
	}
	return d.cooldown * time.Duration(mult)
}

func (d Discord) getAndSetIter(ctx context.Context, key string) int {
	var iter int
	iterByte, err := d.cache.Get(ctx, key)
	if err == nil {
		iter, _ = strconv.Atoi(string(iterByte))
	}
	iter += 1
	iterByte = []byte(strconv.Itoa(iter))
	nextCooldown := d.cooldown*time.Duration(iter) + d.cooldown
	_ = d.cache.Set(ctx, key, iterByte, nextCooldown)
	return iter
}