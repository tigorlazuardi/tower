package block

import (
	"sync"

	"github.com/francoispqt/gojay"
)

var dividerBlockPool = &sync.Pool{New: func() any { return &DividerBlock{} }}

var _ Block = (*DividerBlock)(nil)

type DividerBlock struct {
	BlockID string
}

func NewDividerBlock() *DividerBlock { return dividerBlockPool.Get().(*DividerBlock) } //nolint

// Removes all the element and release the associated elements into their own pool for reuse.
func (c *DividerBlock) Release() {
	c.BlockID = ""
	dividerBlockPool.Put(c)
}

func (b DividerBlock) Build() gojay.MarshalerJSONObject {
	return b
}

func (b DividerBlock) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddStringKey("type", "divider")
	enc.AddStringKeyOmitEmpty("block_id", b.BlockID)
}

func (b DividerBlock) IsNil() bool {
	return false
}
