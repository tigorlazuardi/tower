package goredis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/tigorlazuardi/tower/cache"
)

func Wrap(client *redis.Client) cache.Cacher {
	return &goredis{client: client}
}

type goredis struct {
	client *redis.Client
}

// Set the Cache key and value.
func (goredis *goredis) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return goredis.client.Set(ctx, key, value, ttl).Err()
}

// Get the Value by Key. Returns tower.ErrNilCache if not found or ttl has passed.
func (goredis *goredis) Get(ctx context.Context, key string) ([]byte, error) {
	v, err := goredis.client.Get(ctx, key).Result()
	return []byte(v), err
}

// Delete cache by key.
func (goredis *goredis) Delete(ctx context.Context, key string) {
	goredis.client.Del(ctx, key)
}

// Exist Checks if Key exist in cache.
func (goredis *goredis) Exist(ctx context.Context, key string) bool {
	return goredis.client.Exists(ctx, key).Val() > 0
}

// Separator Returns Accepted separator value for the Cacher implementor.
func (goredis *goredis) Separator() string {
	return ":"
}
