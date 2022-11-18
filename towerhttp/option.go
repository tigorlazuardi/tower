package towerhttp

type ro int

// RO holds all the available options for responding with towerhttp.
const RO ro = 0

type RespondOption interface {
	apply(*option)
}

type RespondOptionFunc func(*option)

func (r RespondOptionFunc) apply(o *option) {
	r(o)
}

type option struct {
	encoder    Encoder
	transfomer BodyTransform
	compressor Compression
	statusCode int
}

// Encoder overrides the encoder to be used for encoding the response body.
func (ro) Encoder(encoder Encoder) RespondOption {
	return RespondOptionFunc(func(o *option) {
		o.encoder = encoder
	})
}

// Transformer overrides the transformer to be used for transforming the response body.
func (ro) Transformer(transformer BodyTransform) RespondOption {
	return RespondOptionFunc(func(o *option) {
		o.transfomer = transformer
	})
}

// Compressor overrides the compressor to be used for compressing the response body.
func (ro) Compressor(compressor Compression) RespondOption {
	return RespondOptionFunc(func(o *option) {
		o.compressor = compressor
	})
}

// StatusCode overrides the status code to be used for the response.
func (ro) StatusCode(i int) RespondOption {
	return RespondOptionFunc(func(o *option) {
		o.statusCode = i
	})
}
