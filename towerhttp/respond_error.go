package towerhttp

import (
	"context"
	"github.com/tigorlazuardi/tower"
	"net/http"
)

func (r Responder) RespondError(ctx context.Context, rw http.ResponseWriter, err error, opts ...RespondOption) {
	var (
		statusCode = http.StatusInternalServerError
		errIO      error
	)
	if ch, ok := err.(tower.HTTPCodeHint); ok {
		statusCode = ch.HTTPCode()
	}
	opt := r.buildOption(statusCode)
	for _, o := range opts {
		o.apply(opt)
	}

	body := r.errorTransformer.ErrorBodyTransform(ctx, err)
	b, errIO := opt.encoder.Encode(body)
	if errIO != nil {
		_ = r.tower.Wrap(errIO).Caller(tower.GetCaller(3)).Log(ctx)
		return
	}
	contentType := opt.encoder.ContentType()
	if contentType != "" {
		rw.Header().Set("Content-Type", contentType)
	}
	compressed, errIO := opt.compressor.Compress(b)
	if errIO != nil {
		_ = r.tower.Wrap(errIO).Caller(tower.GetCaller(3)).Log(ctx)
		return
	}
	contentEncoding := opt.compressor.ContentEncoding()
	if contentEncoding != "" {
		rw.Header().Set("Content-Encoding", contentEncoding)
	}

	rw.WriteHeader(opt.statusCode)
	_, errIO = rw.Write(compressed)
	if errIO != nil {
		_ = r.tower.Wrap(errIO).Caller(tower.GetCaller(3)).Log(ctx)
	}
}
