// Code generated by Gojay. DO NOT EDIT.

package slackrest

import (
	"sync"

	"github.com/francoispqt/gojay"
)

func init() {
	MessageResponsePool = &sync.Pool{
		New: func() interface{} {
			return &MessageResponse{}
		},
	}
	MessageResponseItemPool = &sync.Pool{
		New: func() interface{} {
			return &MessageResponseItem{}
		},
	}
}

var (
	MessageResponsePool     *sync.Pool
	MessageResponseItemPool *sync.Pool
)

// MarshalJSONObject implements MarshalerJSONObject
func (r *MessageResponse) MarshalJSONObject(enc *gojay.Encoder) {
	enc.BoolKey("ok", r.Ok)
	enc.StringKey("channel", r.Channel)
	enc.StringKey("ts", r.Ts)
	enc.ObjectKey("message", &r.Message)
}

// IsNil checks if instance is nil
func (r *MessageResponse) IsNil() bool {
	return r == nil
}

// UnmarshalJSONObject implements gojay's UnmarshalerJSONObject
func (r *MessageResponse) UnmarshalJSONObject(dec *gojay.Decoder, k string) error {
	switch k {
	case "ok":
		return dec.Bool(&r.Ok)

	case "channel":
		return dec.String(&r.Channel)

	case "ts":
		return dec.String(&r.Ts)

	case "message":
		err := dec.Object(&r.Message)

		return err

	}
	return nil
}

// NKeys returns the number of keys to unmarshal
func (r *MessageResponse) NKeys() int { return 4 }

// Reset reset fields
func (r *MessageResponse) Reset() {
	r.Ok = false
	r.Channel = ""
	r.Ts = ""
}

// MarshalJSONObject implements MarshalerJSONObject
func (i *MessageResponseItem) MarshalJSONObject(enc *gojay.Encoder) {
	enc.StringKey("text", i.Text)
	enc.StringKey("username", i.Username)
	enc.StringKey("bot_id", i.BotID)
	attachmentsSlice := gojay.EmbeddedJSON(i.Attachments)
	enc.AddEmbeddedJSONKey("attachments", &attachmentsSlice)
	enc.StringKey("type", i.Type)
	enc.StringKey("subtype", i.Subtype)
	enc.StringKey("ts", i.Ts)
}

// IsNil checks if instance is nil
func (i *MessageResponseItem) IsNil() bool {
	return i == nil
}

// UnmarshalJSONObject implements gojay's UnmarshalerJSONObject
func (i *MessageResponseItem) UnmarshalJSONObject(dec *gojay.Decoder, k string) error {
	switch k {
	case "text":
		return dec.String(&i.Text)

	case "username":
		return dec.String(&i.Username)

	case "bot_id":
		return dec.String(&i.BotID)

	case "attachments":
		value := gojay.EmbeddedJSON{}
		err := dec.AddEmbeddedJSON(&value)
		if err == nil && len(value) > 0 {
			i.Attachments = []byte(value)
		}
		return err

	case "type":
		return dec.String(&i.Type)

	case "subtype":
		return dec.String(&i.Subtype)

	case "ts":
		return dec.String(&i.Ts)

	}
	return nil
}

// NKeys returns the number of keys to unmarshal
func (i *MessageResponseItem) NKeys() int { return 7 }

// Reset reset fields
func (i *MessageResponseItem) Reset() {
	i.Text = ""
	i.Username = ""
	i.BotID = ""
	i.Attachments = nil
	i.Type = ""
	i.Subtype = ""
	i.Ts = ""
}
