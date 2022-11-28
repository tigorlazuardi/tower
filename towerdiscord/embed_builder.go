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

const descriptionLimit = 4096

func (d Discord) defaultEmbedBuilder(ctx context.Context, msg tower.MessageContext, extra *ExtraInformation) ([]*Embed, []bucket.File) {
	files := make([]bucket.File, 0, 5)
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
		em, file := d.buildContextEmbed(msg)
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

func (d Discord) buildSummary(msg tower.MessageContext) (*Embed, bucket.File) {
	embed := &Embed{
		Type:  "rich",
		Title: "Summary",
		Color: 0x188544, // Green Jewel
	}
	b := descBufPool.Get()
	defer descBufPool.Put(b)
	b.Reset()
	b.Grow(descriptionLimit)

	_, _ = b.WriteString("**")
	_, _ = b.WriteString(msg.Message())
	_, _ = b.WriteString("**")
	err := msg.Err()
	if err != nil {
		_, _ = b.WriteString("\n\n**Error**:\n")
		_, _ = b.WriteString("```\n")
		switch err := err.(type) {
		case tower.SummaryWriter:
			lw := tower.NewLineWriter(b).LineBreak("\n").Build()
			err.WriteSummary(lw)
		case tower.Summary:
			_, _ = b.WriteString(err.Summary())
		case tower.ErrorWriter:
			lw := tower.NewLineWriter(b).LineBreak("\n.. ").Build()
			err.WriteError(lw)
		default:
			_, _ = b.WriteString(err.Error())
		}
		_, _ = b.WriteString("\n```")
	}

	data := msg.Context()
	if len(data) > 0 {
		_, _ = b.WriteString("\n\n**Context**:\n")
		_, _ = b.WriteString("```\n")
		for _, c := range data {
			switch c := c.(type) {
			case tower.SummaryWriter:
				lw := tower.NewLineWriter(b).LineBreak("\n").Build()
				c.WriteSummary(lw)
			case tower.Summary:
				_, _ = b.WriteString(c.Summary())
			}
		}
		_, _ = b.WriteString("\n```")
	}
	return d.shouldCreateFile(embed, b, "# Summary")
}

func (d Discord) buildContextEmbed(msg tower.MessageContext) (*Embed, bucket.File) {
	if len(msg.Context()) == 0 {
		return nil, nil
	}
	embed := &Embed{
		Type:  "rich",
		Title: "Context",
		Color: 0x063970, // Dark Blue
	}
	b := descBufPool.Get()
	defer descBufPool.Put(b)
	b.Reset()
	b.Grow(descriptionLimit)
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
	return d.shouldCreateFile(embed, b, "# Data")
}

func (d Discord) buildErrorEmbed(msg tower.MessageContext) (*Embed, bucket.File) {
	err := msg.Err()
	if err == nil {
		return nil, nil
	}
	embed := &Embed{
		Type:  "rich",
		Title: "Error",
		Color: 0x71010b, // Venetian Red
	}
	b := descBufPool.Get()
	defer descBufPool.Put(b)
	b.Reset()
	b.Grow(descriptionLimit)
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
	_, _ = b.WriteString("```")
	return d.shouldCreateFile(embed, b, "# Error")
}

func (d Discord) buildMetadataEmbed(ctx context.Context, msg tower.MessageContext, extra *ExtraInformation) (*Embed, bucket.File) {
	embed := &Embed{
		Type:      "rich",
		Title:     "Metadata",
		Color:     0x645a5b, // Scorpion Grey
		Timestamp: msg.Time().Format(time.RFC3339),
	}
	for _, v := range d.trace.CaptureTrace(ctx) {
		embed.Fields = append(embed.Fields, &EmbedField{
			Name:   v.Key,
			Value:  v.Value,
			Inline: true,
		})
	}
	service := msg.Service()
	if service.Name != "" {
		embed.Fields = append(embed.Fields, &EmbedField{
			Name:   "Service",
			Value:  service.Name,
			Inline: true,
		})
	}
	if service.Type != "" {
		embed.Fields = append(embed.Fields, &EmbedField{
			Name:   "Type",
			Value:  service.Type,
			Inline: true,
		})
	}
	if service.Environment != "" {
		embed.Fields = append(embed.Fields, &EmbedField{
			Name:   "Environment",
			Value:  service.Environment,
			Inline: true,
		})
	}
	embed.Fields = append(embed.Fields, &EmbedField{
		Name:   "Thread ID",
		Value:  extra.ThreadID.String(),
		Inline: true,
	})
	var iteration string
	if msg.SkipVerification() {
		iteration = "(skipped verification)"
	} else {
		iteration = strconv.Itoa(extra.Iteration)
	}
	embed.Fields = append(embed.Fields, &EmbedField{
		Name:   "Message Iteration",
		Value:  iteration,
		Inline: true,
	})
	ts := extra.CooldownTimeEnds.Unix()
	embed.Fields = append(embed.Fields, &EmbedField{
		Name:   "Next Possible Earliest Repeat",
		Value:  fmt.Sprintf("<t:%d:F> | <t:%d:R>", ts, ts),
		Inline: false,
	})
	if len(embed.Fields) > 25 {
		embed.Fields = embed.Fields[:25]
	}
	b := descBufPool.Get()
	defer descBufPool.Put(b)
	b.Reset()
	b.Grow(descriptionLimit)
	_, _ = b.WriteString(`**Caller Origin**`)
	_, _ = b.WriteString("\n```\n")
	_, _ = b.WriteString(msg.Caller().String())
	_, _ = b.WriteString("\n```\n")
	_, _ = b.WriteString(`**Caller Function**`)
	_, _ = b.WriteString("\n```\n")
	_, _ = b.WriteString(msg.Caller().ShortOrigin())
	_, _ = b.WriteString("\n```\n")
	_, _ = b.WriteString(`**Cache Key**`)
	_, _ = b.WriteString("\n```\n")
	_, _ = b.WriteString(extra.CacheKey)
	_, _ = b.WriteString("\n```")
	return d.shouldCreateFile(embed, b, "# Metadata")
}

func (d Discord) buildErrorStackEmbed(msg tower.MessageContext) (*Embed, bucket.File) {
	err := msg.Err()
	if err == nil {
		return nil, nil
	}
	s := make([]string, 0, 4)
	s = stackAccumulator(s, msg.Err())

	if len(s) == 0 {
		return nil, nil
	}
	reverse(s)
	content := strings.Join(s, "\n---\n")
	b := descBufPool.Get()
	defer descBufPool.Put(b)
	b.Reset()
	b.Grow(descriptionLimit)
	_, _ = b.WriteString("```")
	_, _ = b.WriteString(content)
	_, _ = b.WriteString("```")
	content = b.String()
	embed := &Embed{
		Type:  "rich",
		Title: "Error Stack",
		Color: 0x5d0e16, // Cardinal Red Dark
	}
	return d.shouldCreateFile(embed, b, "# Error Stack")
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

func reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func closingTicksTruncated(b *bytes.Buffer, countBack int) bool {
	buf := b.Bytes()
	if len(buf) >= countBack {
		buf = buf[len(buf)-countBack:]
	}
	count := bytes.Count(buf, []byte("```"))
	return count%2 == 0
}

func (d Discord) shouldCreateFile(embed *Embed, b *bytes.Buffer, pretext string) (em *Embed, file bucket.File) {
	content := b.String()
	if b.Len() > descriptionLimit {
		outro := "Content is too long to be displayed fully. See attachment for details"
		if closingTicksTruncated(b, len(outro)+5) {
			outro = "\n```\nContent is too long to be displayed fully. See attachment for details"
		}
		b.Truncate(descriptionLimit - len(outro))
		_, _ = b.WriteString(outro)
		embed.Description = b.String()
		buf := strings.NewReader(content)
		filename := d.snowflake.Generate().String() + ".md"
		file := bucket.NewFile(
			buf,
			"text/markdown",
			bucket.WithFilename(filename),
			bucket.WithPretext(pretext),
		)
		return embed, file
	}
	embed.Description = content
	return embed, nil
}
