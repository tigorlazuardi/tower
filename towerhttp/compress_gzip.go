package towerhttp

import (
	"bytes"
	"compress/gzip"
	"github.com/tigorlazuardi/tower/internal/pool"
	"io"
)

var (
	_ Compressor        = (*GzipCompression)(nil)
	_ StreamCompression = (*GzipCompression)(nil)
)

type GzipCompression struct {
	pool  *pool.Pool[*bytes.Buffer]
	level int
}

// NewGzipCompression creates a new GzipCompression.
func NewGzipCompression() *GzipCompression {
	return NewGzipCompressionWithLevel(gzip.DefaultCompression)
}

// NewGzipCompressionWithLevel creates a new GzipCompression with specified compression level.
func NewGzipCompressionWithLevel(lvl int) *GzipCompression {
	return &GzipCompression{
		pool: pool.New(func() *bytes.Buffer {
			return &bytes.Buffer{}
		}),
		level: lvl,
	}
}

// ContentEncoding implements towerhttp.ContentEncodingHint.
func (g GzipCompression) ContentEncoding() string {
	return "gzip"
}

// Compress implements towerhttp.Compressor.
func (g GzipCompression) Compress(b []byte) ([]byte, bool, error) {
	buf := g.pool.Get()
	buf.Reset()
	w, err := gzip.NewWriterLevel(buf, g.level)
	if err != nil {
		return b, false, err
	}
	_, err = w.Write(b)
	if err != nil {
		return b, false, err
	}
	w.Close()
	c := make([]byte, buf.Len())
	// bytes.Buffer bytes method points to an array that will be reused by the pool.
	// So we need to copy the bytes to a new array.
	copy(c, buf.Bytes())
	g.pool.Put(buf)
	return c, true, err
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
