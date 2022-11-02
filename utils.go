package tower

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Prints the variable and returns the given item after printing.
// Useful for Debugging without breaking the code flow.
func Dbg[T any](a T) T {
	str := &strings.Builder{}
	enc := json.NewEncoder(str)
	enc.SetIndent("", "    ")
	enc.SetEscapeHTML(false)
	err := enc.Encode(a)
	if err != nil {
		str.WriteString(err.Error())
	}
	fmt.Println(str.String())
	return a
}
