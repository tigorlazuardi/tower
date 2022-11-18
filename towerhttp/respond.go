package towerhttp

import (
	"context"
	"github.com/tigorlazuardi/tower"
	"net/http"
)

// Respond with the given body and options.
//
// HTTP status by default is http.StatusOK. If body implements tower.HTTPCodeHint, the status code will be set to the
// value returned by the tower.HTTPCodeHint method. If the towerhttp.RO.StatusCode RespondOption is set, it will override
// the status regardless of the tower.HTTPCodeHint.
func (t TowerHttp) Respond(ctx context.Context, rw http.ResponseWriter, body any, opts ...RespondOption) {
	var (
		statusCode = http.StatusOK
		err        error
	)
	defer func() {
		if err != nil {
			_ = t.tower.Wrap(err).Caller(tower.GetCaller(4)).Log(ctx)
		}
	}()

	if ch, ok := body.(tower.HTTPCodeHint); ok {
		statusCode = ch.HTTPCode()
	}

	opt := &option{
		encoder:    t.encoder,
		transfomer: t.transform,
		compressor: t.compressor,
		statusCode: statusCode,
	}

	for _, o := range opts {
		o.apply(opt)
	}

	body, err = opt.transfomer.BodyTransform(ctx, body)
	if err != nil {
		return
	}

	b, err := opt.encoder.Encode(body)
	if err != nil {
		return
	}
	contentType := opt.encoder.ContentType()
	rw.Header().Set("Content-Type", contentType)

	compressed, ok, err := opt.compressor.Compress(b)
	if err != nil {
		_ = t.tower.Wrap(err).Caller(tower.GetCaller(3)).Level(tower.WarnLevel).Log(ctx)
		rw.WriteHeader(opt.statusCode)
		_, err = rw.Write(b)
		return
	}

	if !ok {
		rw.WriteHeader(opt.statusCode)
		_, err = rw.Write(b)
		return
	}

	if len(compressed) < len(b) {
		rw.Header().Set("Content-Encoding", opt.compressor.ContentEncoding())
		rw.WriteHeader(opt.statusCode)
		_, err = rw.Write(compressed)
		return
	}

	rw.WriteHeader(opt.statusCode)
	_, err = rw.Write(compressed)
}
