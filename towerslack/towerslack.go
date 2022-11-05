package towerslack

import (
	"context"
	"time"

	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower-go/queue"
)

type Slack struct {
	rootContext  context.Context
	token        string
	channel      string
	tracer       tower.TraceCapturer
	name         string
	queue        *queue.Queue[tower.MessageContext]
	slackTimeout time.Duration
	template     Templater
	client       Client
}

// Returns the name of the Messenger.
func (s Slack) Name() string {
	if s.name == "" {
		return "slack"
	}
	return s.name
}

// Sends notification.
func (s Slack) SendMessage(msg tower.MessageContext) {
	s.queue.Enqueue(msg)
}

// Waits until all message in the queue or until given channel is received.
//
// Implementer must exit the function as soon as possible when this ctx is canceled.
func (s Slack) Wait(ctx context.Context) error {
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
func (s Slack) setOperationContext(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(s.rootContext, s.slackTimeout)
	return operationContext{
		runningCtx: ctx,
		valueCtx:   parent,
	}, cancel
}
