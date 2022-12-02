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
	encoder          Encoder
	transformer      BodyTransformer
	compressor       Compressor
	statusCode       int
	errorTransformer ErrorBodyTransformer
	callerDepth      int
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
		o.transformer = transformer
	})
}

func (OptionRespondGroup) ErrorTransformer(transformer ErrorBodyTransformer) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.errorTransformer = transformer
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

func (OptionRespondGroup) CallerDepth(i int) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.callerDepth = i
	})
}
