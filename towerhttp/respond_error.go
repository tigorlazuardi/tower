package towerhttp

import (
	"context"
	"net/http"

	"github.com/tigorlazuardi/tower"
)

type errString string

func (e errString) Error() string {
	return string(e)
}

const errInternalServerError errString = "Internal Server Error"

// RespondError writes the given error to the http.ResponseWriter.
//
// error is expected to be a serializable type.
//
// HTTP Status code by default is http.StatusInternalServerError. If error implements tower.HTTPCodeHint, the status code will be set to the
// value returned by the tower.HTTPCodeHint method. If the towerhttp.Option.StatusCode RespondOption is set, it will override
// the status regardless of the tower.HTTPCodeHint.
//
// if err is nil, it will be replaced with "Internal Server Error" message. It is done this way, because the library
// assumes that you mishandled the method and to prevent sending empty values, a generic Internal Server Error message
// will be sent instead. If you wish to send an empty response, use Respond with http.NoBody as body.
func (r Responder) RespondError(ctx context.Context, rw http.ResponseWriter, errPayload error, opts ...RespondOption) {
	var (
		bodyBytes  []byte
		err        error
		statusCode = http.StatusInternalServerError
	)
	if errPayload == nil {
		errPayload = errInternalServerError
	}
	if ch, ok := errPayload.(tower.HTTPCodeHint); ok {
		statusCode = ch.HTTPCode()
	}
	opt := r.buildOption(statusCode)
	for _, o := range opts {
		o.apply(opt)
	}
	defer func() {
		caller := tower.GetCaller(r.callerDepth)
		if logger := loggerFromContext(ctx); logger != nil {
			logger.log(&loggerContext{
				ctx:            ctx,
				responseHeader: rw.Header(),
				responseStatus: opt.statusCode,
				responseBody:   bodyBytes,
				caller:         caller,
				err:            err,
			})
		} else if err != nil {
			_ = r.tower.Wrap(err).Caller(caller).Log(ctx)
		}
	}()

	body := r.errorTransformer.ErrorBodyTransform(ctx, errPayload)
	bodyBytes, err = opt.encoder.Encode(body)
	if err != nil {
		return
	}
	contentType := opt.encoder.ContentType()
	if contentType != "" {
		rw.Header().Set("Content-Type", contentType)
	}
	compressed, _, err := opt.compressor.Compress(bodyBytes)
	if err != nil {
		return
	}
	contentEncoding := opt.compressor.ContentEncoding()
	if contentEncoding != "" {
		rw.Header().Set("Content-Encoding", contentEncoding)
	}

	rw.WriteHeader(opt.statusCode)
	_, err = rw.Write(compressed)
}
