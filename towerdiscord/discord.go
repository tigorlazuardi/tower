package towerdiscord

import (
	"context"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/tigorlazuardi/tower/bucket"

	"github.com/bwmarrin/snowflake"
	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower/cache"
	"github.com/tigorlazuardi/tower/queue"
)

func init() {
	snowflake.Epoch = 1420070400000 // discord epoch
}

type Discord struct {
	name      string
	webhook   string
	cache     cache.Cacher
	queue     *queue.Queue[tower.KeyValue[context.Context, tower.MessageContext]]
	sem       chan struct{}
	working   int32
	trace     tower.TraceCapturer
	builder   EmbedBuilder
	bucket    bucket.Bucket
	globalKey string
	cooldown  time.Duration
	snowflake *snowflake.Node
}

// SetName sets the name of the bot. This is used for identification of the bot for tower.
func (d *Discord) SetName(name string) {
	d.name = name
}

// Webhook returns the registered webhook for the bot.
func (d *Discord) Webhook() string {
	return d.webhook
}

// SetWebhook sets the webhook for the bot.
func (d *Discord) SetWebhook(webhook string) {
	d.webhook = webhook
}

// SetCache sets the cacher engine.
func (d *Discord) SetCache(cache cache.Cacher) {
	d.cache = cache
}

func (d *Discord) SetSnowflakeGenerator(node *snowflake.Node) {
	d.snowflake = node
}

func NewDiscordBot(webhook string) *Discord {
	host, _ := os.Hostname()
	return &Discord{
		name:      "discord",
		webhook:   webhook,
		cache:     cache.NewMemoryCache(),
		queue:     queue.New[tower.KeyValue[context.Context, tower.MessageContext]](),
		sem:       make(chan struct{}, (runtime.NumCPU()/3)+2),
		trace:     tower.NoopTracer{},
		builder:   EmbedBuilderFunc(defaultEmbedBuilder),
		bucket:    nil,
		globalKey: "global",
		cooldown:  time.Minute * 15,
		snowflake: generateSnowflakeNodeFromString(host + webhook),
	}
}

func (d Discord) Name() string {
	if d.name == "" {
		return "discord"
	}
	return d.name
}

func (d Discord) SendMessage(ctx context.Context, msg tower.MessageContext) {
	d.queue.Enqueue(tower.NewKeyValue(ctx, msg))
	d.work()
}

func (d *Discord) work() {
	if atomic.CompareAndSwapInt32(&d.working, 0, 1) {
		go func() {
			for d.queue.Len() > 0 {
				d.sem <- struct{}{}
				go func() {
					kv := d.queue.Dequeue()
					ctx := tower.DetachedContext(kv.Key)
					d.send(ctx, kv.Value)
					<-d.sem
				}()
			}
			atomic.StoreInt32(&d.working, 0)
		}()
	}
}

func (d Discord) Wait(ctx context.Context) error {
	err := make(chan error)
	go func() {
		for {
			select {
			case <-ctx.Done():
				err <- ctx.Err()
				break
			default:
				if d.queue.Len() == 0 {
					err <- nil
					break
				}
				time.Sleep(time.Millisecond * 50)
			}
		}
	}()

	return <-err
}

func generateSnowflakeNodeFromString(s string) *snowflake.Node {
	id := 0
	for _, c := range s {
		id += int(c)
	}
	for id > 1023 {
		id >>= 1
	}
	node, _ := snowflake.NewNode(int64(id))
	return node
}
