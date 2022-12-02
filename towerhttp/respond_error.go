package towerhttp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/tigorlazuardi/tower"
)

type errString string

func (e errString) Error() string {
	return string(e)
}

const errInternalServerError errString = "Internal Server Error"

// RespondError writes the given error to the http.ResponseWriter.
//
// errPayload is expected to be a serializable type.
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
	caller := tower.GetCaller(opt.callerDepth)
	defer func() {
		if err == nil {
			err = errPayload
		}
		if capture, ok := rw.(*responseCapture); ok {
			body := bytes.NewBuffer(bodyBytes)
			capture.SetBody(&clientBodyCloner{
				ReadCloser: io.NopCloser(body),
				clone:      body,
				limit:      -1,
				callback:   nil,
			}).SetCaller(caller).SetTower(r.tower).SetError(err).SetLevel(tower.ErrorLevel)
		} else if capture := responseCaptureFromContext(ctx); capture != nil {
			// just in case the response writer is not the one we capture
			body := bytes.NewBuffer(bodyBytes)
			capture.SetBody(&clientBodyCloner{
				ReadCloser: io.NopCloser(body),
				clone:      body,
				limit:      -1,
				callback:   nil,
			}).SetCaller(caller).SetTower(r.tower).SetError(err).SetLevel(tower.ErrorLevel)
		}
	}()

	body := r.errorTransformer.ErrorBodyTransform(ctx, errPayload)
	bodyBytes, err = opt.encoder.Encode(body)
	if err != nil {
		const errMsg = "ENCODING ERROR"
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Header().Set("Content-Type", "text/plain")
		_, err = io.WriteString(rw, errMsg)
		return
	}
	contentType := opt.encoder.ContentType()
	if contentType != "" {
		rw.Header().Set("Content-Type", contentType)
	}
	compressed, ok, err := opt.compressor.Compress(bodyBytes)
	if err != nil {
		_ = r.tower.Wrap(err).Caller(caller).Level(tower.WarnLevel).Log(ctx)
		rw.Header().Set("Content-Length", strconv.Itoa(len(bodyBytes)))
		rw.WriteHeader(opt.statusCode)
		_, err = rw.Write(bodyBytes)
		return
	}
	if ok {
		contentEncoding := opt.compressor.ContentEncoding()
		rw.Header().Set("Content-Encoding", contentEncoding)
		rw.Header().Set("Content-Length", strconv.Itoa(len(compressed)))
		rw.WriteHeader(opt.statusCode)
		_, err = rw.Write(compressed)
		return
	}
	rw.Header().Set("Content-Length", strconv.Itoa(len(bodyBytes)))
	rw.WriteHeader(opt.statusCode)
	_, err = rw.Write(bodyBytes)
}
