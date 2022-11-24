package towerdiscord

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func (d Discord) defaultEmbedBuilder(ctx context.Context, msg tower.MessageContext, extra *ExtraInformation) ([]*Embed, []*bucket.File) {
	files := make([]*bucket.File, 0, 5)
	embeds := make([]*Embed, 0, 5)
	{
		em, file := d.buildSummary(msg)
		if em != nil {
			embeds = append(embeds, em)
		}
		if file != nil {
			files = append(files, file)
		}
	}
	{
		em, file := d.buildDataEmbed(msg)
		if em != nil {
			embeds = append(embeds, em)
		}
		if file != nil {
			files = append(files, file)
		}
	}
	{
		em, file := d.buildErrorEmbed(msg)
		if em != nil {
			embeds = append(embeds, em)
		}
		if file != nil {
			files = append(files, file)
		}
	}
	{
		em, file := d.buildErrorStackEmbed(msg)
		if em != nil {
			embeds = append(embeds, em)
		}
		if file != nil {
			files = append(files, file)
		}
	}
	{
		em, file := d.buildMetadataEmbed(ctx, msg, extra)
		if em != nil {
			embeds = append(embeds, em)
		}
		if file != nil {
			files = append(files, file)
		}
	}
	return embeds, files
}

func (d Discord) buildSummary(msg tower.MessageContext) (*Embed, *bucket.File) {
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

	_, _ = b.WriteString(msg.Message())
	_, _ = b.WriteString("\n\n")
	_, _ = b.WriteString("```\n")
	err := msg.Err()
	if err != nil {
		switch err := err.(type) {
		case tower.SummaryWriter:
			lw := tower.NewLineWriter(b).LineBreak("\n").Build()
			err.WriteSummary(lw)
		case tower.Summary:
			_, _ = b.WriteString(err.Summary())
		case tower.ErrorWriter:
			lw := tower.NewLineWriter(b).LineBreak("\n").Build()
			err.WriteError(lw)
		default:
			_, _ = b.WriteString(err.Error())
		}
		_, _ = b.WriteString("\n\n")
	}

	for _, c := range msg.Context() {
		switch c := c.(type) {
		case tower.DisplayWriter:
			lw := tower.NewLineWriter(b).LineBreak("\n").Build()
			c.WriteDisplay(lw)
		case tower.Display:
			_, _ = b.WriteString(c.Display())
		}
	}
	_, _ = b.WriteString("\n```")
	if b.Len() > 2000 {
		content := b.String()
		outro := "Content is too long to be displayed fully. See attachment for details"
		if hasClosingTicks(b, len(outro)+5) {
			outro = "\n```\nContent is too long to be displayed fully. See attachment for details"
		}
		b.Truncate(2000 - len(outro))
		_, _ = b.WriteString(outro)
		buf := strings.NewReader(content)
		filename := d.snowflake.Generate().String() + ".md"
		file := bucket.NewFile(buf, filename, "text/markdown")
		file.SetPretext(`# Summary`)
		embed.Description = b.String()
		return embed, file
	}
	embed.Description = b.String()
	return embed, nil
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
			_, _ = b.WriteString("\n\n")
		}
		_, _ = b.WriteString("```")
		switch v := v.(type) {
		case tower.DisplayWriter:
			if hl, ok := v.(HighlightHint); ok {
				_, _ = b.WriteString(hl.DiscordHighlight())
			} else {
				_, _ = b.WriteString("md")
			}
			_, _ = b.WriteRune('\n')
			lw := tower.NewLineWriter(b).LineBreak("\n").Build()
			v.WriteDisplay(lw)
		case tower.Display:
			if hl, ok := v.(HighlightHint); ok {
				_, _ = b.WriteString(hl.DiscordHighlight())
			} else {
				_, _ = b.WriteString("md")
			}
			_, _ = b.WriteRune('\n')
			_, _ = b.WriteString(v.Display())
		default:
			_, _ = b.WriteString("json\n")
			enc := json.NewEncoder(b)
			enc.SetIndent("", "    ")
			enc.SetEscapeHTML(false)
			err := enc.Encode(v)
			if err != nil {
				_, _ = b.WriteString(`{"error":`)
				_, _ = b.WriteString(strconv.Quote(err.Error()))
				_, _ = b.WriteString(`}`)
			}
		}
		_, _ = b.WriteString("\n```")
	}
	content := b.String()
	if b.Len() > 2000 {
		outro := "Content is too long to be displayed fully. See attachment for details"
		if hasClosingTicks(b, len(outro)+5) {
			outro = "\n```\nContent is too long to be displayed fully. See attachment for details"
		}
		b.Truncate(2000 - len(outro))
		_, _ = b.WriteString(outro)
		embed.Description = b.String()
		buf := strings.NewReader(content)
		filename := d.snowflake.Generate().String() + ".md"
		file := bucket.NewFile(buf, filename, "text/markdown")
		file.SetPretext(`# Data`)
		return embed, file
	}
	embed.Description = content
	return embed, nil
}

