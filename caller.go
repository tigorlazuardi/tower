package tower

import (
	"encoding/json"
	"os"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

const sep = string(os.PathSeparator)

type Caller interface {
	Function() *runtime.Func
	Origin() string
	ShortOrigin() string
	ShortSource() string
	String() string
	FormatAsKey() string
}

type caller struct {
	PC   uintptr
	File string
	Line int
}

func (c caller) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c caller) Function() *runtime.Func {
	f := runtime.FuncForPC(c.PC)
	return f
}

func (c caller) Origin() string {
	f := runtime.FuncForPC(c.PC)
	return f.Name()
}

func (c caller) getOrigin() []string {
	return strings.Split(c.Origin(), "/")
}

// ShortOrigin returns only the latest four items maximum in the package path.
func (c caller) ShortOrigin() string {
	s := c.getOrigin()

	for len(s) > 3 {
		s = s[1:]
	}

	return strings.Join(s, "/")
}

// ShortSource returns only the latest three items path in the File Path where the Caller comes from.
func (c caller) ShortSource() string {
	s := strings.Split(c.File, sep)

	for len(s) > 3 {
		s = s[1:]
	}

	return strings.Join(s, sep)
}

// FormatAsKey Like .String(), but runes other than letters, digits, `-` and `.` are set to `_`.
func (c caller) FormatAsKey() string {
	s := &strings.Builder{}
	strLine := strconv.Itoa(c.Line)
	s.Grow(len(c.File) + 1 + len(strLine))
	replaceSymbols(s, c.File, '_')
	s.WriteRune('_')
	s.WriteString(strLine)
	return s.String()
}

// String Sets this caller as `file_path:line` format.
func (c caller) String() string {
	s := &strings.Builder{}
	strLine := strconv.Itoa(c.Line)
	s.Grow(len(c.File) + 1 + len(strLine))
	s.WriteString(c.File)
	s.WriteRune(':')
	s.WriteString(strLine)
	return s.String()
}

func replaceSymbols(builder *strings.Builder, s string, rep rune) {
	for _, c := range s {
		switch {
		case unicode.In(c, unicode.Digit, unicode.Letter), c == '-', c == '.':
			builder.WriteRune(c)
		default:
			builder.WriteRune(rep)
		}
	}
}

// GetCaller returns the caller information for who calls this function. A value of 1 will return this GetCaller location.
// So you may want the value to be 2 or higher if you wrap this call in another function.
//
// Returns zero value if the caller information cannot be obtained.
func GetCaller(depth int) Caller {
	pc, file, line, _ := runtime.Caller(depth)
	return &caller{
		PC:   pc,
		File: file,
		Line: line,
	}
}
