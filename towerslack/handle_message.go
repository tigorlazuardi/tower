package towerslack

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower/towerslack/slackrest"
)

func (s SlackBot) handleMessage(ctx context.Context, msg tower.MessageContext) {
	// TODO: Implement hooks
	key := s.buildKey(msg)

	// use tickers to account for lags.
	ticker := time.NewTicker(time.Millisecond * 300)
	for s.cache.Exist(ctx, s.globalKey) {
		<-ticker.C
	}
	ticker.Stop()
	if err := s.cache.Set(ctx, s.globalKey, []byte("locked"), time.Second*30); err != nil {
		_ = msg.Tower().Wrap(err).Message("failed to set global lock to cache").Log(ctx)
	}
	if msg.SkipVerification() {
		_ = s.postMessage(ctx, msg)
		return
	}
	if s.cache.Exist(ctx, key) {
		s.cache.Delete(ctx, s.globalKey)
		return
	}

	iterKey := key + s.cache.Separator() + "iter"
	iter := s.getAndSetIter(ctx, iterKey)
	err := s.postMessage(ctx, msg)
	if err == nil {
		message := msg.Message()
		if msg.Err() != nil {
			message = msg.Err().Error()
		}
		if err := s.cache.Set(ctx, key, []byte(message), s.countCooldown(msg, iter)); err != nil {
			_ = msg.Tower().
				Wrap(err).
				Message("failed to set message key to cache").
				Context(tower.F{"key": key, "payload": message}).
				Log(ctx)
		}
	}
}

func (s SlackBot) countCooldown(msg tower.MessageContext, iter int) time.Duration {
	mult := (iter * iter) >> 1
	if mult < 1 {
		mult = 1
	}
	cooldown := msg.Cooldown()
	if cooldown == 0 {
		cooldown = s.cooldown
	}
	cooldown *= time.Duration(mult)
	if cooldown > time.Hour*24 {
		cooldown = time.Hour * 24
	}
	return s.cooldown * time.Duration(mult)
}

func (s SlackBot) postMessage(ctx context.Context, msg tower.MessageContext) error {
	payload := slackrest.MessagePayloadPool.Get().(*slackrest.MessagePayload) //nolint
	payload.Reset()
	defer func() {
		slackrest.MessagePayloadPool.Put(payload)
	}()

	blocks, attachments := s.template.BuildTemplate(ctx, msg)
	payload.Blocks = blocks
	payload.Text = msg.Message()
	payload.Mrkdwn = true
	payload.Channel = s.channel
	ctx, cancel := s.setOperationContext(ctx)
	defer cancel()
	resp, err := slackrest.PostMessage(ctx, s.client, s.token, payload)
	go s.deleteGlobalKeyAfterOneSec(ctx)
	if err != nil {
		return msg.Tower().
			Wrap(err).
			Message("failed to post message to slack").
			Context(tower.F{"payload_message": msg.Message()}).
			Log(ctx)
	}
	// TODO: Implement attachments upload.
	_, _ = resp, attachments
	return nil
}

func (s SlackBot) deleteGlobalKeyAfterOneSec(ctx context.Context) {
	<-time.NewTimer(time.Second).C
	s.cache.Delete(ctx, s.globalKey)
}

func (s SlackBot) buildKey(msg tower.MessageContext) string {
	builder := strings.Builder{}
	builder.WriteString(s.Name())
	builder.WriteString(s.cache.Separator())
	service := msg.Service()
	builder.WriteString(service.Environment)
	builder.WriteString(s.cache.Separator())
	builder.WriteString(service.Name)
	builder.WriteString(s.cache.Separator())
	builder.WriteString(service.Type)
	builder.WriteString(s.cache.Separator())

	key := msg.Key()
	if key == "" {
		key = msg.Caller().FormatAsKey()
	}
	builder.WriteString(key)
	return builder.String()
}

func (s SlackBot) getAndSetIter(ctx context.Context, key string) int {
	var iter int
	iterByte, err := s.cache.Get(ctx, key)
	if err == nil {
		iter, _ = strconv.Atoi(string(iterByte))
	}
	iter += 1
	iterByte = []byte(strconv.Itoa(iter))
	nextCooldown := s.cooldown*time.Duration(iter) + s.cooldown
	_ = s.cache.Set(ctx, key, iterByte, nextCooldown)
	return iter
}
