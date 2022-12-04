package towerhttp

type RespondOption interface {
	apply(*respondOption)
}

type RespondOptionFunc func(*respondOption)

func (r RespondOptionFunc) apply(o *respondOption) {
	r(o)
}

type OptionRespondGroup struct{}

type respondOption struct {
	encoder              Encoder
	bodyTransformer      BodyTransformer
	compressor           Compressor
	statusCode           int
	errorBodyTransformer ErrorBodyTransformer
	callerDepth          int
}

// Encoder overrides the encoder to be used for encoding the response body.
func (OptionRespondGroup) Encoder(encoder Encoder) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.encoder = encoder
	})
}

// Transformer overrides the transformer to be used for transforming the response body.
func (OptionRespondGroup) Transformer(transformer BodyTransformer) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.bodyTransformer = transformer
	})
}

func (OptionRespondGroup) ErrorTransformer(transformer ErrorBodyTransformer) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.errorBodyTransformer = transformer
	})
}

// Compressor overrides the compressor to be used for compressing the response body.
func (OptionRespondGroup) Compressor(compressor Compressor) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.compressor = compressor
	})
}

// StatusCode overrides the status code to be used for the response.
func (OptionRespondGroup) StatusCode(i int) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.statusCode = i
	})
}

// CallerSkip overrides the caller skip to be used for the response to get the caller information.
func (OptionRespondGroup) CallerSkip(i int) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.callerDepth = i
	})
}

// AddCallerSkip adds the caller skip value to be used for the response to get the caller information.
func (OptionRespondGroup) AddCallerSkip(i int) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.callerDepth += i
	})
}
