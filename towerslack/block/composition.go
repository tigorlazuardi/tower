package block

import "github.com/francoispqt/gojay"

type Composition interface {
	BuildComposition() gojay.MarshalerJSONObject
	// Removes all the element and release the associated elements into their own pool for reuse.
	Release()
}
