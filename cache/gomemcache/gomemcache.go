package gomemcache

import (
	"context"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/tigorlazuardi/tower/cache"
	"time"
)

type cacher struct {
	client *memcache.Client
}

func (c cacher) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	if len(key) > 250 {
		key = key[:250]
	}
	item := &memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: int32(ttl.Seconds()),
	}
	err := c.client.Set(item)
	if err != nil {
		return fmt.Errorf("unable to set value to key '%s': %w", key, err)
	}
	return nil
}

func (c cacher) Get(_ context.Context, key string) ([]byte, error) {
	if len(key) > 250 {
		key = key[:250]
	}
	item, err := c.client.Get(key)
	if err != nil {
		return nil, fmt.Errorf("unable to get value from key '%s': %w", key, err)
	}
	return item.Value, nil
}

func (c cacher) Delete(_ context.Context, key string) {
	if len(key) > 250 {
		key = key[:250]
	}
	_ = c.client.Delete(key)
}

func (c cacher) Exist(_ context.Context, key string) bool {
	if len(key) > 250 {
		key = key[:250]
	}
	_, err := c.client.Get(key)
	return err == nil
}

func (c cacher) Separator() string {
	return "::"
}

func Wrap(client *memcache.Client) cache.Cacher {
	return cacher{client: client}
}
