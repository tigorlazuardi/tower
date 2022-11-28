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

var _ towerdiscord.Hook = (*testHook)(nil)

type testHook struct {
	t  *testing.T
	wg *sync.WaitGroup
}

func (t testHook) PreMessageHook(ctx context.Context, _ *towerdiscord.WebhookContext) context.Context {
	ctx = context.WithValue(ctx, "test", "test")
	return ctx
}

func (t testHook) PostMessageHook(ctx context.Context, _ *towerdiscord.WebhookContext, err error) {
	defer t.wg.Done()
	if err != nil {
		t.t.Error(err)
	}
	if e, ok := ctx.Value("test").(string); ok {
		if e != "test" {
			t.t.Errorf("context value of test should have value of 'test' not '%s'", e)
		}
	} else {
		t.t.Error("context value of test should exist in PostMessageHook")
	}
}

func (t testHook) PreBucketUploadHook(ctx context.Context, _ *towerdiscord.WebhookContext) context.Context {
	if e, ok := ctx.Value("test").(string); ok {
		if e != "test" {
			t.t.Errorf("context value of test should have value of 'test' not '%s'", e)
		}
	} else {
		t.t.Error("context value of test should exist in PreBucketUploadHook")
	}
	ctx = context.WithValue(ctx, "test-bucket", "test-bucket")
	return ctx
}

func (t testHook) PostBucketUploadHook(ctx context.Context, _ *towerdiscord.WebhookContext, results []bucket.UploadResult) {
	defer t.wg.Done()
	for _, result := range results {
		if result.Error != nil {
			t.t.Error(result.Error)
		}
	}
	if e, ok := ctx.Value("test").(string); ok {
		if e != "test" {
			t.t.Errorf("context value of test should have value of 'test' not '%s'", e)
		}
	} else {
		t.t.Error("context value of test should exist in PreBucketUploadHook")
	}
	if e, ok := ctx.Value("test-bucket").(string); ok {
		if e != "test-bucket" {
			t.t.Errorf("context value of test should have value of 'test-bucket' not '%s'", e)
		}
	} else {
		t.t.Error("context value of test-bucket should exist in PostBucketUploadHook")
	}
}

type foo struct {
	FooMessage string `json:"foo_message,omitempty"`
}

func (f foo) Error() string {
	return f.FooMessage
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
		Environment: "testing-environment",
		Type:        "integration-type",
	})

	wg := &sync.WaitGroup{}
	wg.Add(2)
	bot := towerdiscord.NewDiscordBot(webhook)
	bot.SetName("tower-discord-integration-test")
	bot.SetHook(testHook{t: t, wg: wg})
	tow.RegisterMessenger(bot)
	tow.NewEntry("test %d", 123).Context(tower.F{"foo": "bar", "struct": foo{}}).Notify(ctx)
	origin := tow.Wrap(foo{FooMessage: "something > something < something & Bad Request"}).Code(400).Message("this is bad request error").Context(tower.F{
		"light": map[string]any{"year": 2021, "month": "january"},
		"bar":   "baz",
	}).Freeze()
	wrapped := tow.WrapFreeze(origin, "wrapping error")
	_ = tow.Wrap(wrapped).Message("wrapping error").Context(tower.F{"wrapping": 123, "nil_value": nil}).Notify(ctx)
	err := bot.Wait(ctx)
	if err != nil {
		t.Error(err)
	}

	wg.Wait()
}
