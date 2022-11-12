package bucket

import (
	"context"
	"io"
)

type File struct {
	Data     io.ReadCloser
	Filename string
	Mimetype string
}

type LengthHint interface {
	Length() uint64
}

type UploadResult struct {
	// The URL of the uploaded file, if successful.
	URL string
	// If Error is not nil, the upload is considered failed.
	Error error
}

type Bucket interface {
	// Upload File(s) to the bucket. The File.Data will be closed after the upload is done.
	//
	// The number of result will be the same as the number of files uploaded.
	Upload(ctx context.Context, attachment []File) []UploadResult
}
