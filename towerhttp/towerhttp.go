package towerhttp

import (
	"context"
	"github.com/tigorlazuardi/tower"
	"net/http"
)

type TowerHttp struct {
	encoder    Encoder
	transform  BodyTransform
	tower      *tower.Tower
	compressor Compressor
}

func (t TowerHttp) Respond(ctx context.Context, rw http.ResponseWriter, body any, opts ...RespondOption) {
	var (
		statusCode = http.StatusOK
		err        error
	)
	defer func() {
		if err != nil {
			_ = t.tower.Wrap(err).Caller(tower.GetCaller(3)).Log(ctx)
		}
	}()

	if ch, ok := body.(tower.HTTPCodeHint); ok {
		statusCode = ch.HTTPCode()
	}

	opt := &option{
		encoder:    t.encoder,
		transfomer: t.transform,
		compressor: t.compressor,
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

	compressed, err := opt.compressor.Compress(b)
	if err != nil {
		rw.WriteHeader(statusCode)
		_, err = rw.Write(b)
		return
	}

	if len(compressed) < len(b) {
		rw.Header().Set("Content-Encoding", opt.compressor.ContentEncoding())
		rw.WriteHeader(statusCode)
		_, err = rw.Write(compressed)
		return
	}

	rw.WriteHeader(statusCode)
	_, err = rw.Write(compressed)
}
