package towerslack

import (
	"context"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/tigorlazuardi/tower/bucket"

	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower/cache"
	"github.com/tigorlazuardi/tower/queue"
)

type (
	queueItem     = tower.KeyValue[context.Context, tower.MessageContext]
	fileQueueItem = tower.KeyValue[UploadTarget, *bucket.File]
)

var _ tower.Messenger = (*SlackBot)(nil)

type SlackBot struct {
	rootContext   context.Context
	token         string
	channel       string
	tracer        tower.TraceCapturer
	name          string
	queue         *queue.Queue[queueItem]
	fileQueue     *queue.Queue[fileQueueItem]
	bucket        bucket.Bucket
	slackTimeout  time.Duration
	template      TemplateBuilder
	client        Client
	cache         cache.Cacher
	working       int32
	uploading     int32
	sem           chan struct{}
	globalKey     string
	globalFileKey string
	cooldown      time.Duration
}

// SetBucket sets the bucket to upload files for the slackbot. If not set, upload files to slack instead.
func (s *SlackBot) SetBucket(bucket bucket.Bucket) {
	s.bucket = bucket
}

// NewSlackBot Creates New Slackbot Instance.
//
// If you create multiple bot instances, make sure to set different name for each instance. Otherwise, Tower will treat
// them as same and only registers one instance.
func NewSlackBot(token string, channel string) *SlackBot {
	cache := cache.NewMemoryCache()
	s := &SlackBot{
		rootContext:   context.Background(),
		token:         token,
		channel:       channel,
		tracer:        tower.NoopTracer{},
		queue:         queue.New[tower.KeyValue[context.Context, tower.MessageContext]](500),
		fileQueue:     queue.New[tower.KeyValue[UploadTarget, *bucket.File]](500),
		slackTimeout:  time.Second * 30,
		client:        http.DefaultClient,
		cache:         cache,
		working:       0,
		uploading:     0,
		sem:           make(chan struct{}, runtime.NumCPU()/3+2),
		globalKey:     "global",
		globalFileKey: cache.Separator(),
		cooldown:      time.Minute * 15,
	}
	s.template = TemplateFunc(s.defaultTemplate)
	return s
}

// SetRootContext sets the base context.
// If given context is canceled, all the ongoing messages will have their context canceled as well.
//
// When this method is called, if there are already messages are already on going, they will still use the old context.
func (s *SlackBot) SetRootContext(rootContext context.Context) {
	s.rootContext = rootContext
}

// SetToken changes the token for the bot.
func (s *SlackBot) SetToken(token string) {
	s.token = token
}

// SetChannel changes the channel for the bot.
func (s *SlackBot) SetChannel(channel string) {
	s.channel = channel
}

// SetTracer setups how Slackbot will capture Traces.
func (s *SlackBot) SetTracer(tracer tower.TraceCapturer) {
	s.tracer = tracer
}

// SetName sets the name for this slackbot instance.
func (s *SlackBot) SetName(name string) {
	s.name = name
}

// SetTimeout sets the timeout for requests that are sent to Slack.
func (s *SlackBot) SetTimeout(slackTimeout time.Duration) {
	s.slackTimeout = slackTimeout
}

// SetMessageTemplate sets the block template for the message.
func (s *SlackBot) SetMessageTemplate(template TemplateBuilder) {
	s.template = template
}

// SetClient sets the http client for the slackbot.
func (s *SlackBot) SetClient(client Client) {
	s.client = client
}

// SetCache sets the caching mechanism for the slackbot.
func (s *SlackBot) SetCache(cache cache.Cacher) {
	s.cache = cache
}

// SetBaseCooldown sets the cooldown for all messages if they are not overridden by per message options.
func (s *SlackBot) SetBaseCooldown(cooldown time.Duration) {
	s.cooldown = cooldown
}

// Name Returns the name of the Messenger.
func (s SlackBot) Name() string {
	if s.name == "" {
		return "slack"
	}
	return s.name
}

// SendMessage Sends notification.
func (s SlackBot) SendMessage(ctx context.Context, msg tower.MessageContext) {
	job := tower.KeyValue[context.Context, tower.MessageContext]{Key: ctx, Value: msg}
	s.queue.Enqueue(job)
	s.work()
}

func (s *SlackBot) work() {
	if atomic.CompareAndSwapInt32(&s.working, 0, 1) {
		go func() {
			for s.queue.HasNext() {
				job := s.queue.Dequeue()
				s.sem <- struct{}{}
				go func() {
					ctx := tower.DetachedContext(job.Key)
					s.handleMessage(ctx, job.Value)
					<-s.sem
				}()
			}
			atomic.StoreInt32(&s.working, 0)
		}()
	}
}

// Wait until all message in the queue or until given channel is received.
//
// Implementer must exit the function as soon as possible when this ctx is canceled.
func (s SlackBot) Wait(ctx context.Context) error {
	err := make(chan error)
	go func() {
		for {
			select {
			case <-ctx.Done():
				err <- ctx.Err()
				break
			default:
				if s.queue.Len() == 0 {
					err <- nil
					break
				}
				time.Sleep(time.Millisecond * 50)
			}
		}
	}()

	return <-err
}

type operationContext struct {
	runningCtx context.Context
	valueCtx   context.Context
}

// Deadline implements context.Context.
func (o operationContext) Deadline() (deadline time.Time, ok bool) {
	return o.runningCtx.Deadline()
}

// Done implements context.Context.
func (o operationContext) Done() <-chan struct{} {
	return o.runningCtx.Done()
}

// Err implements context.Context.
func (o operationContext) Err() error {
	return o.runningCtx.Err()
}

// Value implements context.Context.
func (o operationContext) Value(key any) any {
	// Checks the Valuer first, because, in perspective, it contains lower lifetime scope of information.
	if v := o.valueCtx.Value(key); v != nil {
		return v
	}
	// Fallback to the root context, just in case the user sets it on the root context.
	return o.runningCtx.Value(key)
}

// Detaches given context's deadline and replaces it with owns deadline, but the channel is left untouched.
func (s SlackBot) setOperationContext(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(s.rootContext, s.slackTimeout)
	return operationContext{
		runningCtx: ctx,
		valueCtx:   parent,
	}, cancel
}
