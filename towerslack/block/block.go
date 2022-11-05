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
	Build() gojay.MarshalerJSONObject
}
