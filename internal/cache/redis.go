package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache wraps a Redis client with simple Get/Set/Delete operations
type Cache struct {
	client *redis.Client
}

// ErrCacheMiss is returned when a key doesn't exist in the cache
var ErrCacheMiss = errors.New("cache miss")

// NewCache creates a new Redis-backed cache
func NewCache(addr string) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// verify connection works
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Cache{client: client}, nil
}

// Get retrieves a value from the cache
// Returns ErrCacheMiss if the key doesn't exist
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

// Set stores a value in the cache with the given TTL
// Use 0 for no expiration
func (c *Cache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a key from the cache
// Returns no error if the key didn't exist
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Close shuts down the Redis client
func (c *Cache) Close() error {
	return c.client.Close()
}