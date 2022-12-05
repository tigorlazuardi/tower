package towerhttp

import "github.com/tigorlazuardi/tower"

type RespondOption interface {
	apply(*RespondContext)
}

type RespondOptionFunc func(*RespondContext)

func (r RespondOptionFunc) apply(o *RespondContext) {
	r(o)
}

type OptionRespondGroup struct{}

type RespondContext struct {
	Encoder              Encoder
	BodyTransformer      BodyTransformer
	Compressor           Compressor
	StatusCode           int
	ErrorBodyTransformer ErrorBodyTransformer
	CallerDepth          int
	Caller               tower.Caller
}

// Encoder overrides the Encoder to be used for encoding the response body.
func (OptionRespondGroup) Encoder(encoder Encoder) RespondOption {
	return RespondOptionFunc(func(o *RespondContext) {
		o.Encoder = encoder
	})
}

// Transformer overrides the transformer to be used for transforming the response body.
func (OptionRespondGroup) Transformer(transformer BodyTransformer) RespondOption {
	return RespondOptionFunc(func(o *RespondContext) {
		o.BodyTransformer = transformer
	})
}

func (OptionRespondGroup) ErrorTransformer(transformer ErrorBodyTransformer) RespondOption {
	return RespondOptionFunc(func(o *RespondContext) {
		o.ErrorBodyTransformer = transformer
	})
}

// Compressor overrides the Compressor to be used for compressing the response body.
func (OptionRespondGroup) Compressor(compressor Compressor) RespondOption {
	return RespondOptionFunc(func(o *RespondContext) {
		o.Compressor = compressor
	})
}

// StatusCode overrides the status code to be used for the response.
func (OptionRespondGroup) StatusCode(i int) RespondOption {
	return RespondOptionFunc(func(o *RespondContext) {
		o.StatusCode = i
	})
}

// CallerSkip overrides the caller skip to be used for the response to get the caller information.
func (OptionRespondGroup) CallerSkip(i int) RespondOption {
	return RespondOptionFunc(func(o *RespondContext) {
		o.CallerDepth = i
	})
}

// AddCallerSkip adds the caller skip value to be used for the response to get the caller information.
func (OptionRespondGroup) AddCallerSkip(i int) RespondOption {
	return RespondOptionFunc(func(o *RespondContext) {
		o.CallerDepth += i
	})
}
