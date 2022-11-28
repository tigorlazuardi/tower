package towerdiscord

import (
	"bytes"
	"encoding/json"
	"github.com/tigorlazuardi/tower"
	"io"
)

type CodeBlockBuilder interface {
	Build(w io.Writer, value []any) error
	BuildError(w io.Writer, err error) error
}

type JSONCodeBlockBuilder struct{}

type valueMarshaler []any

var _ tower.CodeBlockJSONMarshaler = (valueMarshaler)(nil)

func (v valueMarshaler) CodeBlockJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if len(v) == 1 {
		err := enc.Encode(v[0])
		return buf.Bytes(), err
	}
	j := make([]json.RawMessage, 0, len(v))
	for _, value := range v {
		buf.Reset()
		if vm, ok := value.(tower.CodeBlockJSONMarshaler); ok {
			raw, err := vm.CodeBlockJSON()
			if err != nil {
				return nil, err
			}
			j = append(j, raw)
			continue
		}
		err := enc.Encode(value)
		if err != nil {
			return nil, err
		}
		j = append(j, json.RawMessage(buf.String()))
	}
	buf.Reset()
	enc.SetIndent("", "  ")
	_ = enc.Encode(j)
	return buf.Bytes(), nil
}

func (J JSONCodeBlockBuilder) Build(w io.Writer, value []any) error {
	_, err := io.WriteString(w, "```json\n")
	if err != nil {
		return err
	}
	defer func(w io.Writer, s string) {
		_, _ = io.WriteString(w, s)
	}(w, "```")
	b, err := valueMarshaler(value).CodeBlockJSON()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func (J JSONCodeBlockBuilder) BuildError(w io.Writer, e error) error {
	_, err := io.WriteString(w, "```json\n")
	if err != nil {
		return err
	}
	defer func(w io.Writer, s string) {
		_, _ = io.WriteString(w, s)
	}(w, "```")
	if e, ok := e.(tower.CodeBlockJSONMarshaler); ok {
		b, err := e.CodeBlockJSON()
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		if err != nil {
			return err
		}
		return nil
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(e)
}
