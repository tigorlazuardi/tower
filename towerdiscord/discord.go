package towerdiscord

import (
	"context"
	"net/http"
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

type QueueItem tower.KeyValue[context.Context, tower.MessageContext]

func NewQueueItem(ctx context.Context, messageContext tower.MessageContext) QueueItem {
	w := tower.NewKeyValue(ctx, messageContext)
	return QueueItem(w)
}

type Discord struct {
	name             string
	webhook          string
	cache            cache.Cacher
	queue            *queue.Queue[QueueItem]
	sem              chan struct{}
	working          int32
	trace            tower.TraceCapturer
	builder          EmbedBuilder
	bucket           bucket.Bucket
	globalKey        string
	cooldown         time.Duration
	snowflake        *snowflake.Node
	client           Client
	hook             Hook
	dataEncoder      DataEncoder
	codeBlockBuilder CodeBlockBuilder
}

// NewDiscordBot creates a new discord bot.
func NewDiscordBot(webhook string, opts ...DiscordOption) *Discord {
	host, _ := os.Hostname()
	d := &Discord{
		name:             "discord",
		webhook:          webhook,
		cache:            cache.NewLocalCache(),
		queue:            queue.New[QueueItem](500),
		sem:              make(chan struct{}, (runtime.NumCPU()/3)+2),
		trace:            tower.NoopTracer{},
		globalKey:        "global",
		cooldown:         time.Minute * 15,
		snowflake:        generateSnowflakeNodeFromString(host + webhook),
		client:           http.DefaultClient,
		hook:             NoopHook{},
		dataEncoder:      JSONDataEncoder{},
		codeBlockBuilder: JSONCodeBlockBuilder{},
	}
	d.builder = EmbedBuilderFunc(d.defaultEmbedBuilder)
	for _, opt := range opts {
		opt.apply(d)
	}
	return d
}

// Name implements tower.Messenger interface.
func (d Discord) Name() string {
	if d.name == "" {
		return "discord"
	}
	return d.name
}

// SendMessage implements tower.Messenger interface.
func (d Discord) SendMessage(ctx context.Context, msg tower.MessageContext) {
	d.queue.Enqueue(NewQueueItem(ctx, msg))
	d.work()
}

func (d *Discord) work() {
	if atomic.CompareAndSwapInt32(&d.working, 0, 1) {
		go func() {
			for d.queue.HasNext() {
				d.sem <- struct{}{}
				kv := d.queue.Dequeue()
				go func() {
					ctx := tower.DetachedContext(kv.Key)
					d.send(ctx, kv.Value)
					<-d.sem
				}()
			}
			atomic.StoreInt32(&d.working, 0)
		}()
	}
}

// Wait implements tower.Messenger interface.
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
