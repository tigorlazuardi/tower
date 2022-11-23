package towerdiscord

import (
	"bytes"
	"context"
	"encoding/json"
	"mime"
	"strings"
	"time"

	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower/bucket"
	"github.com/tigorlazuardi/tower/internal/pool"
)

var descBufPool = pool.New(func() *bytes.Buffer {
	return &bytes.Buffer{}
})

func (d Discord) defaultEmbedBuilder(ctx context.Context, msg tower.MessageContext) ([]*Embed, []*bucket.File) {
	files := make([]*bucket.File, 0, 3)
	embeds := make([]*Embed, 0, 5)
	embeds = append(embeds, d.buildSummary(msg))
	em, file := d.buildDataEmbed(msg)
	embeds = append(embeds, em)
	if file != nil {
		files = append(files, file)
	}
	return embeds, files
}

func (d Discord) buildSummary(msg tower.MessageContext) *Embed {
	embed := &Embed{
		Type:      "rich",
		Title:     "Summary",
		Color:     0x188544, // Green Jewel
		Timestamp: msg.Time().Format(time.RFC3339),
	}
	b := descBufPool.Get()
	defer descBufPool.Put(b)
	b.Reset()
	b.Grow(2000)
	err := msg.Err()

	b.WriteString(msg.Message())
	b.WriteString("\n\n")
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

func (d Discord) buildDataEmbed(msg tower.MessageContext) (*Embed, *bucket.File) {
	if len(msg.Context()) == 0 {
		return nil, nil
	}
	embed := &Embed{
		Type:      "rich",
		Title:     "Summary",
		Color:     0x2596be, // Green Jewel
		Timestamp: msg.Time().Format(time.RFC3339),
	}
	b := descBufPool.Get()
	defer descBufPool.Put(b)
	b.Reset()
	b.Grow(2000)
	filename := d.snowflake.Generate().String()
	filetype := "text/plain"
	for i, v := range msg.Context() {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString("```")
		switch v := v.(type) {
		case tower.DisplayWriter:
			if hl, ok := v.(HighlightHint); ok {
				b.WriteString(hl.DiscordHighlight())
			}
			if hl, ok := v.(MimetypeHint); ok {
				filetype = hl.Mimetype()
			}
			b.WriteString("\n")
			lw := tower.NewLineWriter(b).LineBreak("\n").Build()
			v.WriteDisplay(lw)
		case tower.Display:
			if hl, ok := v.(HighlightHint); ok {
				b.WriteString(hl.DiscordHighlight())
			}
			if hl, ok := v.(MimetypeHint); ok {
				filetype = hl.Mimetype()
			}
			b.WriteString("\n")
			b.WriteString(v.Display())
		default:
			filetype = "application/json"
			b.WriteString("json\n")
			enc := json.NewEncoder(b)
			enc.SetEscapeHTML(false)
			enc.SetIndent("", "  ")
			err := enc.Encode(v)
			if err != nil {
				b.WriteString("json encode error: ")
				b.WriteString(err.Error())
			}
		}
		b.WriteString("\n")
		b.WriteString("```")
	}
	exts, _ := mime.ExtensionsByType(filetype)
	if len(exts) > 0 {
		filename += exts[0]
	}
	content := b.String()
	if b.Len() > 2000 {
		msg := "\n---\nContent truncated. See attachment for more details\n---\n```"
		b.Truncate(2000 - len(msg))
		b.WriteString(msg)
		embed.Description = b.String()
		r := strings.NewReader(content)
		return embed, bucket.NewFile(r, filename, filetype)
	}
	return embed, nil
}
