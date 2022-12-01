package towerhttp

import (
	"github.com/tigorlazuardi/tower"
)

// Responder handles the response and writing to http.ResponseWriter.
type Responder struct {
	encoder          Encoder
	transformer      BodyTransformer
	errorTransformer ErrorBodyTransformer
	tower            *tower.Tower
	compressor       Compression
	callerDepth      int
}

// NewResponder creates a new Responder instance.
//
// It has the following default values:
//
// - Encoder: JSONEncoder (encodes to JSON)
//
// - BodyTransformer: NoopBodyTransform (does nothing to whatever value you pass in)
//
// - ErrorBodyTransformer: SimpleErrorTransformer (encodes error to {"error": "message/err.Error()"}) with JSONEncoder.
// Different encoder may have different output.
//
// - Tower: points to the global tower instance
//
// - Compression: NoCompression.
func NewResponder() *Responder {
	return &Responder{
		encoder:          NewJSONEncoder(),
		transformer:      NoopBodyTransform{},
		errorTransformer: SimpleErrorTransformer{},
		tower:            tower.Global.Tower(),
		compressor:       NoCompression{},
		callerDepth:      3,
	}
}

// SetErrorTransformer sets the ErrorBodyTransformer to be used by the Responder.
func (r *Responder) SetErrorTransformer(errorTransformer ErrorBodyTransformer) {
	r.errorTransformer = errorTransformer
}

// SetEncoder sets the encoder to be used by the Responder.
func (r *Responder) SetEncoder(encoder Encoder) {
	r.encoder = encoder
}

// SetBodyTransformer sets the BodyTransformer to be used by the Responder.
func (r *Responder) SetBodyTransformer(transform BodyTransformer) {
	r.transformer = transform
}

// SetTower sets the tower instance to be used by the Responder.
func (r *Responder) SetTower(t *tower.Tower) {
	r.tower = t
}

// SetCompressor sets the compression to be used by the Responder.
func (r *Responder) SetCompressor(compressor Compression) {
	r.compressor = compressor
}

// SetCallerDepth sets the caller depth to be used to get caller function by the Responder.
func (r *Responder) SetCallerDepth(depth int) {
	r.callerDepth = depth
}

func (r Responder) buildOption(statusCode int) *respondOption {
	opt := &respondOption{
		encoder:     r.encoder,
		transformer: r.transformer,
		compressor:  r.compressor,
		statusCode:  statusCode,
	}
	return opt
}
