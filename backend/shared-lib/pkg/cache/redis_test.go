package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheKeyGenerators(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) string
		input    string
		expected string
	}{
		{"AccountCacheKey", AccountCacheKey, "acc-123", "account:acc-123"},
		{"BalanceCacheKey", BalanceCacheKey, "acc-456", "balance:acc-456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultCacheTTL(t *testing.T) {
	assert.Equal(t, 5*time.Minute, DefaultCacheTTL)
}

func TestKeyPrefixes(t *testing.T) {
	assert.Equal(t, "account:", KeyPrefixAccount)
	assert.Equal(t, "balance:", KeyPrefixBalance)
}

func TestConfig(t *testing.T) {
	cfg := Config{
		Addr:     "localhost:6379",
		Password: "secret",
		DB:       0,
	}

	assert.Equal(t, "localhost:6379", cfg.Addr)
	assert.Equal(t, "secret", cfg.Password)
	assert.Equal(t, 0, cfg.DB)
}

// RedisClientInterface for testing
type MockRedisClient struct {
	data map[string]string
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]string),
	}
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", nil // Key doesn't exist
}

func (m *MockRedisClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *MockRedisClient) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func TestMockRedisClient_SetAndGet(t *testing.T) {
	client := NewMockRedisClient()
	ctx := context.Background()

	// Set a value
	err := client.Set(ctx, "test-key", "test-value", time.Minute)
	assert.NoError(t, err)

	// Get the value
	val, err := client.Get(ctx, "test-key")
	assert.NoError(t, err)
	assert.Equal(t, "test-value", val)
}

func TestMockRedisClient_GetNonExistent(t *testing.T) {
	client := NewMockRedisClient()
	ctx := context.Background()

	// Get non-existent key
	val, err := client.Get(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Empty(t, val)
}

func TestMockRedisClient_Delete(t *testing.T) {
	client := NewMockRedisClient()
	ctx := context.Background()

	// Set and then delete
	client.Set(ctx, "to-delete", "value", time.Minute)
	err := client.Delete(ctx, "to-delete")
	assert.NoError(t, err)

	// Verify deleted
	val, _ := client.Get(ctx, "to-delete")
	assert.Empty(t, val)
}
