package tower

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"go.uber.org/zap/zapcore"
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

func (c Caller) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("source", fmt.Sprintf("%s:%d", c.File, c.Line))
	enc.AddString("origin", c.ShortOrigin())
	return nil
}

func (c Caller) getOrigin() []string {
	f := runtime.FuncForPC(c.PC)
	return strings.Split(f.Name(), "/")
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
