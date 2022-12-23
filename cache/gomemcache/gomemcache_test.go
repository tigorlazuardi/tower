package gomemcache_test

import (
	"context"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"gomemcache"
	"os"
	"strings"
	"testing"
)

func createClient() (*memcache.Client, func(), error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(err)
	}
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "memcached",
		Tag:        "alpine",
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return nil, nil, err
	}
	host := "localhost"
	if os.Getenv("CI") != "" {
		host = resource.Container.Name
		host = strings.TrimPrefix(host, "/")
	}
	target := fmt.Sprintf("%s:%s", host, resource.GetPort("11211/tcp"))
	var client *memcache.Client
	if err := pool.Retry(func() error {
		client = memcache.New(target)
		return client.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		return nil, nil, fmt.Errorf("could not connect to docker target '%s': %w", target, err)
	}
	cleanup := func() {
		_ = pool.Purge(resource)
	}

	return client, cleanup, nil
}

func TestGoMemcache(t *testing.T) {
	client, cleanup, err := createClient()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	key := strings.Repeat("foo", 250)
	cache := gomemcache.Wrap(client)
	ctx := context.Background()
	err = cache.Set(ctx, key, []byte("bar"), 0)
	if err != nil {
		t.Fatal(err)
	}
	val, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if string(val) != "bar" {
		t.Fatalf("expected 'bar', got '%s'", string(val))
	}
	cache.Delete(ctx, key)
	if cache.Exist(ctx, key) {
		t.Fatal("expected key to be deleted")
	}
	if cache.Separator() != "::" {
		t.Fatalf("expected separator to be '::', got '%s'", cache.Separator())
	}

	baz := strings.Repeat("f", 1024*1024*2)
	err = cache.Set(ctx, key, []byte(baz), 0)
	if err == nil {
		t.Fatal("expected error when value is too large")
	}
	_, err = cache.Get(ctx, "not-exist")
	if err == nil {
		t.Fatal("expected error when key is not exist")
	}
}
