package towerslack

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower/bucket"
	"github.com/tigorlazuardi/tower/towerslack/slackrest"
)

type Target int

func (t Target) String() string {
	switch t {
	case TargetChannel:
		return "channel"
	case TargetThread:
		return "thread"
	default:
		return "unknown"
	}
}

const (
	TargetChannel Target = iota
	TargetThread
)

// UploadTarget is the target on which file attachments are posted to.
type UploadTarget interface {
	Kind() Target
	Value() string
	Context() context.Context
	Tower() *tower.Tower
}

type target struct {
	kind  Target
	value string
	ctx   context.Context
	tower *tower.Tower
}

func (c target) Context() context.Context {
	return c.ctx
}

func (c target) Kind() Target {
	return c.kind
}

func (c target) Value() string {
	return c.value
}

func (c target) Tower() *tower.Tower {
	return c.tower
}

// PostToChannel posts the file attachment to the channel.
func PostToChannel(ctx context.Context, tower *tower.Tower, channel string) UploadTarget {
	return &target{
		kind:  TargetChannel,
		value: channel,
		ctx:   ctx,
		tower: tower,
	}
}

// PostToThread posts the file attachment to the thread.
func PostToThread(ctx context.Context, tower *tower.Tower, thread string) UploadTarget {
	return &target{
		kind:  TargetThread,
		value: thread,
		ctx:   ctx,
		tower: tower,
	}
}

func (s SlackBot) uploadAttachments(ctx context.Context, msg tower.MessageContext, resp *slackrest.MessageResponse, attachments []bucket.File) {
	if s.bucket != nil {
		s.uploadToBucket(ctx, msg, resp, attachments)
		return
	}
	s.uploadToSlack(ctx, msg, resp, attachments)
}

func (s SlackBot) uploadToBucket(ctx context.Context, msg tower.MessageContext, resp *slackrest.MessageResponse, attachments []bucket.File) {
	results := s.bucket.Upload(ctx, attachments)
	for _, result := range results {
		if result.Error != nil {
			_ = msg.Tower().WrapFreeze(result.Error, "failed to upload file").Log(ctx)
			continue
		}
	}
}

func (s SlackBot) uploadToSlack(ctx context.Context, msg tower.MessageContext, resp *slackrest.MessageResponse, attachments []bucket.File) {
	for _, attachment := range attachments {
		key := PostToThread(ctx, msg.Tower(), resp.Ts)
		value := attachment
		item := tower.NewKeyValue(key, value)
		s.fileQueue.Enqueue(item)
		s.upload()
	}
}

func (s SlackBot) upload() {
	if !s.isUploading() {
		atomic.AddInt32(&s.uploading, 1)
		go func() {
			for s.fileQueue.Len() > 0 {
				job := s.fileQueue.Dequeue()
				s.sem <- struct{}{}
				go func() {
					ctx := tower.DetachedContext(job.Key.Context())
					err := s.uploadFile(ctx, job.Key.(UploadTarget), job.Value)
					if err != nil {
						_ = job.Key.Tower().Wrap(err).Message("failed to upload file").Log(ctx)
					}
					<-s.sem
				}()
			}
			atomic.AddInt32(&s.uploading, -1)
		}()
	}
}

func (s SlackBot) isUploading() bool {
	return atomic.LoadInt32(&s.uploading) == 1
}

func (s SlackBot) uploadFile(ctx context.Context, target UploadTarget, file bucket.File) error {
	defer func(file bucket.File) {
		_ = file.Close()
	}(file)
	ticker := time.NewTicker(time.Millisecond * 300)
	for s.cache.Exist(ctx, s.globalFileKey) {
		<-ticker.C
	}
	ticker.Stop()
	err := s.cache.Set(ctx, s.globalFileKey, []byte(file.Filename()), time.Minute*5)
	if err != nil {
		_ = target.Tower().WrapFreeze(err, "failed to set global file key").Log(ctx)
	}
	defer s.cache.Delete(ctx, s.globalFileKey)
	payload := slackrest.FilesUploadPayload{
		File:           file,
		Filename:       file.Filename(),
		Filetype:       file.ContentType(),
		InitialComment: file.Pretext(),
		Title:          file.Filename(),
	}
	switch target.Kind() {
	case TargetChannel:
		payload.Channels = []string{target.Value()}
	case TargetThread:
		payload.ThreadTS = target.Value()
	default:
		return target.Tower().
			Bail("unknown post target").
			Context(tower.F{
				"target": target.Kind(),
				"value":  target.Value(),
			}).
			Freeze()
	}
	_, err = slackrest.FileUpload(ctx, s.client, payload)
	if err != nil {
		time.Sleep(time.Second * 1)
		return target.Tower().WrapFreeze(err, "failed to upload file")
	}
	time.Sleep(time.Second * 3)
	return nil
}
