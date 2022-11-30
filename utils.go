package tower

import (
	"fmt"
)

func printDebugImpl(a any) bool {
	if a == nil {
		fmt.Println("nil")
		return true
	}
	if dbg, ok := a.(interface {
		Debug()
	}); ok {
		dbg.Debug()
		return true
	}
	return false
}

func printDisplayImpl(a any) bool {
	if a == nil {
		fmt.Println("nil")
		return true
	}
	if dbg, ok := a.(Display); ok {
		fmt.Println(dbg.Display())
		return true
	}
	return false
}

// Dbg Prints the variable and returns the given item after printing.
// Useful for Debugging without breaking the code flow.
func Dbg[T any](a T) T {
	if printDebugImpl(a) {
		return a
	}
	if printDisplayImpl(a) {
		return a
	}
	fmt.Printf("%#v\n", a)
	return a
}

// Cast turns any slice into a slice of any.
func Cast[T any](in []T) []any {
	out := make([]any, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}
