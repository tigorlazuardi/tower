package towerdiscord

import (
	"encoding/json"
	"io"
)

type CodeBlockBuilder interface {
	Build(w io.Writer, value []any) error
	BuildError(w io.Writer, err error) error
}

type JSONCodeBlockBuilder struct{}

func (J JSONCodeBlockBuilder) Build(w io.Writer, value []any) error {
	_, err := io.WriteString(w, "```json\n")
	if err != nil {
		return err
	}
	defer func(w io.Writer, s string) {
		_, _ = io.WriteString(w, s)
	}(w, "```")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	var v any
	if len(value) == 1 {
		v = value[0]
	}
	return enc.Encode(v)
}

func (J JSONCodeBlockBuilder) BuildError(w io.Writer, e error) error {
	_, err := io.WriteString(w, "```json\n")
	if err != nil {
		return err
	}
	defer func(w io.Writer, s string) {
		_, _ = io.WriteString(w, s)
	}(w, "```")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(e)
}
