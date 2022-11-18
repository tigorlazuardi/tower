package towerhttp

import (
	"io"
)

type Compressor interface {
	// ContentEncoding returns the value of the Content-Encoding header. If empty, the content-encoding header will be
	// set by the http framework.
	ContentEncoding() string
	// Compress the given bytes.
	Compress([]byte) ([]byte, error)
	// StreamCompress compresses the given reader and returns a new reader that will give the compressed data.
	StreamCompress(origin io.Reader) (io.Reader, error)
}

var _ Compressor = (*NoopCompressor)(nil)

// NoopCompressor is a compressor that does nothing. Basically it's an Uncompressed operation.
type NoopCompressor struct{}

func (n NoopCompressor) StreamCompress(origin io.Reader) (io.Reader, error) { return origin, nil }
func (n NoopCompressor) ContentEncoding() string                            { return "" }
func (n NoopCompressor) Compress(bytes []byte) ([]byte, error)              { return bytes, nil }
