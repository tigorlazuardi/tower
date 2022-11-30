package towerhttp

import (
	"bytes"
	"io"
)

func newResponseBodyCapture(e Encoder, c Compression) (*responseBodyCapture, Encoder, Compression) {
	enc := &responseBodyCapture{e, nil}
	if sc, ok := c.(StreamCompression); ok {
		return enc, enc, &responseBOdyCaptureStream{enc, c, sc}
	}
	return enc, enc, c
}

type responseBodyCapture struct {
	enc  Encoder
	body io.Reader
}

func (r responseBodyCapture) ContentType() string {
	return r.enc.ContentType()
}

func (r responseBodyCapture) Encode(input any) ([]byte, error) {
	b, err := r.enc.Encode(input)
	if err != nil {
		r.body = &bytes.Buffer{}
		return b, err
	}
	r.body = bytes.NewReader(b)
	return b, nil
}

type responseBOdyCaptureStream struct {
	*responseBodyCapture
	Compression
	comp StreamCompression
}

func (r responseBOdyCaptureStream) ContentEncoding() string {
	return r.comp.ContentEncoding()
}

type teeCloser struct {
	io.Reader
	io.Closer
	clone *bytes.Buffer
}

func (t teeCloser) Read(p []byte) (n int, err error) {
	n, err = t.Reader.Read(p)
	if n > 0 {
		_, _ = t.clone.Write(p[:n])
	}
	return n, err
}

func newTeeCloser(r io.Reader) (*teeCloser, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	if rc, ok := r.(io.Closer); ok {
		return &teeCloser{r, rc, buf}, buf
	}
	return &teeCloser{r, io.NopCloser(r), buf}, buf
}

func (r responseBOdyCaptureStream) StreamCompress(origin io.Reader) io.Reader {
	reader, buf := newTeeCloser(origin)
	r.body = buf
	return r.comp.StreamCompress(reader)
}
