package tower

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

type Fields map[string]any

var (
	_ Display       = (Fields)(nil)
	_ DisplayWriter = (Fields)(nil)
)

// Returns a short summary of this type.
func (f Fields) Summary() string {
	s := NewLineWriterBuilder().Separator("\n").Build()
	f.WriteSummary(s)
	return s.String()
}

// Writes the Summary() string to the writer instead of being allocated as value.
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
			w.WriteSeparator()
		}
		i++

		w.WriteIndent()
		w.WritePrefix()
		fmt.Fprintf(w, "%-*s : ", prefixLength, k)
		switch v := v.(type) {
		case SummaryWriter:
			w.WriteSeparator()
			v.WriteSummary(NewLineWriterBuilder().Writer(w).Indent("    ").Build())
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

// Display returns a human readable and rich with information for the implementer.
func (f Fields) Display() string {
	s := NewLineWriterBuilder().Separator("\n").Build()
	f.WriteDisplay(s)
	return s.String()
}

// Writes the Display() string to the writer instead of being allocated as value.
func (f Fields) WriteDisplay(w LineWriter) {
	keys := getSortedKeys(f)
	for i, k := range keys {
		v := f[k]
		if i > 0 {
			w.WriteSeparator()
		}
		i++
		w.WritePrefix()
		w.WriteIndent()
		// support for 0 length keys. For debugging utilities.
		if len(k) > 0 {
			_, _ = w.WriteString(k)
			_, _ = w.WriteString(":")
		}
		switch v := v.(type) {
		case DisplayWriter:
			w.WriteSeparator()
			v.WriteDisplay(NewLineWriterBuilder().Writer(w).Indent("    ").Build())
		case Display:
			dsp := v.Display()
			if len(dsp) > 50 {
				w.WriteSeparator()
				w.WritePrefix()
				w.WriteIndent()
			} else {
				_, _ = w.WriteString(" ")
			}
			_, _ = w.WriteString(v.Display())
		case ErrorWriter:
			w.WriteSeparator()
			v.WriteError(NewLineWriterBuilder().Writer(w).Indent("    ").Build())
		case error:
			indented := NewLineWriterBuilder().Writer(w).Indent("    ").Build()
			buf := &bytes.Buffer{}
			enc := json.NewEncoder(buf)
			enc.SetIndent(w.GetPrefix()+indented.GetIndentation(), "    ")
			enc.SetEscapeHTML(false)
			err := enc.Encode(v)
			if err != nil {
				_, _ = w.WriteString(err.Error())
			} else {
				content := buf.Bytes()
				strip := bytes.TrimSpace(content)
				if !(len(strip) == 2 && (string(strip) == "{}" || string(strip) == "[]")) {
					indented.WriteSeparator()
					indented.WritePrefix()
					indented.WriteIndent()
					_, _ = indented.Write(bytes.TrimSpace(content))
					w.WriteSeparator()
					w.WritePrefix()
					w.WriteIndent()
					_, _ = w.WriteString(k)
					_, _ = w.WriteString("_summary")
					_, _ = w.WriteString(":")
				}
			}
			str := v.Error()
			if len(str) > 50 {
				indented.WriteSeparator()
				indented.WritePrefix()
				indented.WriteIndent()
			} else {
				_, _ = w.WriteString(" ")
			}
			_, _ = w.WriteString(str)
		case fmt.Stringer:
			str := v.String()
			if len(str) > 50 {
				w := NewLineWriterBuilder().Writer(w).Indent("    ").Build()
				w.WriteSeparator()
				w.WritePrefix()
				w.WriteIndent()
			} else {
				_, _ = w.WriteString(" ")
			}
			_, _ = w.WriteString(v.String())
		case string:
			if len(v) > 50 {
				w := NewLineWriterBuilder().Writer(w).Indent("    ").Build()
				w.WriteSeparator()
				w.WritePrefix()
				w.WriteIndent()
			} else {
				_, _ = w.WriteString(" ")
			}
			_, _ = w.WriteString(v)
		case []byte:
			if isJson(v) {
				w := NewLineWriterBuilder().Writer(w).Indent("    ").Build()
				w.WriteSeparator()
				w.WritePrefix()
				w.WriteIndent()
				buf := &bytes.Buffer{}
				_ = json.Indent(buf, v, w.GetPrefix()+w.GetIndentation(), "    ")
				_, _ = io.Copy(w, buf)
				continue
			}
			if len(v) > 50 {
				w := NewLineWriterBuilder().Writer(w).Indent("    ").Build()
				w.WriteSeparator()
				w.WritePrefix()
				w.WriteIndent()
			} else {
				_, _ = w.WriteString(" ")
			}
			_, _ = w.Write(v)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128:
			_, _ = w.WriteString(" ")
			_, _ = fmt.Fprintf(w, "%v", v)
		default:
			w := NewLineWriterBuilder().Writer(w).Indent("    ").Build()
			w.WriteSeparator()
			w.WritePrefix()
			w.WriteIndent()
			buf := &bytes.Buffer{}
			enc := json.NewEncoder(buf)
			enc.SetIndent(w.GetPrefix()+w.GetIndentation(), "    ")
			enc.SetEscapeHTML(false)
			err := enc.Encode(v)
			if err != nil {
				_, _ = w.WriteString(err.Error())
			} else {
				_, _ = w.Write(bytes.TrimSpace(buf.Bytes()))
			}
		}
		w.WriteSuffix()
	}
}

func isJson(b []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(b, &js) == nil
}

func getSortedKeys[K Ordered, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
