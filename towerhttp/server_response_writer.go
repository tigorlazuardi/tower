package towerhttp

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
)

type responseCallbackFunc = func(status, size int, err error)

// newResponseCallback runs the callback after the response is finished written.
func newResponseCallback(ctx context.Context, w http.ResponseWriter, callback responseCallbackFunc) responseCallback {
	listener := &responseListener{
		w:        w,
		callback: callback,
	}
	listener.Listen(ctx)
	if cn, ok := w.(http.CloseNotifier); ok {
		return &responseListenerCN{
			responseListener: listener,
			CloseNotifier:    cn,
		}
	}
	return listener
}

type responseCallback interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	Status() int
	Size() int
}

var _ responseCallback = (*responseListener)(nil)

type responseListener struct {
	w          http.ResponseWriter
	status     int
	size       int
	callback   responseCallbackFunc
	writeError error
}

func (l *responseListener) Header() http.Header {
	return l.w.Header()
}

type responseListenerCN struct {
	*responseListener
	http.CloseNotifier
}

func (l *responseListener) Write(b []byte) (int, error) {
	size, err := l.w.Write(b)
	l.size += size
	if err != nil && err != io.EOF {
		l.writeError = err
	}
	return size, err
}

func (l *responseListener) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func (l *responseListener) Status() int {
	return l.status
}

func (l *responseListener) Size() int {
	return l.size
}

func (l *responseListener) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := l.w.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("tower-http: ResponseWriter does not implement http.Hijacker")
	}
	return h.Hijack()
}

func (l *responseListener) Flush() {
	f, ok := l.w.(http.Flusher)
	if !ok {
		return
	}
	f.Flush()
}

func (l *responseListener) Listen(ctx context.Context) {
	if ctx.Done() != nil && l.callback != nil {
		go func() {
			<-ctx.Done()
			l.callback(l.status, l.size, l.writeError)
		}()
	}
}
