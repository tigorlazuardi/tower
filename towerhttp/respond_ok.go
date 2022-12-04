package towerhttp

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/tigorlazuardi/tower"
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
func (r Responder) Respond(rw http.ResponseWriter, request *http.Request, body any, opts ...RespondOption) {
	var (
		statusCode  = http.StatusOK
		err         error
		bodyBytes   []byte
		rejectDefer bool
		ctx         = request.Context()
	)

	if ch, ok := body.(tower.HTTPCodeHint); ok {
		statusCode = ch.HTTPCode()
	}

	opt := r.buildOption(statusCode, opts...)
	caller := tower.GetCaller(opt.callerDepth)
	defer func() {
		if !rejectDefer {
			capture, _ := rw.(*responseCapture)
			if capture == nil {
				// just in case the response writer is not the one we capture. e.g. wrapped in another response writer implementer.
				capture = responseCaptureFromContext(ctx)
			}
			if capture != nil {
				clonedBody := wrapBodyCloner(bytes.NewReader(bodyBytes), -1)
				capture.SetBody(clonedBody).SetCaller(caller).SetTower(r.tower).SetError(err).SetLevel(tower.ErrorLevel)
			}
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
		opts := append(opts,
			Option.Respond().StatusCode(http.StatusInternalServerError),
			Option.Respond().AddCallerSkip(1),
		)
		r.RespondError(rw, request, err, opts...)
		rejectDefer = true
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
