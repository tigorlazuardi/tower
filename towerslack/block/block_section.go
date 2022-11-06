package block

import (
	"sync"

	"github.com/francoispqt/gojay"
)

var sectionBlockPool = &sync.Pool{New: func() any {
	return &SectionBlock{}
}}

type SectionBlock struct {
	Text      *TextComposition
	Fields    []*TextComposition
	BlockID   string
	Accessory Element
}

func NewSectionBlockText(t TextType, text string) *SectionBlock {
	sb := sectionBlockPool.Get().(*SectionBlock) //nolint
	sb.Text = NewTextComposition(t, text)
	return sb
}

func NewSectionBlockFields(t TextType, texts ...string) *SectionBlock {
	sb := sectionBlockPool.Get().(*SectionBlock) //nolint
	for _, text := range texts {
		sb.Fields = append(sb.Fields, NewTextComposition(t, text))
	}
	return sb
}

func (s SectionBlock) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddStringKey("type", "section")
	if s.Text != nil {
		enc.AddObjectKey("text", s.Text)
	}
	if len(s.Fields) > 0 {
		enc.AddArrayKey("fields", gojay.EncodeArrayFunc(func(e *gojay.Encoder) {
			for _, v := range s.Fields {
				e.AddObject(v)
			}
		}))
	}
	enc.AddStringKeyOmitEmpty("block_id", s.BlockID)
	if s.Accessory != nil {
		enc.AddObjectKey("accessory", s.Accessory.BuildElement())
	}
}

func (s SectionBlock) IsNil() bool {
	return s.Text == nil && len(s.Fields) == 0
}

// Prep this block for Marshaling.
func (s SectionBlock) Build() gojay.MarshalerJSONObject {
	return s
}

// Removes all the element and release the associated elements into their own pool for reuse.
func (s *SectionBlock) Release() {
	if s.Text != nil {
		s.Text.Release()
		s.Text = nil
	}
	if s.Fields != nil {
		for _, v := range s.Fields {
			v.Release()
		}
		s.Fields = s.Fields[:0]
	}

	s.BlockID = ""
	if s.Accessory != nil {
		s.Accessory.Release()
		s.Accessory = nil
	}
	sectionBlockPool.Put(s)
}
