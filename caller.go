package tower

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

const sep = string(os.PathSeparator)

type Caller struct {
	PC   uintptr
	File string
	Line int
}

func (c Caller) Function() *runtime.Func {
	f := runtime.FuncForPC(c.PC)
	return f
}

func (c Caller) Origin() string {
	f := runtime.FuncForPC(c.PC)
	return f.Name()
}

func (c Caller) getOrigin() []string {
	return strings.Split(c.Origin(), "/")
}

// Gets only the latest four items maximum in the package path.
func (c Caller) ShortOrigin() string {
	s := c.getOrigin()

	for len(s) > 3 {
		s = s[1:]
	}

	return strings.Join(s, "/")
}

// Gets only the latest three items path in the File Path where the Caller comes from.
func (c Caller) ShortSource() string {
	s := strings.Split(c.File, sep)

	for len(s) > 3 {
		s = s[1:]
	}

	return strings.Join(s, sep)
}

// Like .String(), but runes other than letters, digits, `-` and `.` are set to `_`.
func (c Caller) FormatAsKey() string {
	s := &strings.Builder{}
	strLine := strconv.Itoa(c.Line)
	s.Grow(len(c.File) + 1 + len(strLine))
	replaceSymbols(s, c.File, '_')
	s.WriteRune('_')
	s.WriteString(strLine)
	return s.String()
}

// Sets this caller as `file_path:line` format.
func (c Caller) String() string {
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

// Gets the caller information for who calls this function. A value of 1 will return this GetCaller location.
// So you may want the value to be 2 or higher if you wrap this call in another function.
//
// Returns false when you ask out of bounds caller depth or golang has already garbage collected the stack information.
func GetCaller(depth int) (Caller, bool) {
	pc, file, line, ok := runtime.Caller(depth)
	if !ok {
		return Caller{}, false
	}

	return Caller{
		PC:   pc,
		File: file,
		Line: line,
	}, true
}
