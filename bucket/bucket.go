package bucket

import (
	"context"
	"github.com/bwmarrin/snowflake"
	"os"
)

var snowflakeNode *snowflake.Node

func init() {
	host, _ := os.Hostname()
	snowflakeNode = generateSnowflakeNodeFromString("tower-bucket-" + host)
}

type UploadResult struct {
	// The URL of the uploaded file, if successful.
	URL string
	// The file instance used to upload the file.
	// The body of this file may have already been garbage collected.
	// So do not consume this file content again and only use the remaining metadata.
	File File
	// If Error is not nil, the upload is considered failed.
	Error error
}

type Bucket interface {
	// Upload File(s) to the bucket.
	// If File.data implements io.Closer, the close method will be called after upload is done.
	// Whether the Upload operation is successful or not.
	//
	// The number of result will be the same as the number of files uploaded.
	Upload(ctx context.Context, files []File) []UploadResult
}

type FileOption interface {
	apply(*implFile)
}

type FileOptionFunc func(*implFile)

func (f FileOptionFunc) apply(file *implFile) {
	f(file)
}

func WithPretext(pretext string) FileOption {
	return FileOptionFunc(func(file *implFile) {
		file.pretext = pretext
	})
}

func WithFilesize(size int) FileOption {
	return FileOptionFunc(func(file *implFile) {
		file.size = size
	})
}

func WithFilename(filename string) FileOption {
	return FileOptionFunc(func(file *implFile) {
		file.filename = filename
	})
}

type LengthHint interface {
	Len() int
}

func generateSnowflakeNodeFromString(s string) *snowflake.Node {
	id := 0
	for _, c := range s {
		id += int(c)
	}
	for id > 1023 {
		id >>= 1
	}
	node, _ := snowflake.NewNode(int64(id))
	return node
}
