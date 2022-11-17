package towerhttp

import (
	"bytes"
	"compress/gzip"
	"sync"
)

type Compressor interface {
	ContentEncoding() string
	Compress([]byte) ([]byte, error)
}

type GzipCompressor struct {
	pool *sync.Pool
}

func (g GzipCompressor) ContentEncoding() string {
	return "gzip"
}

func (g GzipCompressor) Compress(b []byte) ([]byte, error) {
	buf := g.pool.Get().(*bytes.Buffer) //nolint
	buf.Reset()
	w, _ := gzip.NewWriterLevel(buf, gzip.BestCompression)
	defer w.Close()
	_, err := w.Write(b)
	return buf.Bytes(), err
}
