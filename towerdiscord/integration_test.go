package towerdiscord_test

import (
	"context"
	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower/bucket"
	"github.com/tigorlazuardi/tower/internal/loader"
	"github.com/tigorlazuardi/tower/towerdiscord"
	"os"
	"sync"
	"testing"
	"time"
)

type testHook struct {
	t  *testing.T
	wg *sync.WaitGroup
}

func (t testHook) PreMessageHook(ctx context.Context, msg tower.MessageContext, extra *towerdiscord.ExtraInformation) context.Context {
	return ctx
}

func (t testHook) PostMessageHook(ctx context.Context, msg tower.MessageContext, err error) {
	if err != nil {
		t.t.Error(err)
	}
	t.wg.Done()
}

func (t testHook) PreBucketUploadHook(ctx context.Context, bucket bucket.Bucket, files []*bucket.File) context.Context {
	return ctx
}

func (t testHook) PostBucketUploadHook(ctx context.Context, msg tower.MessageContext, results []bucket.UploadResult) {
	for _, result := range results {
		if result.Error != nil {
			t.t.Error(result.Error)
		}
	}
}

func TestIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	loader.LoadEnv()
	webhook := os.Getenv("DISCORD_WEBHOOK")
	if os.Getenv("DISCORD_WEBHOOK") == "" {
		t.Skip("DISCORD_WEBHOOK env is not set. Skipping integration test")
	}
	tow := tower.NewTower(tower.Service{
		Name:        "discord-integration-test",
		Environment: "test",
		Type:        "integration",
	})

	wg := &sync.WaitGroup{}
	wg.Add(1)
	bot := towerdiscord.NewDiscordBot(webhook)
	bot.SetHook(testHook{t: t, wg: wg})
	tow.RegisterMessenger(bot)
	tow.NewEntry("test").Notify(ctx)
	err := bot.Wait(ctx)
	if err != nil {
		t.Error(err)
	}
	wg.Wait()
}
