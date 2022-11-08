package towerslack

import (
	"io"

	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower-go/towerslack/block"
)

type FileAttachment struct {
	Body          io.ReadCloser
	Filename      string
	Mimetype      string
	ContentLength int
	ForceBucket   bool
}

type TemplateBuilder interface {
	// Avoid long side effects as much as possible when building the template (like request to DB). It may blocks up the queue,
	// because towerslack limits the goroutine it will spin up to consume message queue.
	//
	// If any of the block overpasses the content limit, you may instead return file attachments and towerslack will upload those files to a Bucket and post them as a reply to the main message.
	// If the file's Mimetype is human readable and ContentLength is under 1MB, it will be uploaded as Snippet instead in a reply thread, unless ForceBucket option is true.
	//
	// After attachment has been uploaded, The Close method on the attachment's body will be called by towerslack.
	//
	// If you have no attachments to upload, a simple nil return on the attachments is safe.
	//
	// Note: The blocks are required and must not be nil (returning empty blocks are safe however), regardless of attachments.
	BuildTemplate(msg tower.MessageContext) (blocks block.Blocks, attachments []FileAttachment)
}

var _ TemplateBuilder = (TemplateFunc)(nil)

type TemplateFunc func(msg tower.MessageContext) (block.Blocks, []FileAttachment)

// implements Templater interface.
func (f TemplateFunc) BuildTemplate(msg tower.MessageContext) (block.Blocks, []FileAttachment) {
	return f(msg)
}

func (s Slack) defaultTemplate(msg tower.MessageContext) (block.Blocks, []FileAttachment) {
	panic("Not implemented") // TODO: Implement
}
