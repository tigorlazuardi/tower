package bucket

import (
	"context"
	"io"
)

type File struct {
	data     io.Reader
	filename string
	mimetype string
	pretext  string
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
func NewFile(data io.Reader, filename string, mimetype string) *File {
	return &File{data: data, filename: filename, mimetype: mimetype}
}

type UploadResult struct {
	// The URL of the uploaded file, if successful.
	URL string
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
