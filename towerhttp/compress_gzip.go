package towerhttp

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"
)

var _ Compression = (*GzipCompression)(nil)

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

func (g GzipCompression) ContentEncoding() string {
	return "gzip"
}

func (g GzipCompression) Compress(b []byte) ([]byte, bool, error) {
	if len(b) <= 1500 {
		return b, false, nil
	}
	buf := g.pool.Get().(*bytes.Buffer) //nolint
	buf.Reset()
	w, _ := gzip.NewWriterLevel(buf, gzip.BestCompression)
	defer w.Close()
	_, err := w.Write(b)
	if err != nil {
		return b, false, err
	}
	if buf.Len() > len(b) {
		return b, false, err
	}
	c := make([]byte, buf.Len())
	copy(c, buf.Bytes())
	g.pool.Put(buf)
	return c, true, err
}

func (g GzipCompression) StreamCompress(origin io.Reader) (io.Reader, error) {
	pr, pw := io.Pipe()
	w, _ := gzip.NewWriterLevel(pw, gzip.BestCompression)
	go func() {
		_, err := io.Copy(w, origin)
		w.Close()
		_ = pw.CloseWithError(err)
	}()
	return pr, nil
}
