package towerhttp

import (
	"context"
	"github.com/tigorlazuardi/tower"
	"io"
	"net/http"
)

// Responder handles the response and writing to http.ResponseWriter.
type Responder struct {
	encoder     Encoder
	transformer BodyTransform
	tower       *tower.Tower
	compressor  Compression
}

// SetEncoder sets the encoder to be used by the Responder.
func (r *Responder) SetEncoder(encoder Encoder) {
	r.encoder = encoder
}

// SetTransformer sets the BodyTransform to be used by the Responder.
func (r *Responder) SetTransformer(transform BodyTransform) {
	r.transformer = transform
}

// SetTower sets the tower instance to be used by the Responder.
func (r *Responder) SetTower(tower *tower.Tower) {
	r.tower = tower
}

// SetCompressor sets the compression to be used by the Responder.
func (r *Responder) SetCompressor(compressor Compression) {
	r.compressor = compressor
}

func (r Responder) buildOption(statusCode int) *option {
	opt := &option{
		encoder:    r.encoder,
		transfomer: r.transformer,
		compressor: r.compressor,
		statusCode: statusCode,
	}
	return opt
}

// Respond with the given body and options.
//
// body is expected to be a serializable type. For streams, use RespondStream.
//
// HTTP status by default is http.StatusOK. If body implements tower.HTTPCodeHint, the status code will be set to the
// value returned by the tower.HTTPCodeHint method. If the towerhttp.RO.StatusCode RespondOption is set, it will override
// the status regardless of the tower.HTTPCodeHint.
//
// There's a special case if you pass http.NoBody as body, there will be no respond body related operations executed.
// StatusCode default value is set to http.StatusNoContent. You can still override this output by setting the
// related RespondOption.
// With http.NoBody as body, Towerhttp will immediately respond with status code after RespondOption are evaluated
// and end the process.
//
// Body of nil has different treatment with http.NoBody. if body is nil, default status code is still http.StatusOK, and
// the nil value is still passed to the BodyTransform implementer, therefore the final result body may not actually be empty.
func (r Responder) Respond(ctx context.Context, rw http.ResponseWriter, body any, opts ...RespondOption) {
	var (
		statusCode = http.StatusOK
		err        error
	)

	if ch, ok := body.(tower.HTTPCodeHint); ok {
		statusCode = ch.HTTPCode()
	}

	if body == http.NoBody {
		statusCode = http.StatusNoContent
	}

	opt := r.buildOption(statusCode)
	for _, o := range opts {
		o.apply(opt)
	}
	if body == http.NoBody {
		rw.WriteHeader(opt.statusCode)
		return
	}

	body = opt.transfomer.BodyTransform(ctx, body)
	if body == nil {
		rw.WriteHeader(opt.statusCode)
		return
	}

	b, err := opt.encoder.Encode(body)
	if err != nil {
		_ = r.tower.Wrap(err).Caller(tower.GetCaller(3)).Log(ctx)
		return
	}
	contentType := opt.encoder.ContentType()
	if contentType != "" {
		rw.Header().Set("Content-Type", contentType)
	}

	compressed, err := opt.compressor.Compress(b)
	if err != nil {
		_ = r.tower.Wrap(err).Caller(tower.GetCaller(3)).Level(tower.WarnLevel).Log(ctx)
		rw.WriteHeader(opt.statusCode)
		_, err = rw.Write(b)
		if err != nil {
			_ = r.tower.Wrap(err).Caller(tower.GetCaller(3)).Log(ctx)
		}
		return
	}

	contentEncoding := opt.compressor.ContentEncoding()
	if contentEncoding != "" {
		rw.Header().Set("Content-Encoding", contentEncoding)
	}

	rw.WriteHeader(opt.statusCode)
	_, err = rw.Write(compressed)
	if err != nil {
		_ = r.tower.Wrap(err).Caller(tower.GetCaller(3)).Log(ctx)
	}
}

// RespondStream writes the given stream to the http.ResponseWriter.
//
// If the stream implements tower.HTTPCodeHint, the status code will be set to the value returned by the tower.HTTPCodeHint.
//
// If the Compression supports StreamCompression, the stream will be compressed by said StreamCompression and
// written to the http.ResponseWriter.
//
// There's a special case if you pass http.NoBody as body, there will be no respond body related operations executed.
// StatusCode default value is set to http.StatusNoContent. You can still override this output by setting the
// related RespondOption.
// With http.NoBody as body, Towerhttp will immediately respond with status code after RespondOption are evaluated
// and end the process.
//
// Body of nil will be treated as http.NoBody.
func (r Responder) RespondStream(ctx context.Context, rw http.ResponseWriter, contentType string, body io.Reader, opts ...RespondOption) {
	var (
		statusCode = http.StatusOK
		err        error
	)
	if body == nil {
		body = http.NoBody
	}
	if body == http.NoBody {
		statusCode = http.StatusNoContent
	}
	opt := r.buildOption(statusCode)
	if ch, ok := body.(tower.HTTPCodeHint); ok {
		opt.statusCode = ch.HTTPCode()
	}
	for _, o := range opts {
		o.apply(opt)
	}
	if body == http.NoBody {
		rw.WriteHeader(opt.statusCode)
		return
	}

	if sc, ok := opt.compressor.(StreamCompression); ok {
		body = sc.StreamCompress(body)
		contentEncoding := sc.ContentEncoding()
		if contentEncoding != "" {
			rw.Header().Set("Content-Encoding", contentEncoding)
		}
	}
	rw.Header().Set("Content-Type", contentType)
	rw.WriteHeader(opt.statusCode)
	_, err = io.Copy(rw, body)
	if err != nil {
		_ = r.tower.Wrap(err).Caller(tower.GetCaller(3)).Log(ctx)
	}
}
