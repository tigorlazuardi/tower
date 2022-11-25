package bucket

import (
	"context"
	"github.com/bwmarrin/snowflake"
	"io"
	"os"
)

var snowflakeNode *snowflake.Node

func init() {
	host, _ := os.Hostname()
	snowflakeNode = generateSnowflakeNodeFromString("tower-bucket-" + host)
}

type File struct {
	data     io.Reader
	filename string
	mimetype string
	pretext  string
	size     int
}

func (f File) Data() io.Reader {
	return f.data
}

func (f File) Filename() string {
	return f.filename
}

func (f File) Mimetype() string {
	return f.mimetype
}

func (f *File) Read(p []byte) (n int, err error) {
	return f.data.Read(p)
}

func (f File) Pretext() string {
	return f.pretext
}

func (f File) Size() int {
	return f.size
}

// SetPretext Sets the pretext of the file. Depending on the bucket implementation, this may or may not be used.
func (f *File) SetPretext(pretext string) {
	f.pretext = pretext
}

func (f *File) Close() error {
	if closer, ok := f.data.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// NewFile Creates a new file with the given data, filename, and mimetype.
func NewFile(data io.Reader, mimetype string, opts ...FileOption) *File {
	var size int
	if lh, ok := data.(LengthHint); ok {
		size = lh.Len()
	}
	f := &File{
		data:     data,
		filename: snowflakeNode.Generate().String(),
		mimetype: mimetype,
		size:     size,
	}
	for _, opt := range opts {
		opt.apply(f)
	}
	return f
}

type UploadResult struct {
	// The URL of the uploaded file, if successful.
	URL string
	// The file instance used to upload the file.
	// The body of this file may have already been garbage collected.
	// So do not consume this file content again and only use the remaining metadata.
	File *File
	// If Error is not nil, the upload is considered failed.
	Error error
}

type Bucket interface {
	// Upload File(s) to the bucket.
	// If File.data implements io.Closer, the close method will be called after upload is done.
	// Whether the Upload operation is successful or not.
	//
	// The number of result will be the same as the number of files uploaded.
	Upload(ctx context.Context, attachment []*File) []UploadResult
}

type FileOption interface {
	apply(*File)
}

type FileOptionFunc func(*File)

func (f FileOptionFunc) apply(file *File) {
	f(file)
}

func WithPretext(pretext string) FileOption {
	return FileOptionFunc(func(file *File) {
		file.pretext = pretext
	})
}

func WithFilesize(size int) FileOption {
	return FileOptionFunc(func(file *File) {
		file.size = size
	})
}

func WithFilename(filename string) FileOption {
	return FileOptionFunc(func(file *File) {
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
