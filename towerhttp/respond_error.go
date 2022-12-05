package towerhttp

import (
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
func (r Responder) RespondError(rw http.ResponseWriter, request *http.Request, errPayload error, opts ...RespondOption) {
	var (
		ctx            = request.Context()
		encodedBody    []byte
		err            error
		statusCode     = tower.Query.GetHTTPCode(errPayload)
		compressedBody []byte
	)
	if errPayload == nil {
		errPayload = errInternalServerError
	}
	opt := r.buildOption(statusCode, opts...)
	if len(r.hooks) > 0 {
		defer func() {
			var requestBody ClonedBody = NoopCloneBody{}
			if b, ok := request.Body.(ClonedBody); ok {
				requestBody = b
			} else if c := clonedBodyFromContext(request.Context()); c != nil {
				requestBody = c
			}
			hookContext := &RespondErrorHookContext{
				baseHook: &baseHook{
					Context:        opt,
					Request:        request,
					RequestBody:    requestBody,
					ResponseStatus: opt.StatusCode,
					ResponseHeader: rw.Header(),
					Tower:          r.tower,
					Error:          err,
				},
				ResponseBody: RespondErrorBody{
					PreEncoded:     errPayload,
					PostEncoded:    encodedBody,
					PostCompressed: compressedBody,
				},
			}
			for _, hook := range r.hooks {
				hook.RespondErrorHookContext(hookContext)
			}
		}()
	}
	body := r.errorTransformer.ErrorBodyTransform(ctx, errPayload)
	if body == nil {
		rw.WriteHeader(opt.StatusCode)
		return
	}
	encodedBody, err = opt.Encoder.Encode(body)
	if err != nil {
		const errMsg = "ENCODING ERROR"
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Header().Set("Content-Type", "text/plain")
		_, err = io.WriteString(rw, errMsg)
		return
	}
	contentType := opt.Encoder.ContentType()
	if contentType != "" {
		rw.Header().Set("Content-Type", contentType)
	}
	compressedBody, ok, err := opt.Compressor.Compress(encodedBody)
	if err != nil {
		_ = r.tower.Wrap(err).Caller(opt.Caller).Level(tower.WarnLevel).Log(ctx)
		rw.Header().Set("Content-Length", strconv.Itoa(len(encodedBody)))
		rw.WriteHeader(opt.StatusCode)
		_, err = rw.Write(encodedBody)
		return
	}
	if ok {
		contentEncoding := opt.Compressor.ContentEncoding()
		rw.Header().Set("Content-Encoding", contentEncoding)
		rw.Header().Set("Content-Length", strconv.Itoa(len(compressedBody)))
		rw.WriteHeader(opt.StatusCode)
		_, err = rw.Write(compressedBody)
		return
	}
	rw.Header().Set("Content-Length", strconv.Itoa(len(encodedBody)))
	rw.WriteHeader(opt.StatusCode)
	_, err = rw.Write(encodedBody)
}
