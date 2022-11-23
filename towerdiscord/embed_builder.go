package towerdiscord

import (
	"bytes"
	"context"
	"time"

	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower/bucket"
	"github.com/tigorlazuardi/tower/internal/pool"
)

var descBufPool = pool.New(func() *bytes.Buffer {
	return &bytes.Buffer{}
})

func defaultEmbedBuilder(ctx context.Context, msg tower.MessageContext) ([]*Embed, []*bucket.File) {
	embeds := make([]*Embed, 0, 5)
	embeds = append(embeds, buildSummary(msg))
	return embeds, nil
}

func buildSummary(msg tower.MessageContext) *Embed {
	embed := &Embed{
		Type:      "rich",
		Title:     msg.Message(),
		Color:     0x2596be, // Green Jewel
		Timestamp: msg.Time().Format(time.RFC3339),
	}
	b := descBufPool.Get()
	defer descBufPool.Put(b)
	b.Reset()
	b.Grow(2000)
	err := msg.Err()

	b.WriteString("**Summary**\n")
	b.WriteString("```\n")
	if err != nil {
		switch err := err.(type) {
		case tower.ErrorWriter:
			lw := tower.NewLineWriter(b).LineBreak("\n").Build()
			err.WriteError(lw)
		default:
			b.WriteString(err.Error())
		}
		b.WriteString("\n\n")
	}

	for _, c := range msg.Context() {
		switch c := c.(type) {
		case tower.DisplayWriter:
			lw := tower.NewLineWriter(b).LineBreak("\n").Build()
			c.WriteDisplay(lw)
		case tower.Display:
			b.WriteString(c.Display())
		}
	}
	b.WriteString("\n```")
	if b.Len() > 2000 {
		b.Truncate(1992)
		b.WriteString("\n...\n```")
	}
	embed.Description = b.String()
	return embed
}
