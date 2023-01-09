package s3_test

import (
	"context"
	"github.com/tigorlazuardi/tower/bucket"
	"github.com/tigorlazuardi/tower/bucket/s3/v2"
	"strings"
)

func ExampleNewS3Bucket_integrated() {
	bkt, err := s3.NewS3Bucket("my-bucket.s3.us-east-1.amazonaws.com")
	if err != nil {
		return
	}
	f := strings.NewReader("hello world")
	file := bucket.NewFile(f, "text/plain; charset=utf-8")
	for _, result := range bkt.Upload(context.Background(), []bucket.File{file}) {
		if result.Error != nil {
			// handle error
			return
		}
	}
}
