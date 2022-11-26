package towerslack

import (
	"context"
	"fmt"
	"github.com/tigorlazuardi/tower/bucket"
	"strings"
	"time"

	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower/towerslack/block"
)

type TemplateBuilder interface {
	// BuildTemplate builds template for SlackBot Message.
	// Avoid long side effects as much as possible when building the template (like request to DB). It may block up the queue,
	// because towerslack limits the goroutine it will spin up to consume message queue.
	//
	// If any of the block overpasses the content limit, you may instead return bucket.File attachments and towerslack will upload those files to a Bucket and post them as a reply to the main message.
	//
	// After attachment has been uploaded, The Close method on the attachment's body will be called by towerslack.
	//
	// If you have no attachments to upload, a simple nil return on the attachments is safe.
	//
	// Note: The blocks are required and must not be nil (returning empty blocks are safe however), regardless of attachments.
	BuildTemplate(ctx context.Context, msg tower.MessageContext) (blocks block.Blocks, attachments []bucket.File)
}

var _ TemplateBuilder = (TemplateFunc)(nil)

type TemplateFunc func(ctx context.Context, msg tower.MessageContext) (block.Blocks, []bucket.File)

// BuildTemplate implements Templater interface.
func (f TemplateFunc) BuildTemplate(ctx context.Context, msg tower.MessageContext) (block.Blocks, []bucket.File) {
	return f(ctx, msg)
}

func (s SlackBot) defaultTemplate(ctx context.Context, msg tower.MessageContext) (block.Blocks, []bucket.File) {
	blocks := make(block.Blocks, 0, 6)
	attachments := make([]bucket.File, 0, 5)
	blocks = append(blocks, buildHeadline(msg))
	blocks = append(blocks, block.NewDividerBlock())

	blocks = append(blocks, block.NewHeaderBlock("Summary"))
	blocks = append(blocks, buildSummary(msg))
	blocks = append(blocks, s.buildMetadata(ctx, msg))

	return blocks, attachments
}

func buildHeadline(msg tower.MessageContext) *block.SectionBlock {
	var headline string
	service := msg.Service()
	if msg.Err() != nil {
		headline = fmt.Sprintf("<!here> an error has occured on service **%s** on type **%s** on environment **%s**", service.Name, service.Type, service.Environment)
	} else {
		headline = fmt.Sprintf("<!here> a message from service **%s** on type **%s** on environment **%s*", service.Name, service.Type, service.Environment)
	}
	return block.NewSectionBlockText(block.TextMrkdwn, headline)
}

func buildSummary(msg tower.MessageContext) *block.SectionBlock {
	summary := &strings.Builder{}
	summary.Grow(3000)
	summary.WriteString("```")
	summary.WriteString(msg.Message())
	summary.WriteString("\n\n")
	if msg.Err() != nil {
		if ew, ok := msg.Err().(tower.ErrorWriter); ok {
			lw := tower.NewLineWriter(summary).LineBreak("\n    ").Build()
			ew.WriteError(lw)
		} else {
			summary.WriteString(msg.Err().Error())
		}
		summary.WriteString("\n\n")
	}

	for i, v := range msg.Context() {
		if i > 0 {
			summary.WriteString("\n\n")
		}
		switch v := v.(type) {
		case tower.SummaryWriter:
			lw := tower.NewLineWriter(summary).LineBreak("\n").Build()
			v.WriteSummary(lw)
		case tower.Summary:
			summary.WriteString(v.Summary())
		}
	}

	summary.WriteString("```")
	content := summary.String()
	if len(content) > 3000 {
		content = content[:2997] + "```"
	}
	return block.NewSectionBlockText(block.TextMrkdwn, content)
}

func (s SlackBot) buildMetadata(ctx context.Context, msg tower.MessageContext) *block.SectionBlock {
	texts := make([]string, 0, 10)

	if trace := s.tracer.CaptureTrace(ctx); trace != nil {
		for _, v := range trace {
			text := fmt.Sprintf("**%s**\n%s", v.Key, v.Value)
			if len(text) > 100 {
				text = text[:100]
			}
			texts = append(texts, text)
		}
	}

	texts = append(texts,
		fmt.Sprintf("**Timestamp**\n<!date^%d^{date_short_pretty} {time_secs}|%s>", msg.Time().Unix(), msg.Time().Format(time.RubyDate)),
	)

	service := msg.Service()
	texts = append(texts, fmt.Sprintf("**Service**\n%s", service.Name))
	texts = append(texts, fmt.Sprintf("**Type**\n%s", service.Type))
	texts = append(texts, fmt.Sprintf("**Environment**\n%s", service.Environment))

	if len(texts) > 10 {
		texts = texts[:10]
	}

	return block.NewSectionBlockFields(block.TextMrkdwn, texts...)
}
