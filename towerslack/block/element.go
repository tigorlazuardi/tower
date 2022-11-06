package block

import "github.com/francoispqt/gojay"

type Elements []Element

func (e Elements) MarshalJSONArray(enc *gojay.Encoder) {
	for _, v := range e {
		enc.AddObject(v.BuildElement())
	}
}

func (e Elements) IsNil() bool {
	return len(e) == 0
}

type Element interface {
	BuildElement() gojay.MarshalerJSONObject
	// Removes all the element and release the associated elements into their own pool for reuse.
	Release()
}
