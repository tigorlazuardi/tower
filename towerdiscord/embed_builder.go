package towerdiscord

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
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
	{
		em, file := d.buildDataEmbed(msg)
		if em != nil {
			embeds = append(embeds, em)
		}
		if file != nil {
			files = append(files, file)
		}
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
		Color:     0x063970, // Dark Blue
		Timestamp: msg.Time().Format(time.RFC3339),
	}
	b := descBufPool.Get()
	defer descBufPool.Put(b)
	b.Reset()
	b.Grow(2000)
	for i, v := range msg.Context() {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString("```")
		switch v := v.(type) {
		case tower.DisplayWriter:
			if hl, ok := v.(HighlightHint); ok {
				b.WriteString(hl.DiscordHighlight())
			} else {
				b.WriteString("md")
			}
			b.WriteRune('\n')
			lw := tower.NewLineWriter(b).LineBreak("\n").Build()
			v.WriteDisplay(lw)
		case tower.Display:
			if hl, ok := v.(HighlightHint); ok {
				b.WriteString(hl.DiscordHighlight())
			} else {
				b.WriteString("md")
			}
			b.WriteRune('\n')
			b.WriteString(v.Display())
		default:
			b.WriteString("json\n")
			enc := json.NewEncoder(b)
			enc.SetIndent("", "    ")
			enc.SetEscapeHTML(false)
			err := enc.Encode(v)
			if err != nil {
				b.WriteString(`{"error":`)
				b.WriteString(strconv.Quote(err.Error()))
				b.WriteString(`}`)
			}
		}
		b.WriteString("\n```")
	}
	content := b.String()
	if b.Len() > 2000 {
		outro := "Content is too long to be displayed fully. See attachment for details"
		if hasClosingTicks(b, len(outro)+5) {
			outro = "\n```\nContent is too long to be displayed fully. See attachment for details"
		}
		b.Truncate(2000 - len(outro))
		b.WriteString(outro)
		embed.Description = b.String()
		buf := strings.NewReader(content)
		filename := d.snowflake.Generate().String() + ".md"
		return embed, bucket.NewFile(buf, filename, "text/markdown")
	}
	embed.Description = content
	return embed, nil
}

func hasClosingTicks(b *bytes.Buffer, countback int) bool {
	buf := b.Bytes()
	if len(buf) >= countback {
		buf = buf[len(buf)-countback:]
	}
	count := bytes.Count(buf, []byte("```"))
	return count%2 == 0
}
