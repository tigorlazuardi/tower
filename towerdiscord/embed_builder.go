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

const discordLimit = 6000

func (d Discord) defaultEmbedBuilder(ctx context.Context, msg tower.MessageContext, extra *ExtraInformation) ([]*Embed, []bucket.File) {
	var (
		files  = make([]bucket.File, 0, 5)
		embeds = make([]*Embed, 0, 5)
		limit  = discordLimit - 150 // we have to take account for titles and timestamps.
	)
	summary, fileSummary, written := d.buildSummary(msg, 500, extra)
	limit -= written

	metadata, fileMetadata, written := d.buildMetadataEmbed(ctx, msg, extra, 500)
	limit -= written

	errorStackEmbed, fileErrorStack, written := d.buildErrorStackEmbed(msg, 1000, extra)
	limit -= written

	// Data limit is 50% of the remaining limit at max when error is available, otherwise 100% until 4096.
	dataLimit := limit
	if msg.Err() == nil && dataLimit > 4096 {
		dataLimit = 4096
	} else {
		dataLimit /= 2
	}

	dataEmbed, fileData, written := d.buildContextEmbed(msg, dataLimit, extra)
	limit -= written

	if limit > 4096 {
		limit = 4096
	}

	// Error will take the remaining limit.
	errorEmbed, errorData, written := d.buildErrorEmbed(msg, limit, extra)

	embeds = append(embeds, summary)
	if errorEmbed != nil {
		embeds = append(embeds, errorEmbed)
	}
	if dataEmbed != nil {
		embeds = append(embeds, dataEmbed)
	}
	if errorStackEmbed != nil {
		embeds = append(embeds, errorStackEmbed)
	}
	embeds = append(embeds, metadata)

	if fileSummary != nil {
		files = append(files, fileSummary)
	}
	if errorData != nil {
		files = append(files, errorData)
	}
	if fileData != nil {
		files = append(files, fileData)
	}
	if fileErrorStack != nil {
		files = append(files, fileErrorStack)
	}
	if fileMetadata != nil {
		files = append(files, fileMetadata)
	}
	return embeds, files
}

func (d Discord) buildSummary(msg tower.MessageContext, limit int, extra *ExtraInformation) (*Embed, bucket.File, int) {
	embed := &Embed{
		Type:  "rich",
		Title: "Summary",
		Color: 0x188544, // Green Jewel
	}
	display, data := descBufPool.Get(), descBufPool.Get()
	defer descBufPool.Put(display)
	defer descBufPool.Put(data)
	display.Reset()
	display.Grow(limit)
	data.Reset()
	data.Grow(limit)

	_, _ = display.WriteString("**")
	_, _ = display.WriteString(msg.Message())
	_, _ = display.WriteString("**")
	err := msg.Err()
	if err != nil {
		_, _ = display.WriteString("\n\n**Error**:\n")
		_, _ = display.WriteString("```\n")
		switch err := err.(type) {
		case tower.SummaryWriter:
			lw := tower.NewLineWriter(display).LineBreak("\n").Build()
			err.WriteSummary(lw)
		case tower.Summary:
			_, _ = display.WriteString(err.Summary())
		case tower.ErrorWriter:
			lw := tower.NewLineWriter(display).LineBreak("\n.. ").Build()
			err.WriteError(lw)
		default:
			_, _ = display.WriteString(err.Error())
		}
		_, _ = display.WriteString("\n```")
	}

	dataContext := msg.Context()
	if len(dataContext) > 0 {
		for _, c := range dataContext {
			switch c := c.(type) {
			case tower.SummaryWriter:
				_, _ = display.WriteString("\n\n**Context**:\n")
				_, _ = display.WriteString("```")
				if _, ok := c.(tower.Fields); ok {
					_, _ = display.WriteString("yaml")
				}
				_, _ = display.WriteString("\n")
				lw := tower.NewLineWriter(display).LineBreak("\n").Build()
				c.WriteSummary(lw)
				_, _ = display.WriteString("\n```")
			case tower.Summary:
				_, _ = display.WriteString("\n\n**Context**:\n")
				_, _ = display.WriteString("```\n")
				_, _ = display.WriteString(c.Summary())
				_, _ = display.WriteString("\n```")
			}
		}
	}
	if display.Len() > limit {
		_, _ = data.Write(display.Bytes())
	}

	return shouldCreateFile(&createFileContext{
		embed:          embed,
		display:        display,
		data:           data,
		contentType:    "text/markdown",
		fileExtension:  "md",
		suffixFilename: "_summary",
		limit:          limit,
		extra:          extra,
	})
}

