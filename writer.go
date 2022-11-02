package tower

import (
	"io"
	"strings"
)

type Writer interface {
	io.Writer
	io.StringWriter

	// Returns the accumulated values.
	String() string
}

type LineWriter interface {
	Writer
	// Writes a predetermined separator to the writer.
	WriteSeparator()
	// Writes a predetermined prefix to the writer.
	WritePrefix()
	// Writes a predetermined suffix to the writer.
	WriteSuffix()
	// Writes Indentation characters.
	WriteIndent()
	GetSeparator() string
	GetPrefix() string
	GetSuffix() string
	GetIndentation() string
}

type LineWriterBuilder struct {
	writer    Writer
	indent    string
	separator string
	prefix    string
	suffix    string
}

// Sets the Writer target.
func (builder *LineWriterBuilder) Writer(w Writer) *LineWriterBuilder {
	builder.writer = w
	return builder
}

// Sets the Indentation.
func (builder *LineWriterBuilder) Indent(s string) *LineWriterBuilder {
	builder.indent = s
	return builder
}

// Sets the Linebreak character(s).
func (builder *LineWriterBuilder) Separator(s string) *LineWriterBuilder {
	builder.separator = s
	return builder
}

// Sets the Prefix.
func (builder *LineWriterBuilder) Prefix(s string) *LineWriterBuilder {
	builder.prefix = s
	return builder
}

// Sets the Suffix.
func (builder *LineWriterBuilder) Suffix(s string) *LineWriterBuilder {
	builder.suffix = s
	return builder
}

// Turn this writer into proper LineWriter.
func (builder *LineWriterBuilder) Build() LineWriter {
	return &lineWriter{
		Writer:    builder.writer,
		separator: builder.separator,
		prefix:    builder.prefix,
		suffix:    builder.suffix,
		indent:    builder.indent,
	}
}

// Creates a new LineWriterBuilder. You have to call .Build() to actually use LineWriter.
func NewLineWriterBuilder() *LineWriterBuilder {
	return &LineWriterBuilder{
		writer: &strings.Builder{},
	}
}

var _ LineWriter = (*lineWriter)(nil)

type lineWriter struct {
	Writer
	separator string
	prefix    string
	suffix    string
	indent    string
}

// Writes Indentation characters.
func (l *lineWriter) WriteIndent() {
	if lw, ok := l.Writer.(LineWriter); ok {
		lw.WriteIndent()
	}
	_, _ = l.Writer.WriteString(l.indent)
}

func (l lineWriter) GetIndentation() string {
	if lw, ok := l.Writer.(LineWriter); ok {
		return l.indent + lw.GetIndentation()
	}
	return l.indent
}

func (l *lineWriter) WriteSeparator() {
	if lw, ok := l.Writer.(LineWriter); ok {
		lw.WriteSeparator()
	}
	_, _ = l.WriteString(l.separator)
}

func (l *lineWriter) WritePrefix() {
	if lw, ok := l.Writer.(LineWriter); ok {
		lw.WritePrefix()
	}
	_, _ = l.WriteString(l.prefix)
}

func (l *lineWriter) WriteSuffix() {
	if lw, ok := l.Writer.(LineWriter); ok {
		lw.WriteSuffix()
	}
	_, _ = l.WriteString(l.suffix)
}

func (l lineWriter) GetSeparator() string {
	if lw, ok := l.Writer.(LineWriter); ok {
		return lw.GetSeparator() + l.separator
	}
	return l.separator
}

func (l lineWriter) GetPrefix() string {
	if lw, ok := l.Writer.(LineWriter); ok {
		return lw.GetPrefix() + l.prefix
	}
	return l.prefix
}

func (l lineWriter) GetSuffix() string {
	if lw, ok := l.Writer.(LineWriter); ok {
		return lw.GetSuffix() + l.suffix
	}
	return l.suffix
}

type Display interface {
	// Display returns a human readable and rich with information for the implementer.
	Display() string
}

type DisplayWriter interface {
	// Writes the Display() string to the writer instead of being allocated as value.
	WriteDisplay(w LineWriter)
}

type ErrorWriter interface {
	// Writes the error.Error to the writer instead of being allocated as value.
	WriteError(w LineWriter)
}

type Summary interface {
	// Returns a short summary of the implementer.
	Summary() string
}

type SummaryWriter interface {
	// Writes the Summary() string to the writer instead of being allocated as value.
	WriteSummary(w Writer)
}
