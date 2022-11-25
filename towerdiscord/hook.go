package towerdiscord

import (
	"context"
	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower/bucket"
)

type Hook interface {
	PreMessageHook(ctx context.Context, msg tower.MessageContext, extra *ExtraInformation) context.Context
	PostMessageHook(ctx context.Context, msg tower.MessageContext, err error)
	PreBucketUploadHook(ctx context.Context, bucket bucket.Bucket, files []*bucket.File) context.Context
	PostBucketUploadHook(ctx context.Context, msg tower.MessageContext, results []bucket.UploadResult)
}

type NoopHook struct{}

func (n NoopHook) PreMessageHook(ctx context.Context, msg tower.MessageContext, extra *ExtraInformation) context.Context {
	return ctx
}
func (n NoopHook) PostMessageHook(ctx context.Context, msg tower.MessageContext, err error) {}
func (n NoopHook) PreBucketUploadHook(ctx context.Context, bucket bucket.Bucket, files []*bucket.File) context.Context {
	return ctx
}
func (n NoopHook) PostBucketUploadHook(ctx context.Context, msg tower.MessageContext, results []bucket.UploadResult) {
}