func (d Discord) buildErrorEmbed(msg tower.MessageContext) (*Embed, *bucket.File) {
	err := msg.Err()
	if err == nil {
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
	_, _ = b.WriteString("```")
	switch err := err.(type) {
	case tower.DisplayWriter:
		if err := err.(HighlightHint); err != nil {
			_, _ = b.WriteString(err.DiscordHighlight())
		} else {
			_, _ = b.WriteString("md")
		}
		_, _ = b.WriteRune('\n')
		lw := tower.NewLineWriter(b).LineBreak("\n").Build()
		err.WriteDisplay(lw)
	case tower.Display:
		if err := err.(HighlightHint); err != nil {
			_, _ = b.WriteString(err.DiscordHighlight())
		} else {
			_, _ = b.WriteString("md")
		}
		_, _ = b.WriteRune('\n')
		_, _ = b.WriteString(err.Display())
	default:
		_, _ = b.WriteString("json\n")
		enc := json.NewEncoder(b)
		enc.SetIndent("", "    ")
		enc.SetEscapeHTML(false)
		errEncode := enc.Encode(err)
		if errEncode != nil {
			_ = enc.Encode(map[string]string{"error": err.Error()})
		}
	}
	content := b.String()
	if b.Len() > 2000 {
		outro := "Content is too long to be displayed fully. See attachment for details"
		if hasClosingTicks(b, len(outro)+5) {
			outro = "\n```\nContent is too long to be displayed fully. See attachment for details"
		}
		b.Truncate(2000 - len(outro))
		_, _ = b.WriteString(outro)
		embed.Description = b.String()
		buf := strings.NewReader(content)
		filename := d.snowflake.Generate().String() + ".md"
		file := bucket.NewFile(buf, filename, "text/markdown")
		file.SetPretext(`# Error`)
		return embed, file
	}
	embed.Description = content
	return embed, nil
}

func (d Discord) buildMetadataEmbed(ctx context.Context, msg tower.MessageContext, extra *ExtraInformation) (*Embed, *bucket.File) {
	embed := &Embed{
		Type:      "rich",
		Title:     "Metadata",
		Color:     0x063970, // Dark Blue
		Timestamp: msg.Time().Format(time.RFC3339),
	}
	for _, v := range d.trace.CaptureTrace(ctx) {
		embed.Fields = append(embed.Fields, &EmbedField{
			Name:   v.Key,
			Value:  v.Value,
			Inline: true,
		})
	}
	embed.Fields = append(embed.Fields, &EmbedField{
		Name:   "Message Iteration",
		Value:  strconv.Itoa(extra.Iteration),
		Inline: true,
	})
	ts := extra.CooldownTimeEnds.Unix()
	embed.Fields = append(embed.Fields, &EmbedField{
		Name:   "Next Earliest Repeat",
		Value:  fmt.Sprintf("<t:%d:F> (<t:%d:R>)", ts, ts),
		Inline: true,
	})
	embed.Fields = append(embed.Fields, &EmbedField{
		Name:   "Cache Key",
		Value:  extra.CacheKey,
		Inline: false,
	})
	if len(embed.Fields) > 10 {
		embed.Fields = embed.Fields[:10]
	}
	b := descBufPool.Get()
	defer descBufPool.Put(b)
	b.Reset()
	b.Grow(2000)
	_, _ = b.WriteString(`**Caller Origin**`)
	_, _ = b.WriteString("\n```\n")
	_, _ = b.WriteString(msg.Caller().String())
	_, _ = b.WriteString("\n```\n")
	_, _ = b.WriteString(`**Caller Function**`)
	_, _ = b.WriteString("\n```\n")
	_, _ = b.WriteString(msg.Caller().ShortOrigin())
	_, _ = b.WriteString("\n```")
	content := b.String()
	if b.Len() > 2000 {
		outro := "Content is too long to be displayed fully. See attachment for details"
		if hasClosingTicks(b, len(outro)+5) {
			outro = "\n```\nContent is too long to be displayed fully. See attachment for details"
		}
		b.Truncate(2000 - len(outro))
		_, _ = b.WriteString(outro)
		embed.Description = b.String()
		buf := strings.NewReader(content)
		filename := d.snowflake.Generate().String() + ".md"
		file := bucket.NewFile(buf, filename, "text/markdown")
		file.SetPretext(`# Metadata`)
		return embed, file
	}
	embed.Description = content
	return embed, nil
}

func (d Discord) buildErrorStackEmbed(msg tower.MessageContext) (*Embed, *bucket.File) {
	err := msg.Err()
	if err == nil {
		return nil, nil
	}
	s := make([]string, 0, 4)
	s = stackAccumulator(s, msg.Err())

	if len(s) == 0 {
		return nil, nil
	}

	content := strings.Join(s, "\n---\n")
	b := descBufPool.Get()
	defer descBufPool.Put(b)
	b.Reset()
	b.Grow(2000)
	_, _ = b.WriteString("```")
	_, _ = b.WriteString(content)
	_, _ = b.WriteString("```")
	content = b.String()
	embed := &Embed{
		Type:      "rich",
		Title:     "Error Stack",
		Color:     0x063970, // Dark Blue
		Timestamp: msg.Time().Format(time.RFC3339),
	}
	if b.Len() > 2000 {
		outro := "Content is too long to be displayed fully. See attachment for details"
		if hasClosingTicks(b, len(outro)+5) {
			outro = "\n```\nContent is too long to be displayed fully. See attachment for details"
		}
		b.Truncate(2000 - len(outro))
		_, _ = b.WriteString(outro)
		embed.Description = b.String()
		buf := strings.NewReader(content)
		filename := d.snowflake.Generate().String() + ".md"
		file := bucket.NewFile(buf, filename, "text/markdown")
		file.SetPretext(`# Error Stack`)
		return embed, file
	}
	return embed, nil
}

func stackAccumulator(s []string, err error) []string {
	if err == nil {
		return s
	}
	ss := &strings.Builder{}
	chWritten := false
	if ch, ok := err.(tower.CallerHint); ok {
		chWritten = true
		ss.WriteString(ch.Caller().String())
	}
	if chWritten {
		if mh, ok := err.(tower.MessageHint); ok {
			ss.WriteString(": ")
			ss.WriteString(mh.Message())
		}
	}
	if ss.Len() > 0 {
		s = append(s, ss.String())
	}
	return stackAccumulator(s, errors.Unwrap(err))
}

func hasClosingTicks(b *bytes.Buffer, countBack int) bool {
	buf := b.Bytes()
	if len(buf) >= countBack {
		buf = buf[len(buf)-countBack:]
	}
	count := bytes.Count(buf, []byte("```"))
	return count%2 == 0
}
