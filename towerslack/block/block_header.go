package block

import (
	"sync"

	"github.com/francoispqt/gojay"
)

var headerBlockPool = &sync.Pool{New: func() any { return &HeaderBlock{} }}

type HeaderBlock struct {
	Text    *TextComposition
	BlockID string
}

// Creates New HeaderBlock. Text with length higher than 150 will be truncated to that length.
func NewHeaderBlock(text string) *HeaderBlock {
	hb := headerBlockPool.Get().(*HeaderBlock) //nolint
	if len(text) > 150 {
		text = text[:150]
	}
	hb.Text = NewTextComposition(TextPlain, text)
	return hb
}

func (b HeaderBlock) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddStringKey("type", "header")
	enc.AddObjectKey("text", b.Text)
	enc.AddStringKeyOmitEmpty("block_id", b.BlockID)
}

func (b HeaderBlock) IsNil() bool {
	return b.Text.IsNil()
}

// Prep this block for Marshaling.
func (b HeaderBlock) Build() gojay.MarshalerJSONObject {
	return b
}

// Removes all the element and release the associated elements into their own pool for reuse.
func (b *HeaderBlock) Release() {
	b.Text.Release()
	b.Text = nil
	b.BlockID = ""
	headerBlockPool.Put(b)
}
