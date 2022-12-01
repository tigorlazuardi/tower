package towerhttp

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
)

var responseCaptureKey struct{ key int } = struct{ key int }{777}

type responseCapturer interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	Status() int
	// Size returns the size of the response body post compression.
	Size() int
	// Body returns the response body before compression.
	Body() ClonedBody
}

type responseCapture struct {
	w          http.ResponseWriter
	status     int
	size       int
	writeError error
	body       ClonedBody
}

type responseCaptureCN struct {
	*responseCapture
	http.CloseNotifier
}

func newResponseCapture(rw http.ResponseWriter) responseCapturer {
	rc := &responseCapture{
		w:      rw,
		status: http.StatusOK,
		body:   noopCloneBody{},
	}
	if cn, ok := rw.(http.CloseNotifier); ok {
		return &responseCaptureCN{rc, cn}
	}
	return rc
}

func (r responseCapture) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.w.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("tower-http: ResponseWriter does not implement http.Hijacker")
	}
	return h.Hijack()
}

func (r responseCapture) Flush() {
	f, ok := r.w.(http.Flusher)
	if !ok {
		return
	}
	f.Flush()
}

func (r responseCapture) Status() int {
	return r.status
}

func (r responseCapture) Size() int {
	return r.size
}

func (r responseCapture) Body() ClonedBody {
	return r.body
}

func (r responseCapture) Header() http.Header {
	return r.w.Header()
}

func (r responseCapture) Write(bytes []byte) (int, error) {
	n, err := r.w.Write(bytes)
	if err != nil && err != io.EOF {
		r.writeError = err
	}
	r.size += n
	return n, err
}

func (r responseCapture) WriteHeader(statusCode int) {
	r.w.WriteHeader(statusCode)
	r.status = statusCode
}
