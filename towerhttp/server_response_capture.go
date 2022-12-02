package towerhttp

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/tigorlazuardi/tower"
)

var responseCaptureKey struct{ key int } = struct{ key int }{777}

func contextWithResponseCapture(ctx context.Context, rc *responseCapture) context.Context {
	return context.WithValue(ctx, responseCaptureKey, rc)
}

func responseCaptureFromContext(ctx context.Context) *responseCapture {
	rc, _ := ctx.Value(responseCaptureKey).(*responseCapture)
	return rc
}

var (
	_ http.ResponseWriter = (*responseCapture)(nil)
	_ http.Hijacker       = (*responseCapture)(nil)
	_ http.Flusher        = (*responseCapture)(nil)
)

type responseCapture struct {
	r          *http.Request
	w          http.ResponseWriter
	status     int
	size       int
	writeError error
	body       ClonedBody
	logger     ServerLogger
	caller     tower.Caller
	tower      *tower.Tower
	level      tower.Level
}

func newResponseCapture(rw http.ResponseWriter, r *http.Request, logger ServerLogger) *responseCapture {
	return &responseCapture{
		w:      rw,
		status: http.StatusOK,
		body:   noopCloneBody{},
		logger: logger,
		r:      r,
		level:  tower.InfoLevel,
	}
}

func (r *responseCapture) SetTower(tower *tower.Tower) *responseCapture {
	r.tower = tower
	return r
}

func (r *responseCapture) SetBody(body ClonedBody) *responseCapture {
	r.body = body
	return r
}

func (r *responseCapture) SetCaller(caller tower.Caller) *responseCapture {
	r.caller = caller
	return r
}

func (r *responseCapture) SetError(err error) *responseCapture {
	r.writeError = err
	return r
}

func (r *responseCapture) SetBodyStream(body io.Reader, contentType string) io.Reader {
	n := r.logger.ReceiveResponseBodyStream(contentType, r.r)
	if n == 0 {
		return body
	}
	clone := wrapClientBodyCloner(body, n, nil)
	r.body = clone
	return clone
}

func (r *responseCapture) SetLevel(level tower.Level) *responseCapture {
	r.level = level
	return r
}

func (r *responseCapture) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.w.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("tower-http: ResponseWriter does not implement http.Hijacker")
	}
	return h.Hijack()
}

func (r *responseCapture) Flush() {
	f, ok := r.w.(http.Flusher)
	if !ok {
		return
	}
	f.Flush()
}

func (r *responseCapture) Status() int {
	return r.status
}

func (r *responseCapture) Size() int {
	return r.size
}

func (r *responseCapture) Body() ClonedBody {
	return r.body
}

func (r *responseCapture) Header() http.Header {
	return r.w.Header()
}

func (r *responseCapture) Write(bytes []byte) (int, error) {
	n, err := r.w.Write(bytes)
	if err != nil && err != io.EOF {
		r.writeError = err
	}
	r.size += n
	return n, err
}

func (r *responseCapture) WriteHeader(statusCode int) {
	r.w.WriteHeader(statusCode)
	r.status = statusCode
}
