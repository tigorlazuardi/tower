package towerhttp

import (
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
		rejectDefer bool
		ctx         = request.Context()
	)
	var (
		encodedBody    []byte
		compressedBody []byte
	)

	if ch, ok := body.(tower.HTTPCodeHint); ok {
		statusCode = ch.HTTPCode()
	}

	opt := r.buildOption(statusCode, opts...)
	caller := tower.GetCaller(opt.CallerDepth)
	if len(r.hooks) > 0 {
		defer func() {
			if !rejectDefer {
				var requestBody ClonedBody = NoopCloneBody{}
				if b, ok := request.Body.(ClonedBody); ok {
					requestBody = b
				} else if c := clonedBodyFromContext(request.Context()); c != nil {
					requestBody = c
				}
				hookContext := &RespondHookContext{
					baseHook: &baseHook{
						Context:        opt,
						Request:        request,
						RequestBody:    requestBody,
						ResponseStatus: opt.StatusCode,
						ResponseHeader: rw.Header(),
						Tower:          r.tower,
						Error:          err,
					},
					ResponseBody: RespondBody{
						PreEncoded:     body,
						PostEncoded:    encodedBody,
						PostCompressed: compressedBody,
					},
				}
				for _, hook := range r.hooks {
					hook.RespondHook(hookContext)
				}
			}
		}()
	}

	if body == http.NoBody {
		rw.WriteHeader(opt.StatusCode)
		return
	}

	body = opt.BodyTransformer.BodyTransform(ctx, body)
	if body == nil {
		rw.WriteHeader(opt.StatusCode)
		return
	}

	encodedBody, err = opt.Encoder.Encode(body)
	if err != nil {
		opts := append(opts,
			Option.Respond().StatusCode(http.StatusInternalServerError),
			Option.Respond().AddCallerSkip(1),
		)
		r.RespondError(rw, request, err, opts...)
		rejectDefer = true
		return
	}
	contentType := opt.Encoder.ContentType()
	if contentType != "" {
		rw.Header().Set("Content-Type", contentType)
	}

	compressedBody, ok, err := opt.Compressor.Compress(encodedBody)
	if err != nil {
		_ = r.tower.Wrap(err).Caller(caller).Level(tower.WarnLevel).Log(ctx)
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
