package minio

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/tigorlazuardi/tower/bucket"
	"io"
	"os"
	"strings"
	"testing"
)

func createClient() (*minio.Client, func(), error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(err)
	}
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "quay.io/minio/minio",
		Tag:        "RELEASE.2023-01-02T09-40-09Z",
		Cmd:        []string{"server", "/data"},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return nil, nil, err
	}
	host := "172.17.0.1" // docker0 interface
	if os.Getenv("DOCKER_TEST_HOST") != "" {
		host = os.Getenv("DOCKER_TEST_HOST")
	}
	target := fmt.Sprintf("%s:%s", host, resource.GetPort("9000/tcp"))
	var client *minio.Client
	if err := pool.Retry(func() error {
		var err error
		client, err = minio.New(target, &minio.Options{
			Creds: credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		})
		if err != nil {
			fmt.Println(err.Error())
		}
		return err
	}); err != nil {
		_ = pool.Purge(resource)
		return nil, nil, fmt.Errorf("could not connect to docker target '%s': %w", target, err)
	}
	cleanup := func() {
		_ = pool.Purge(resource)
	}
	return client, cleanup, nil
}

func TestMinio_Upload(t *testing.T) {
	const test = "test"
	client, clean, err := createClient()
	if err != nil {
		t.Fatal(err)
	}
	defer clean()
	wc := Wrap(client, test,
		WithMakeBucketOption(minio.MakeBucketOptions{}),
		WithPutObjectOption(func(ctx context.Context, file bucket.File) minio.PutObjectOptions {
			return minio.PutObjectOptions{
				ContentType: file.ContentType(),
			}
		}),
		WithFilePrefixStringer(StringerFunc(func() string {
			return "prefix/"
		})))
	f := bucket.NewFile(strings.NewReader(test), "text/plain; charset=utf-8",
		bucket.WithFilename("test.txt"),
		bucket.WithFilesize(len(test)),
	)
	results := wc.Upload(context.Background(), []bucket.File{f})
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	for _, result := range results {
		func(result bucket.UploadResult) {
			if result.Error != nil {
				t.Fatalf("unexpected error: %v", result.Error)
			}
			obj, err := client.GetObject(context.Background(), "test", "prefix/"+result.File.Filename(), minio.GetObjectOptions{})
			if err != nil {
				t.Fatalf("could not get object: %v", err)
			}
			defer func(obj *minio.Object) {
				err := obj.Close()
				if err != nil {
					t.Fatalf("could not close object: %v", err)
				}
			}(obj)
			stats, err := obj.Stat()
			if err != nil {
				t.Fatalf("could not stat object: %v", err)
			}
			if stats.ContentType != result.File.ContentType() {
				t.Errorf("expected content type '%s' but got '%s'", result.File.ContentType(), stats.ContentType)
			}
			content, err := io.ReadAll(obj)
			if err != nil {
				t.Fatalf("could not read object: %v", err)
			}
			if string(content) != "test" {
				t.Fatalf("unexpected content: %s", content)
			}
			if !strings.Contains(result.URL, "test.txt") {
				t.Errorf("unexpected url: %s", result.URL)
			}
		}(result)
	}
	wc = Wrap(client, test, WithFilePrefix("test/"))
	f = bucket.NewFile(strings.NewReader(test), "text/plain; charset=utf-8",
		bucket.WithFilename("test.txt"),
		bucket.WithFilesize(len(test)),
	)
	results = wc.Upload(context.Background(), []bucket.File{f})
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	for _, result := range results {
		func(result bucket.UploadResult) {
			if result.Error != nil {
				t.Fatalf("unexpected error: %v", result.Error)
			}
			obj, err := client.GetObject(context.Background(), "test", "test/"+result.File.Filename(), minio.GetObjectOptions{})
			if err != nil {
				t.Fatalf("could not get object: %v", err)
			}
			defer func(obj *minio.Object) {
				err := obj.Close()
				if err != nil {
					t.Fatalf("could not close object: %v", err)
				}
			}(obj)
		}(result)
	}

	wc = Wrap(client, test)
	f = bucket.NewFile(strings.NewReader(test), "text/plain; charset=utf-8",
		bucket.WithFilename("test.txt"),
		bucket.WithFilesize(len(test)),
	)
	results = wc.Upload(context.Background(), []bucket.File{f})
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	for _, result := range results {
		func(result bucket.UploadResult) {
			if result.Error != nil {
				t.Fatalf("unexpected error: %v", result.Error)
			}
			obj, err := client.GetObject(context.Background(), "test", result.File.Filename(), minio.GetObjectOptions{})
			if err != nil {
				t.Fatalf("could not get object: %v", err)
			}
			defer func(obj *minio.Object) {
				err := obj.Close()
				if err != nil {
					t.Fatalf("could not close object: %v", err)
				}
			}(obj)
		}(result)
	}
}