//goland:noinspection GoUnhandledErrorResult
func (d Discord) buildContextEmbed(msg tower.MessageContext, limit int, extra *ExtraInformation) (*Embed, bucket.File, int) {
	if len(msg.Context()) == 0 {
		return nil, nil, 0
	}
	embed := &Embed{
		Type:  "rich",
		Title: "Context",
		Color: 0x063970, // Dark Blue
	}

	display, data := descBufPool.Get(), descBufPool.Get()
	defer descBufPool.Put(display)
	defer descBufPool.Put(data)
	display.Reset()
	display.Grow(limit)
	data.Reset()
	data.Grow(limit)

	contextData := msg.Context()
	err := d.codeBlockBuilder.Build(display, contextData)
	if err != nil {
		_, _ = display.WriteString("Error building context: ")
		display.WriteString("```")
		_, _ = display.WriteString(err.Error())
		display.WriteString("```\n")
	}
	if display.Len() > limit {
		var v any = contextData
		if len(msg.Context()) == 1 {
			v = contextData[0]
		}
		err := d.dataEncoder.Encode(data, v)
		if err != nil {
			display.WriteString("Error encoding context to file: ")
			display.WriteString("```")
			display.WriteString(err.Error())
			display.WriteString("```\n")
		}
	}

	return shouldCreateFile(&createFileContext{
		embed:          embed,
		display:        display,
		data:           data,
		contentType:    d.dataEncoder.ContentType(),
		fileExtension:  d.dataEncoder.FileExtension(),
		suffixFilename: "_context",
		limit:          limit,
		extra:          extra,
	})
}

func (d Discord) buildErrorEmbed(msg tower.MessageContext, limit int, extra *ExtraInformation) (*Embed, bucket.File, int) {
	err := msg.Err()
	if err == nil {
		return nil, nil, 0
	}
	embed := &Embed{
		Type:  "rich",
		Title: "Error",
		Color: 0x71010b, // Venetian Red
	}
	display, data := descBufPool.Get(), descBufPool.Get()
	defer descBufPool.Put(display)
	defer descBufPool.Put(data)
	display.Reset()
	display.Grow(limit)
	data.Reset()
	data.Grow(limit)
	if err := d.codeBlockBuilder.BuildError(display, err); err != nil {
		_, _ = display.WriteString("Error building error as display: ")
		_, _ = display.WriteString("```")
		_, _ = display.WriteString(err.Error())
		_, _ = display.WriteString("```\n")
	}
	if display.Len() > limit {
		err := d.dataEncoder.Encode(data, err)
		if err != nil {
			_, _ = display.WriteString("Error encoding error to file: ")
			_, _ = display.WriteString("```")
			_, _ = display.WriteString(err.Error())
			_, _ = display.WriteString("```\n")
		}
	}
	return shouldCreateFile(&createFileContext{
		embed:          embed,
		display:        display,
		data:           data,
		contentType:    d.dataEncoder.ContentType(),
		fileExtension:  d.dataEncoder.FileExtension(),
		suffixFilename: "_error",
		limit:          limit,
		extra:          extra,
	})
}

