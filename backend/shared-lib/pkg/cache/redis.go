package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis client
type RedisClient struct {
	client *redis.Client
}

// Config holds Redis configuration
type Config struct {
	Addr     string
	Password string
	DB       int
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg Config) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	slog.Info("Successfully connected to Redis", "addr", cfg.Addr)
	return &RedisClient{client: client}, nil
}

// Get retrieves a value by key
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key doesn't exist
	}
	if err != nil {
		return "", err
	}
	slog.Debug("Cache hit", "key", key)
	return val, nil
}

// Set stores a value with TTL
func (r *RedisClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return err
	}
	slog.Debug("Cache set", "key", key, "ttl", ttl)
	return nil
}

// SetJSON stores a JSON-serialized value
func (r *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Set(ctx, key, string(data), ttl)
}

// GetJSON retrieves and unmarshals a JSON value
func (r *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	if val == "" {
		return redis.Nil // Not found
	}
	return json.Unmarshal([]byte(val), dest)
}

// Delete removes a key
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	slog.Debug("Cache deleted", "key", key)
	return nil
}

// DeletePattern removes all keys matching a pattern
func (r *RedisClient) DeletePattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Cache key prefixes
const (
	KeyPrefixAccount = "account:"
	KeyPrefixBalance = "balance:"
)

// AccountCacheKey returns the cache key for an account
func AccountCacheKey(accountID string) string {
	return KeyPrefixAccount + accountID
}

// BalanceCacheKey returns the cache key for an account balance
func BalanceCacheKey(accountID string) string {
	return KeyPrefixBalance + accountID
}

// Default TTL for cached items
const DefaultCacheTTL = 5 * time.Minute
