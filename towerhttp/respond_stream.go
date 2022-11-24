package towerhttp

import (
	"context"
	"io"
	"net/http"

	"github.com/tigorlazuardi/tower"
)

// RespondStream writes the given stream to the http.ResponseWriter.
//
// If the stream implements tower.HTTPCodeHint, the status code will be set to the value returned by the tower.HTTPCodeHint.
//
// If the Compression supports StreamCompression, the stream will be compressed by said StreamCompression and
// written to the http.ResponseWriter.
//
// There's a special case if you pass http.NoBody as body, there will be no respond body related operations executed.
// StatusCode default value is set to http.StatusNoContent. You can still override this output by setting the
// related RespondOption.
// With http.NoBody as body, Towerhttp will immediately respond with status code after RespondOption are evaluated
// and end the process.
//
// Body of nil will be treated as http.NoBody.
func (r Responder) RespondStream(ctx context.Context, rw http.ResponseWriter, contentType string, body io.Reader, opts ...RespondOption) {
	var (
		statusCode = http.StatusOK
		err        error
	)
	if body == nil {
		body = http.NoBody
	}
	opt := r.buildOption(statusCode)
	if ch, ok := body.(tower.HTTPCodeHint); ok {
		opt.statusCode = ch.HTTPCode()
	}
	for _, o := range opts {
		o.apply(opt)
	}
	if body == http.NoBody {
		rw.WriteHeader(opt.statusCode)
		return
	}

	if sc, ok := opt.compressor.(StreamCompression); ok {
		body = sc.StreamCompress(body)
		contentEncoding := sc.ContentEncoding()
		if contentEncoding != "" {
			rw.Header().Set("Content-Encoding", contentEncoding)
		}
	}
	rw.Header().Set("Content-Type", contentType)
	rw.WriteHeader(opt.statusCode)
	_, err = io.Copy(rw, body)
	if err != nil {
		_ = r.tower.Wrap(err).Caller(tower.GetCaller(r.callerDepth)).Log(ctx)
	}
}
