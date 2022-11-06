package block

import (
	"sync"

	"github.com/francoispqt/gojay"
)

var fileBlockPool = &sync.Pool{New: func() any { return &FileBlock{} }}

var _ Block = (*FileBlock)(nil)

type FileBlock struct {
	ExternalID string
	BlockID    string
}

// Removes all the element and release the associated elements into their own pool for reuse.
func (b *FileBlock) Release() {
	b.ExternalID = ""
	b.BlockID = ""
	fileBlockPool.Put(b)
}

func NewFileBlock(externalID string) *FileBlock {
	fb := fileBlockPool.Get().(*FileBlock) // nolint
	fb.ExternalID = externalID
	return fb
}

func (b FileBlock) Build() gojay.MarshalerJSONObject {
	return b
}

func (b FileBlock) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddStringKey("type", "file")
	enc.AddStringKey("source", "remote")
	enc.AddStringKey("external_id", b.ExternalID)
	enc.AddStringKeyOmitEmpty("block_id", b.BlockID)
}

func (b FileBlock) IsNil() bool {
	return b.ExternalID == ""
}
