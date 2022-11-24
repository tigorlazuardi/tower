package towerhttp

type option struct{}

// Option holds all the available options for responding with towerhttp.
var Option option

type RespondOption interface {
	apply(*respondOption)
}

type RespondOptionFunc func(*respondOption)

func (r RespondOptionFunc) apply(o *respondOption) {
	r(o)
}

type respondOption struct {
	encoder    Encoder
	transfomer BodyTransformer
	compressor Compression
	statusCode int
}

// Encoder overrides the encoder to be used for encoding the response body.
func (option) Encoder(encoder Encoder) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.encoder = encoder
	})
}

// Transformer overrides the transformer to be used for transforming the response body.
func (option) Transformer(transformer BodyTransformer) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.transfomer = transformer
	})
}

// Compressor overrides the compressor to be used for compressing the response body.
func (option) Compressor(compressor Compression) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.compressor = compressor
	})
}

// StatusCode overrides the status code to be used for the response.
func (option) StatusCode(i int) RespondOption {
	return RespondOptionFunc(func(o *respondOption) {
		o.statusCode = i
	})
}
