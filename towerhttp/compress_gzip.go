package towerhttp

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"
)

var (
	_ Compression       = (*GzipCompression)(nil)
	_ StreamCompression = (*GzipCompression)(nil)
)

type GzipCompression struct {
	pool *sync.Pool
}

// NewGzipCompression creates a new GzipCompression.
func NewGzipCompression() *GzipCompression {
	return &GzipCompression{
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

// ContentEncoding implements towerhttp.ContentEncodingHint.
func (g GzipCompression) ContentEncoding() string {
	return "gzip"
}

// Compress implements towerhttp.Compression.
func (g GzipCompression) Compress(b []byte) ([]byte, error) {
	buf := g.pool.Get().(*bytes.Buffer) //nolint
	buf.Reset()
	w, _ := gzip.NewWriterLevel(buf, gzip.BestCompression)
	defer w.Close()
	_, err := w.Write(b)
	if err != nil {
		return b, err
	}
	if buf.Len() > len(b) {
		return b, err
	}
	c := make([]byte, buf.Len())
	copy(c, buf.Bytes())
	g.pool.Put(buf)
	return c, err
}

// StreamCompress implements towerhttp.StreamCompression.
func (g GzipCompression) StreamCompress(origin io.Reader) io.Reader {
	pr, pw := io.Pipe()
	w, _ := gzip.NewWriterLevel(pw, gzip.BestCompression)
	go func() {
		_, err := io.Copy(w, origin)
		w.Close()
		_ = pw.CloseWithError(err)
	}()
	return pr
}
