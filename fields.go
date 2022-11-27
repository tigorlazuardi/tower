package tower

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Fields map[string]any

// F Alias to tower.Fields.
type F = Fields

var (
	_ Summary       = (Fields)(nil)
	_ SummaryWriter = (Fields)(nil)
)

// Summary Returns a short summary of this type.
func (f Fields) Summary() string {
	s := &strings.Builder{}
	lw := NewLineWriter(s).LineBreak("\n").Build()
	f.WriteSummary(lw)
	return s.String()
}

// WriteSummary Writes the Summary() string to the writer instead of being allocated as value.
func (f Fields) WriteSummary(w LineWriter) {
	prefixLength := 0
	for k := range f {
		if prefixLength < len(k) {
			prefixLength = len(k)
		}
	}
	i := 0
	for k, v := range f {
		if i > 0 {
			w.WriteLineBreak()
		}
		i++

		w.WriteIndent()
		w.WritePrefix()
		_, _ = fmt.Fprintf(w, "%-*s: ", prefixLength, k)
		if v == nil {
			_, _ = w.WriteString("null")
			w.WriteSuffix()
			continue
		}
		switch v := v.(type) {
		case SummaryWriter:
			w.WriteLineBreak()
			v.WriteSummary(NewLineWriter(w).Indent("    ").Build())
		case Summary:
			_, _ = w.WriteString(v.Summary())
		case fmt.Stringer:
			_, _ = w.WriteString(v.String())
		case json.RawMessage:
			if len(v) <= 32 {
				_, _ = w.Write(v)
			} else {
				_, _ = w.WriteString("[object]")
			}
		case []byte:
			if len(v) <= 32 {
				_, _ = w.Write(v)
			} else {
				_, _ = w.WriteString("[object]")
			}
		case string:
			_, _ = w.WriteString(v)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128:
			_, _ = fmt.Fprintf(w, "%v", v)
		default:
			_, _ = w.WriteString("[object]")
		}
		w.WriteSuffix()
	}
}
