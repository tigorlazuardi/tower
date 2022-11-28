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
	id := d.snowflake.Generate()
	extra := &ExtraInformation{CacheKey: key, ThreadID: id}
	ticker.Stop()
	if err := d.cache.Set(ctx, d.globalKey, []byte("locked"), time.Second*30); err != nil {
		_ = msg.Tower().Wrap(err).Caller(msg.Caller()).Message("%s: failed to set global lock to cache", d.Name()).Log(ctx)
	}
	if msg.SkipVerification() {
		extra.CooldownTimeEnds = time.Now().Add(time.Second * 2)
		_ = d.postMessage(ctx, msg, extra)
		d.deleteGlobalCacheKeyAfter2Seconds(ctx)
		return
	}
	if d.cache.Exist(ctx, key) {
		d.cache.Delete(ctx, d.globalKey)
		return
	}
	defer d.deleteGlobalCacheKeyAfter2Seconds(ctx)
	iterKey := key + d.cache.Separator() + "iter"
	iter := d.getAndSetIter(ctx, iterKey)
	cooldown := d.countCooldown(msg, iter)
	extra.Iteration = iter
	extra.CooldownTimeEnds = time.Now().Add(cooldown)
	err := d.postMessage(ctx, msg, extra)
	if err == nil {
		message := msg.Message()
		if msg.Err() != nil {
			message = msg.Err().Error()
		}
		if err := d.cache.Set(ctx, key, []byte(message), d.countCooldown(msg, iter)); err != nil {
			_ = msg.Tower().
				Wrap(err).
				Message("%s: failed to set message key to cache", d.Name()).
				Caller(msg.Caller()).
				Context(tower.F{"key": key, "payload": message}).
				Log(ctx)
		}
	}
}

func (d Discord) deleteGlobalCacheKeyAfter2Seconds(ctx context.Context) {
	time.Sleep(time.Second * 2)
	d.cache.Delete(ctx, d.globalKey)
}

func (d Discord) postMessage(ctx context.Context, msg tower.MessageContext, extra *ExtraInformation) error {
	var intro string
	service := msg.Service()
	err := msg.Err()
	if err != nil {
		intro = fmt.Sprintf("@here an error has occurred on service **%s** of type **%s** on environment **%s**", service.Name, service.Type, service.Environment)
	} else {
		intro = fmt.Sprintf("@here message from service **%s** of type **%s** on environment **%s**", service.Name, service.Type, service.Environment)
	}

	if extra.ThreadID == 0 {
		extra.ThreadID = d.snowflake.Generate()
	}

	embeds, files := d.builder.BuildEmbed(ctx, msg, extra)
	payload := &WebhookPayload{
		Wait:     true,
		ThreadID: extra.ThreadID,
		Content:  intro,
		Embeds:   embeds,
	}

	webhookContext := &WebhookContext{
		Message: msg,
		Files:   files,
		Payload: payload,
		Extra:   extra,
	}

	switch {
	case d.bucket != nil && len(files) > 0:
		payload, errUpload := d.bucketUpload(ctx, webhookContext)
		webhookContext.Payload = payload
		err := d.PostWebhookJSON(ctx, webhookContext)
		switch {
		case err != nil:
			return err
		case errUpload != nil:
			return errUpload
		default:
			return nil
		}
	case len(files) > 0:
		return d.PostWebhookMultipart(ctx, webhookContext)
	}
	return d.PostWebhookJSON(ctx, webhookContext)
}

func (d Discord) bucketUpload(ctx context.Context, web *WebhookContext) (*WebhookPayload, error) {
	ctx = d.hook.PreBucketUploadHook(ctx, web)
	results := d.bucket.Upload(ctx, web.Files)
	d.hook.PostBucketUploadHook(ctx, web, results)
	payload := web.Payload
	errs := make([]error, 0, len(results))
	for i, result := range results {
		if result.Error != nil {
			errs = append(errs, result.Error)
			continue
		}
		var height, width int
		if imgHint, ok := result.File.Data().(ImageSizeHint); ok {
			height, width = imgHint.ImageSize()
		}
		payload.Attachments = append(payload.Attachments, &Attachment{
			ID:          i,
			Filename:    result.File.Filename(),
			Description: result.File.Pretext(),
			ContentType: result.File.ContentType(),
			Size:        result.File.Size(),
			URL:         result.URL,
			Height:      height,
			Width:       width,
		})
	}
	if len(errs) > 0 {
		return payload, tower.
			Bail("failed to upload some file(s) to bucket").
			Caller(web.Message.Caller()).
			Context(tower.F{"errors": errs}).
			Freeze()
	}
	return payload, nil
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
	multiplier := (iter * iter) >> 1
	if multiplier < 1 {
		multiplier = 1
	}
	cooldown := msg.Cooldown()
	if cooldown == 0 {
		cooldown = d.cooldown
	}
	cooldown *= time.Duration(multiplier)
	if cooldown > time.Hour*24 {
		cooldown = time.Hour * 24
	}
	return cooldown
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
