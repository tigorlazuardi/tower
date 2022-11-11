package towerslack

import (
	"context"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower-go/queue"
	"github.com/tigorlazuardi/tower/cache"
)

var _ tower.Messenger = (*SlackBot)(nil)

type SlackBot struct {
	rootContext  context.Context
	token        string
	channel      string
	tracer       tower.TraceCapturer
	name         string
	queue        *queue.Queue[tower.KeyValue[context.Context, tower.MessageContext]]
	slackTimeout time.Duration
	template     TemplateBuilder
	client       Client
	cache        cache.Cacher
	working      int32
	sem          chan struct{}
	globalKey    string
	cooldown     time.Duration
}

// NewSlackBot Creates New Slackbot Instance.
func NewSlackBot(rootContext context.Context, token string, channel string) *SlackBot {
	s := &SlackBot{
		rootContext:  rootContext,
		token:        token,
		channel:      channel,
		tracer:       tower.NoopTracer{},
		queue:        queue.New[tower.KeyValue[context.Context, tower.MessageContext]](),
		slackTimeout: time.Second * 10,
		client:       http.DefaultClient,
		cache:        cache.NewMemoryCache(),
		working:      0,
		sem:          make(chan struct{}, runtime.NumCPU()/3+2),
		globalKey:    "global",
		cooldown:     time.Minute * 15,
	}
	s.template = TemplateFunc(s.defaultTemplate)
	return s
}

func (s *SlackBot) SetRootContext(rootContext context.Context) {
	s.rootContext = rootContext
}

func (s *SlackBot) SetToken(token string) {
	s.token = token
}

func (s *SlackBot) SetChannel(channel string) {
	s.channel = channel
}

func (s *SlackBot) SetTracer(tracer tower.TraceCapturer) {
	s.tracer = tracer
}

func (s *SlackBot) SetName(name string) {
	s.name = name
}

func (s *SlackBot) SetTimeout(slackTimeout time.Duration) {
	s.slackTimeout = slackTimeout
}

func (s *SlackBot) SetMessageTemplate(template TemplateBuilder) {
	s.template = template
}

func (s *SlackBot) SetClient(client Client) {
	s.client = client
}

func (s *SlackBot) SetCache(cache cache.Cacher) {
	s.cache = cache
}

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
	if !s.isWorking() {
		atomic.AddInt32(&s.working, 1)
		go func() {
			for s.queue.Len() > 0 {
				job := s.queue.Dequeue()
				s.sem <- struct{}{}
				go func() {
					ctx := tower.DetachedContext(job.Key)
					s.handleMessage(ctx, job.Value)
					<-s.sem
				}()
			}
			atomic.AddInt32(&s.working, -1)
		}()
	}
}

func (s SlackBot) isWorking() bool {
	return atomic.LoadInt32(&s.working) == 1
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

func (o operationContext) Deadline() (deadline time.Time, ok bool) {
	return o.runningCtx.Deadline()
}

func (o operationContext) Done() <-chan struct{} {
	return o.runningCtx.Done()
}

func (o operationContext) Err() error {
	return o.runningCtx.Err()
}

func (o operationContext) Value(key any) any {
	return o.valueCtx.Value(key)
}

// Detaches given context's deadline and replaces it with own's deadline, but the value is left untouched.
func (s SlackBot) setOperationContext(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(s.rootContext, s.slackTimeout)
	return operationContext{
		runningCtx: ctx,
		valueCtx:   parent,
	}, cancel
}
