package block

import (
	"sync"

	"github.com/francoispqt/gojay"
)

var textCompositionPool = &sync.Pool{New: func() any { return &TextComposition{} }}

type TextType string

const (
	TextMrkdwn TextType = "mrkdwn"
	TextPlain  TextType = "plain_text"
)

// See https://api.slack.com/reference/block-kit/composition-objects#text for details.
type TextComposition struct {
	Type     TextType
	Text     string
	Emoji    bool
	Verbatim bool
}

func NewTextComposition(t TextType, text string) *TextComposition {
	tc := textCompositionPool.Get().(*TextComposition) //nolint
	tc.Type = t
	tc.Text = text
	tc.Emoji = true
	return tc
}

func (t TextComposition) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddStringKey("type", string(t.Type))
	enc.AddStringKey("text", t.Text)
	enc.AddBoolKeyOmitEmpty("emoji", t.Emoji)
	enc.AddBoolKeyOmitEmpty("verbatim", t.Verbatim)
}

func (t TextComposition) IsNil() bool {
	return len(t.Text) == 0
}

func (t TextComposition) BuildComposition() gojay.MarshalerJSONObject {
	return t
}

// Removes all the element and release the associated elements into their own pool for reuse.
func (t *TextComposition) Release() {
	t.Type = ""
	t.Text = ""
	t.Emoji = false
	t.Verbatim = false
	textCompositionPool.Put(t)
}
