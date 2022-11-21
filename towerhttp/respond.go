package towerhttp

import (
	"context"
	"github.com/tigorlazuardi/tower"
	"io"
	"net/http"
)

// Responder handles the response and writing to http.ResponseWriter.
type Responder struct {
	encoder    Encoder
	transform  BodyTransform
	tower      *tower.Tower
	compressor Compression
}

func (r Responder) buildOption(statusCode int) *option {
	opt := &option{
		encoder:    r.encoder,
		transfomer: r.transform,
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
func (r Responder) Respond(ctx context.Context, rw http.ResponseWriter, body any, opts ...RespondOption) {
	var (
		statusCode = http.StatusOK
		err        error
	)
	defer func() {
		if err != nil {
			_ = r.tower.Wrap(err).Caller(tower.GetCaller(3)).Log(ctx)
		}
	}()

	if ch, ok := body.(tower.HTTPCodeHint); ok {
		statusCode = ch.HTTPCode()
	}
	opt := r.buildOption(statusCode)
	for _, o := range opts {
		o.apply(opt)
	}

	body = opt.transfomer.BodyTransform(ctx, body)
	if body == nil {
		rw.WriteHeader(opt.statusCode)
		return
	}

	b, err := opt.encoder.Encode(body)
	if err != nil {
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
		return
	}

	contentEncoding := opt.compressor.ContentEncoding()
	if contentEncoding != "" {
		rw.Header().Set("Content-Encoding", contentEncoding)
	}

	rw.WriteHeader(opt.statusCode)
	_, err = rw.Write(compressed)
}

// RespondStream writes the given stream to the http.ResponseWriter.
//
// If the stream implements tower.HTTPCodeHint, the status code will be set to the value returned by the tower.HTTPCodeHint.
//
// If the Compression supports StreamCompression, the stream will be compressed by said StreamCompression and
// written to the http.ResponseWriter.
func (r Responder) RespondStream(ctx context.Context, rw http.ResponseWriter, contentType string, body io.Reader, opts ...RespondOption) {
	var (
		statusCode = http.StatusOK
		err        error
	)
	defer func() {
		if err != nil {
			_ = r.tower.Wrap(err).Caller(tower.GetCaller(3)).Log(ctx)
		}
	}()

	opt := r.buildOption(statusCode)
	if ch, ok := body.(tower.HTTPCodeHint); ok {
		opt.statusCode = ch.HTTPCode()
	}
	for _, o := range opts {
		o.apply(opt)
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
}
