package bucket

import "io"

type File interface {
	Data() io.Reader
	Filename() string
	ContentType() string
	Read(p []byte) (n int, err error)
	Pretext() string
	Size() int
	DataSource() io.Reader
	Close() error
}

type implFile struct {
	data     io.Reader
	filename string
	mimetype string
	pretext  string
	size     int
}

func (f implFile) Data() io.Reader {
	return f.data
}

func (f implFile) Filename() string {
	return f.filename
}

func (f implFile) ContentType() string {
	return f.mimetype
}

func (f *implFile) Read(p []byte) (n int, err error) {
	return f.data.Read(p)
}

func (f implFile) Pretext() string {
	return f.pretext
}

func (f implFile) Size() int {
	return f.size
}

func (f *implFile) DataSource() io.Reader {
	return f.data
}

// SetPretext Sets the pretext of the file. Depending on the bucket implementation, this may or may not be used.
func (f *implFile) SetPretext(pretext string) {
	f.pretext = pretext
}

func (f *implFile) Close() error {
	if closer, ok := f.data.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// NewFile is a built-in constructor for File implementor.
func NewFile(data io.Reader, mimetype string, opts ...FileOption) File {
	var size int
	if lh, ok := data.(LengthHint); ok {
		size = lh.Len()
	}
	f := &implFile{
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
