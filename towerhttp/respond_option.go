package towerhttp

type RespondOption interface {
	apply(*respondOption)
}

type RespondOptionFunc func(*respondOption)

func (r RespondOptionFunc) apply(o *respondOption) {
	r(o)
}

type OptionRespondGroup interface {
	Encoder(encoder Encoder) RespondOption
	Transformer(transformer BodyTransformer) RespondOption
	Compressor(compressor Compressor) RespondOption
	StatusCode(i int) RespondOption
}

type optionRespondGroup struct{}

type respondOption struct {
	encoder     Encoder
	transformer BodyTransformer
	compressor  Compressor
	statusCode  int
}

// Encoder overrides the encoder to be used for encoding the response body.
func (optionRespondGroup) Encoder(encoder Encoder) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.encoder = encoder
	})
}

// Transformer overrides the transformer to be used for transforming the response body.
func (optionRespondGroup) Transformer(transformer BodyTransformer) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.transformer = transformer
	})
}

// Compressor overrides the compressor to be used for compressing the response body.
func (optionRespondGroup) Compressor(compressor Compressor) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.compressor = compressor
	})
}

// StatusCode overrides the status code to be used for the response.
func (optionRespondGroup) StatusCode(i int) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.statusCode = i
	})
}
