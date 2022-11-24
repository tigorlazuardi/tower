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
// value returned by the tower.HTTPCodeHint method. If the towerhttp.RO.StatusCode RespondOption is set, it will override
// the status regardless of the tower.HTTPCodeHint.
//
// if err is nil, it will be replaced with "Internal Server Error" message. It is done this way, because the library
// assumes that you mishandled the method and to prevent sending empty values, a generic Internal Server Error message
// will be sent instead. If you wish to send an empty response, use Respond with http.NoBody as body.
func (r Responder) RespondError(ctx context.Context, rw http.ResponseWriter, err error, opts ...RespondOption) {
	if err == nil {
		err = errInternalServerError
	}
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
		_ = r.tower().Wrap(errIO).Caller(tower.GetCaller(r.callerDepthh)).Log(ctx)
		return
	}
	contentType := opt.encoder.ContentType()
	if contentType != "" {
		rw.Header().Set("Content-Type", contentType)
	}
	compressed, errIO := opt.compressor.Compress(b)
	if errIO != nil {
		_ = r.tower().Wrap(errIO).Caller(tower.GetCaller(r.callerDepthh)).Log(ctx)
		return
	}
	contentEncoding := opt.compressor.ContentEncoding()
	if contentEncoding != "" {
		rw.Header().Set("Content-Encoding", contentEncoding)
	}

	rw.WriteHeader(opt.statusCode)
	_, errIO = rw.Write(compressed)
	if errIO != nil {
		_ = r.tower().Wrap(errIO).Caller(tower.GetCaller(r.callerDepthh)).Log(ctx)
	}
}