func (d Discord) buildMetadataEmbed(ctx context.Context, msg tower.MessageContext, extra *ExtraInformation, limit int) (*Embed, bucket.File, int) {
	count := 0
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
		count += len(v.Key) + len(v.Value)
	}
	service := msg.Service()
	if service.Name != "" {
		const name = "Service"
		embed.Fields = append(embed.Fields, &EmbedField{
			Name:   name,
			Value:  service.Name,
			Inline: true,
		})
		count += len(name) + len(service.Name)
	}
	if service.Type != "" {
		const sType = "Type"
		embed.Fields = append(embed.Fields, &EmbedField{
			Name:   sType,
			Value:  service.Type,
			Inline: true,
		})
		count += len(sType) + len(service.Type)
	}
	if service.Environment != "" {
		const env = "Environment"
		embed.Fields = append(embed.Fields, &EmbedField{
			Name:   env,
			Value:  service.Environment,
			Inline: true,
		})
		count += len(env) + len(service.Type)
	}
	const threadIDName = "Thread ID"
	embed.Fields = append(embed.Fields, &EmbedField{
		Name:   threadIDName,
		Value:  extra.ThreadID.String(),
		Inline: true,
	})
	count += len(threadIDName) + len(extra.ThreadID.String())
	var iteration string
	if msg.SkipVerification() {
		iteration = "(skipped verification)"
	} else {
		iteration = strconv.Itoa(extra.Iteration)
	}
	const messageIteration = "Message Iteration"
	embed.Fields = append(embed.Fields, &EmbedField{
		Name:   messageIteration,
		Value:  iteration,
		Inline: true,
	})
	count += len(messageIteration) + len(iteration)
	ts := extra.CooldownTimeEnds.Unix()
	const nextPossibleEarliestRepeat = "Next Possible Earliest Repeat"
	repeatValue := fmt.Sprintf("<t:%d:F> | <t:%d:R>", ts, ts)
	embed.Fields = append(embed.Fields, &EmbedField{
		Name:   nextPossibleEarliestRepeat,
		Value:  repeatValue,
		Inline: false,
	})
	count += len(messageIteration) + len(iteration)
	if len(embed.Fields) > 25 {
		embed.Fields = embed.Fields[:25]
	}
	display, data := descBufPool.Get(), descBufPool.Get()
	defer descBufPool.Put(display)
	defer descBufPool.Put(data)
	display.Reset()
	display.Grow(limit)
	data.Reset()
	data.Grow(limit)
	_, _ = display.WriteString(`**Caller Origin**`)
	_, _ = display.WriteString("\n```\n")
	_, _ = display.WriteString(msg.Caller().String())
	_, _ = display.WriteString("\n```\n")
	_, _ = display.WriteString(`**Caller Function**`)
	_, _ = display.WriteString("\n```\n")
	_, _ = display.WriteString(msg.Caller().ShortOrigin())
	_, _ = display.WriteString("\n```\n")
	_, _ = display.WriteString(`**Cache Key**`)
	_, _ = display.WriteString("\n```\n")
	_, _ = display.WriteString(extra.CacheKey)
	_, _ = display.WriteString("\n```")

	if display.Len() > limit {
		_, _ = data.Write(display.Bytes())
		_, _ = data.WriteString("\n```json\n")
		enc := json.NewEncoder(data)
		enc.SetIndent("", "    ")
		enc.SetEscapeHTML(false)
		_ = enc.Encode(embed.Fields)
		_, _ = data.WriteString("\n```")
	}

	embed, file, written := shouldCreateFile(&createFileContext{
		embed:          embed,
		display:        display,
		data:           bytes.NewBufferString(display.String()),
		contentType:    "text/markdown",
		fileExtension:  "md",
		suffixFilename: "_metadata",
		limit:          limit,
		extra:          extra,
	})
	count += written
	return embed, file, count
}

func (d Discord) buildErrorStackEmbed(msg tower.MessageContext, limit int, extra *ExtraInformation) (*Embed, bucket.File, int) {
	err := msg.Err()
	if err == nil {
		return nil, nil, 0
	}
	s := make([]string, 0, 4)
	s = stackAccumulator(s, msg.Err())

	if len(s) == 0 {
		return nil, nil, 0
	}
	reverse(s)
	content := strings.Join(s, "\n---\n")
	display, data := descBufPool.Get(), descBufPool.Get()
	defer descBufPool.Put(display)
	defer descBufPool.Put(data)
	display.Reset()
	display.Grow(limit)
	_, _ = display.WriteString("```")
	_, _ = display.WriteString(content)
	_, _ = display.WriteString("```")
	content = display.String()
	embed := &Embed{
		Type:  "rich",
		Title: "Error Stack",
		Color: 0x5d0e16, // Cardinal Red Dark
	}
	if display.Len() > limit {
		_, _ = data.Write(display.Bytes())
	}
	return shouldCreateFile(&createFileContext{
		embed:          embed,
		display:        display,
		data:           data,
		contentType:    "text/plain",
		fileExtension:  "txt",
		suffixFilename: "_error_stack",
		limit:          limit,
		extra:          extra,
	})
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
	return count%2 != 0
}

type createFileContext struct {
	embed          *Embed
	display        *bytes.Buffer
	data           *bytes.Buffer
	contentType    string
	fileExtension  string
	suffixFilename string
	limit          int
	extra          *ExtraInformation
}

func shouldCreateFile(ctx *createFileContext) (em *Embed, file bucket.File, written int) {
	display := ctx.display
	if display.Len() > ctx.limit {
		outro := "Content is too long to be displayed fully. See attachment for details"
		if closingTicksTruncated(display, len(outro)+5) {
			outro = "\n```\nContent is too long to be displayed fully. See attachment for details"
		}
		display.Truncate(ctx.limit - len(outro))
		display.WriteString(outro)
		ctx.embed.Description = display.String()

		filename := fmt.Sprintf("%s%s.%s", ctx.extra.ThreadID, ctx.suffixFilename, ctx.fileExtension)
		file = bucket.NewFile(
			ctx.data,
			ctx.contentType,
			bucket.WithFilename(filename),
		)
		return ctx.embed, file, display.Len()
	}
	ctx.embed.Description = display.String()
	return ctx.embed, nil, display.Len()
}
