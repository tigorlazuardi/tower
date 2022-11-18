package towerhttp

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"
)

var _ Compression = (*GzipCompressor)(nil)

type GzipCompressor struct {
	pool *sync.Pool
}

func (g GzipCompressor) ContentEncoding() string {
	return "gzip"
}

func (g GzipCompressor) Compress(b []byte) ([]byte, error) {
	if len(b) <= 1500 {
		return b, nil
	}
	buf := g.pool.Get().(*bytes.Buffer) //nolint
	buf.Reset()
	w, _ := gzip.NewWriterLevel(buf, gzip.BestCompression)
	defer w.Close()
	_, err := w.Write(b)
	return buf.Bytes(), err
}

func (g GzipCompressor) StreamCompress(origin io.Reader) (io.Reader, error) {
	pr, pw := io.Pipe()
	w, _ := gzip.NewWriterLevel(pw, gzip.BestCompression)
	go func() {
		_, err := io.Copy(w, origin)
		w.Close()
		_ = pw.CloseWithError(err)
	}()
	return pr, nil
}
