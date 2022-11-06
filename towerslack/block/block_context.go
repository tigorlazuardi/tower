package block

import (
	"sync"

	"github.com/francoispqt/gojay"
)

var contextBlockPool = &sync.Pool{New: func() any { return &ContextBlock{} }}

var _ Block = (*ContextBlock)(nil)

type ContextBlock struct {
	Elements Elements
	BlockID  string
}

func (c *ContextBlock) Release() {
	for _, el := range c.Elements {
		el.Release()
	}
	c.Elements = c.Elements[:0]
	c.BlockID = ""
	contextBlockPool.Put(c)
}

func NewContextBlock(elements ...Element) *ContextBlock {
	cb := contextBlockPool.Get().(*ContextBlock) //nolint
	cb.Elements = elements
	return cb
}

func (b ContextBlock) Build() gojay.MarshalerJSONObject {
	return b
}

func (b ContextBlock) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddStringKey("type", "context")
	enc.AddArrayKey("elements", b.Elements)
	enc.AddStringKeyOmitEmpty("block_id", b.BlockID)
}

func (b ContextBlock) IsNil() bool {
	return len(b.Elements) == 0
}
