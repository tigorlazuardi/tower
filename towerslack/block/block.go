package block

import "github.com/francoispqt/gojay"

type Blocks []Block

func (b Blocks) MarshalJSONArray(enc *gojay.Encoder) {
	for _, b := range b {
		enc.AddObject(b.Build())
	}
}

func (b Blocks) IsNil() bool {
	return len(b) == 0
}

type Block interface {
	// Prep this block for Marshaling.
	Build() gojay.MarshalerJSONObject
	// Removes all the element and release the associated elements into their own pool for reuse.
	Release()
}
