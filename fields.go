package tower

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func isJson(b []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(b, &js) == nil
}

// Writes the Display() string to the writer instead of being allocated as value.
func (f Fields) WriteDisplay(w LineWriter) {
	i := 0
	for k, v := range f {
		if i > 0 {
			w.WriteSeparator()
		}
		i++
		w.WritePrefix()
		w.WriteIndent()
		_, _ = w.WriteString(k)
		_, _ = w.WriteString(":")
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
			buf := &bytes.Buffer{}
			enc := json.NewEncoder(buf)
			enc.SetIndent(w.GetIndentation(), "    ")
			enc.SetEscapeHTML(false)
			err := enc.Encode(v)
			if err != nil {
				_, _ = w.WriteString(err.Error())
			} else {
				content := buf.Bytes()
				strip := bytes.TrimSpace(content)
				if !(len(strip) == 2 && (string(strip) == "{}" || string(strip) == "[]")) {
					w.WriteSeparator()
					w.WriteIndent()
					_, _ = w.Write(content)
					w.WriteSeparator()
					w.WriteIndent()
					_, _ = w.WriteString(k)
					_, _ = w.WriteString("_summary")
					_, _ = w.WriteString(":")
				}
			}
			_, _ = w.WriteString(" ")
			_, _ = w.WriteString(v.Error())
		case fmt.Stringer:
			_, _ = w.WriteString(" ")
			_, _ = w.WriteString(v.String())
		case string:
			if len(v) > 50 {
				w.WriteSeparator()
				w.WriteIndent()
			} else {
				_, _ = w.WriteString(" ")
			}
			_, _ = w.WriteString(v)
		case []byte:
			if isJson(v) {
				w.WriteSeparator()
				w.WriteIndent()
				buf := &bytes.Buffer{}
				err := json.Indent(buf, v, w.GetIndentation(), "    ")
				if err != nil {
					_, _ = w.WriteString(err.Error())
					continue
				}
				_, _ = io.Copy(w, buf)
				continue
			}
			if len(v) > 50 {
				w.WriteSeparator()
				w.WriteIndent()
			} else {
				_, _ = w.WriteString(" ")
			}
			_, _ = w.Write(v)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128:
			_, _ = w.WriteString(" ")
			_, _ = fmt.Fprintf(w, "%v", v)
		default:
			w.WriteSeparator()
			w.WriteIndent()
			enc := json.NewEncoder(w)
			enc.SetIndent(w.GetIndentation(), "    ")
			enc.SetEscapeHTML(false)
			err := enc.Encode(v)
			if err != nil {
				_, _ = w.WriteString(err.Error())
			}
		}
		w.WriteSuffix()
	}
}
