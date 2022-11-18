package towerhttp

import (
	"io"
)

type Compression interface {
	// ContentEncoding returns the value of the Content-Encoding header. If empty, the content-encoding header will be
	// set by the http framework.
	ContentEncoding() string
	// Compress the given bytes.
	Compress([]byte) ([]byte, error)
	// StreamCompress compresses the given reader and returns a new reader that will give the compressed data.
	StreamCompress(origin io.Reader) (io.Reader, error)
}

var _ Compression = (*NoCompression)(nil)

// NoCompression is a compressor that does nothing. Basically it's an Uncompressed operation.
type NoCompression struct{}

func (n NoCompression) StreamCompress(origin io.Reader) (io.Reader, error) { return origin, nil }
func (n NoCompression) ContentEncoding() string                            { return "" }
func (n NoCompression) Compress(bytes []byte) ([]byte, error)              { return bytes, nil }
