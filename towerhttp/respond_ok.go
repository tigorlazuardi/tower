package towerhttp

import (
	"context"
	"github.com/tigorlazuardi/tower"
	"net/http"
)

// Respond with the given body and options.
//
// body is expected to be a serializable type. For streams, use RespondStream.
//
// HTTP status by default is http.StatusOK. If body implements tower.HTTPCodeHint, the status code will be set to the
// value returned by the tower.HTTPCodeHint method. If the towerhttp.Option.StatusCode RespondOption is set, it will override
// the status regardless of the tower.HTTPCodeHint.
//
// There's a special case if you pass http.NoBody as body, there will be no respond body related operations executed.
// StatusCode default value is STILL http.StatusOK. If you wish to set the status code to http.StatusNoContent, you
// can still override this output by setting the related RespondOption.
//
// Body of nil has different treatment with http.NoBody. if body is nil, the nil value is still passed to the BodyTransformer implementer,
// therefore the final result body may not actually be empty.
func (r Responder) Respond(ctx context.Context, rw http.ResponseWriter, body any, opts ...RespondOption) {
	caller := tower.GetCaller(r.callerDepth)
	var (
		statusCode = http.StatusOK
		err        error
		bodyBytes  []byte
	)

	if ch, ok := body.(tower.HTTPCodeHint); ok {
		statusCode = ch.HTTPCode()
	}

	opt := r.buildOption(statusCode)
	for _, o := range opts {
		o.apply(opt)
	}
	defer func() {
		if logger := loggerFromContext(ctx); logger != nil {
			logger.log(&loggerContext{
				ctx:            ctx,
				responseHeader: rw.Header(),
				responseStatus: opt.statusCode,
				responseBody:   bodyBytes,
				caller:         caller,
				err:            err,
				tower:          r.tower,
			})
		} else if err != nil {
			_ = r.tower.Wrap(err).Caller(caller).Log(ctx)
		}
	}()

	if body == http.NoBody {
		rw.WriteHeader(opt.statusCode)
		return
	}

	body = opt.transformer.BodyTransform(ctx, body)
	if body == nil {
		rw.WriteHeader(opt.statusCode)
		return
	}

	bodyBytes, err = opt.encoder.Encode(body)
	if err != nil {
		opt.statusCode = http.StatusInternalServerError
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(opt.statusCode)
		bodyBytes = []byte("Encoding Error")
		_, _ = rw.Write(bodyBytes)
		return
	}
	contentType := opt.encoder.ContentType()
	if contentType != "" {
		rw.Header().Set("Content-Type", contentType)
	}

	compressed, err := opt.compressor.Compress(bodyBytes)
	if err != nil {
		_ = r.tower.Wrap(err).Caller(caller).Level(tower.WarnLevel).Log(ctx)
		rw.WriteHeader(opt.statusCode)
		_, err = rw.Write(bodyBytes)
		return
	}

	contentEncoding := opt.compressor.ContentEncoding()
	if contentEncoding != "" {
		rw.Header().Set("Content-Encoding", contentEncoding)
	}

	rw.WriteHeader(opt.statusCode)
	_, err = rw.Write(compressed)
}
