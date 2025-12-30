package aws

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretsProvider_Local(t *testing.T) {
	// Create a temporary secrets file
	tempFile, err := os.CreateTemp("", "secrets-*.json")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	secrets := map[string]interface{}{
		"db-credentials": map[string]interface{}{
			"host":     "localhost",
			"port":     5432,
			"username": "testuser",
			"password": "testpass",
			"database": "testdb",
			"sslmode":  "disable",
		},
		"redis-credentials": map[string]interface{}{
			"host":     "localhost",
			"port":     6379,
			"password": "redispass",
			"tls":      false,
		},
		"simple-secret": "simple-value",
	}

	data, err := json.Marshal(secrets)
	require.NoError(t, err)
	_, err = tempFile.Write(data)
	require.NoError(t, err)
	tempFile.Close()

	ctx := context.Background()
	provider, err := NewSecretsProvider(ctx, SecretsConfig{
		UseLocal:  true,
		LocalPath: tempFile.Name(),
	})
	require.NoError(t, err)
	assert.True(t, provider.IsLocal())

	t.Run("GetSimpleSecret", func(t *testing.T) {
		value, err := provider.GetSecret(ctx, "simple-secret")
		require.NoError(t, err)
		assert.Equal(t, "simple-value", value)
	})

	t.Run("GetDatabaseCredentials", func(t *testing.T) {
		creds, err := provider.GetDatabaseCredentials(ctx, "db-credentials")
		require.NoError(t, err)
		assert.Equal(t, "localhost", creds.Host)
		assert.Equal(t, 5432, creds.Port)
		assert.Equal(t, "testuser", creds.Username)
		assert.Equal(t, "testpass", creds.Password)
		assert.Equal(t, "testdb", creds.Database)
		assert.Equal(t, "disable", creds.SSLMode)
	})

	t.Run("GetRedisCredentials", func(t *testing.T) {
		creds, err := provider.GetRedisCredentials(ctx, "redis-credentials")
		require.NoError(t, err)
		assert.Equal(t, "localhost", creds.Host)
		assert.Equal(t, 6379, creds.Port)
		assert.Equal(t, "redispass", creds.Password)
		assert.False(t, creds.TLS)
	})

	t.Run("SecretNotFound", func(t *testing.T) {
		_, err := provider.GetSecret(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("BuildDSN", func(t *testing.T) {
		creds, err := provider.GetDatabaseCredentials(ctx, "db-credentials")
		require.NoError(t, err)
		dsn := creds.BuildDSN()
		assert.Contains(t, dsn, "host=localhost")
		assert.Contains(t, dsn, "port=5432")
		assert.Contains(t, dsn, "user=testuser")
		assert.Contains(t, dsn, "password=testpass")
		assert.Contains(t, dsn, "dbname=testdb")
		assert.Contains(t, dsn, "sslmode=disable")
	})

	t.Run("BuildRedisAddr", func(t *testing.T) {
		creds, err := provider.GetRedisCredentials(ctx, "redis-credentials")
		require.NoError(t, err)
		addr := creds.BuildAddr()
		assert.Equal(t, "localhost:6379", addr)
	})

	t.Run("CachingWorks", func(t *testing.T) {
		// First call
		value1, err := provider.GetSecret(ctx, "simple-secret")
		require.NoError(t, err)

		// Modify the file (simulating change)
		newSecrets := map[string]interface{}{
			"simple-secret": "modified-value",
		}
		newData, _ := json.Marshal(newSecrets)
		os.WriteFile(tempFile.Name(), newData, 0644)

		// Second call should return cached value
		value2, err := provider.GetSecret(ctx, "simple-secret")
		require.NoError(t, err)
		assert.Equal(t, value1, value2)
	})
}

func TestIsRunningOnAWS(t *testing.T) {
	// Save original env vars
	origToken := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
	origCreds := os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")
	origEnv := os.Getenv("AWS_ENVIRONMENT")

	defer func() {
		os.Setenv("AWS_WEB_IDENTITY_TOKEN_FILE", origToken)
		os.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", origCreds)
		os.Setenv("AWS_ENVIRONMENT", origEnv)
	}()

	t.Run("NotOnAWS", func(t *testing.T) {
		os.Unsetenv("AWS_WEB_IDENTITY_TOKEN_FILE")
		os.Unsetenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")
		os.Unsetenv("AWS_ENVIRONMENT")
		assert.False(t, IsRunningOnAWS())
	})

	t.Run("WithWebIdentityToken", func(t *testing.T) {
		os.Setenv("AWS_WEB_IDENTITY_TOKEN_FILE", "/var/run/secrets/eks.amazonaws.com/token")
		assert.True(t, IsRunningOnAWS())
		os.Unsetenv("AWS_WEB_IDENTITY_TOKEN_FILE")
	})

	t.Run("WithContainerCredentials", func(t *testing.T) {
		os.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "/v2/credentials/xxx")
		assert.True(t, IsRunningOnAWS())
		os.Unsetenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI")
	})

	t.Run("WithExplicitEnvironment", func(t *testing.T) {
		os.Setenv("AWS_ENVIRONMENT", "aws")
		assert.True(t, IsRunningOnAWS())
		os.Unsetenv("AWS_ENVIRONMENT")
	})
}

func TestGetRegion(t *testing.T) {
	origRegion := os.Getenv("AWS_REGION")
	origDefault := os.Getenv("AWS_DEFAULT_REGION")

	defer func() {
		os.Setenv("AWS_REGION", origRegion)
		os.Setenv("AWS_DEFAULT_REGION", origDefault)
	}()

	t.Run("DefaultRegion", func(t *testing.T) {
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("AWS_DEFAULT_REGION")
		assert.Equal(t, "us-east-1", GetRegion())
	})

	t.Run("FromAWSRegion", func(t *testing.T) {
		os.Setenv("AWS_REGION", "eu-west-1")
		assert.Equal(t, "eu-west-1", GetRegion())
		os.Unsetenv("AWS_REGION")
	})

	t.Run("FromAWSDefaultRegion", func(t *testing.T) {
		os.Unsetenv("AWS_REGION")
		os.Setenv("AWS_DEFAULT_REGION", "ap-southeast-1")
		assert.Equal(t, "ap-southeast-1", GetRegion())
	})
}

func TestDatabaseCredentials_BuildDSN_DefaultSSL(t *testing.T) {
	creds := &DatabaseCredentials{
		Host:     "mydb.example.com",
		Port:     5432,
		Username: "admin",
		Password: "secret",
		Database: "production",
		SSLMode:  "", // Empty should default to "require"
	}

	dsn := creds.BuildDSN()
	assert.Contains(t, dsn, "sslmode=require")
}
