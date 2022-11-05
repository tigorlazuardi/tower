package towerslack

import (
	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower-go/towerslack/slackrest"
)

func (s Slack) handleMessage(msg tower.MessageContext) {
	// TODO: Implement hooks

	payload := slackrest.MessagePayloadPool.Get().(*slackrest.MessagePayload) //nolint
	payload.Reset()
	defer func() {
		slackrest.MessagePayloadPool.Put(payload)
	}()

	blocks, attachments := s.template.BuildTemplate(msg)
	payload.Blocks = blocks
	payload.Text = msg.Message()
	payload.Mrkdwn = true

	// TODO(urgent): Implement rate limit and global locks.

	ctx, cancel := s.setOperationContext(msg.Ctx())
	defer cancel()
	resp, err := slackrest.PostMessage(ctx, s.client, s.token, payload)
	if err != nil {
		// TODO: Implement tower wrap.
		return
	}
	_ = resp

	// TODO: implement unlocks.
	// Implement file uploads and thread replies.
	_ = attachments
}
