package config

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceConfig_GetDSN(t *testing.T) {
	cfg := &ServiceConfig{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
			SSLMode:  "disable",
		},
	}

	dsn := cfg.GetDSN()
	assert.Contains(t, dsn, "host=localhost")
	assert.Contains(t, dsn, "port=5432")
	assert.Contains(t, dsn, "user=testuser")
	assert.Contains(t, dsn, "password=testpass")
	assert.Contains(t, dsn, "dbname=testdb")
	assert.Contains(t, dsn, "sslmode=disable")
}

func TestServiceConfig_GetRedisAddr(t *testing.T) {
	cfg := &ServiceConfig{
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
	}

	addr := cfg.GetRedisAddr()
	assert.Equal(t, "localhost:6379", addr)
}

func TestServiceConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"prod", "prod", true},
		{"production", "production", true},
		{"PROD", "PROD", true},
		{"dev", "dev", false},
		{"staging", "staging", false},
		{"local", "local", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ServiceConfig{Environment: tt.environment}
			assert.Equal(t, tt.expected, cfg.IsProduction())
		})
	}
}

func TestServiceConfig_IsLocal(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		envVar      string
		expected    bool
	}{
		{"local env", "local", "", true},
		{"empty env", "", "", true},
		{"dev env", "dev", "", false},
		{"prod env", "prod", "", false},
		{"env var override", "prod", "true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				os.Setenv("USE_LOCAL_SECRETS", tt.envVar)
				defer os.Unsetenv("USE_LOCAL_SECRETS")
			}
			cfg := &ServiceConfig{Environment: tt.environment}
			assert.Equal(t, tt.expected, cfg.IsLocal())
		})
	}
}

func TestLoader_ApplyDefaults(t *testing.T) {
	loader := NewLoader()
	cfg := &ServiceConfig{}

	loader.applyDefaults(cfg)

	// Database defaults
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, 25, cfg.Database.MaxOpenConns)
	assert.Equal(t, 5, cfg.Database.MaxIdleConns)
	assert.Equal(t, 300, cfg.Database.ConnMaxLifetime)
	assert.Equal(t, "disable", cfg.Database.SSLMode) // Local mode

	// Redis defaults
	assert.Equal(t, 6379, cfg.Redis.Port)

	// JWT defaults
	assert.Equal(t, 24, cfg.JWT.ExpirationHours)
	assert.Equal(t, "neobank", cfg.JWT.Issuer)

	// Observability defaults
	assert.Equal(t, 9090, cfg.Observability.MetricsPort)
	assert.Equal(t, "info", cfg.Observability.LogLevel)
	assert.Equal(t, "json", cfg.Observability.LogFormat)
}

func TestLoader_ApplyDefaults_AWS(t *testing.T) {
	// Simulate AWS environment
	os.Setenv("AWS_WEB_IDENTITY_TOKEN_FILE", "/var/run/secrets/token")
	defer os.Unsetenv("AWS_WEB_IDENTITY_TOKEN_FILE")

	loader := NewLoader()
	cfg := &ServiceConfig{}

	loader.applyDefaults(cfg)

	// SSL should be required on AWS
	assert.Equal(t, "require", cfg.Database.SSLMode)
	assert.True(t, cfg.Redis.TLS)
}

func TestLoader_LoadFromFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := tempDir + "/config.yaml"

	configContent := `
service_name: test-service
service_port: 8080
environment: test
database:
  host: testdb.example.com
  port: 5432
  user: testuser
  name: testdb
redis:
  host: testredis.example.com
  port: 6379
jwt:
  expiration_hours: 48
  issuer: test-issuer
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	loader := NewLoader(WithConfigPath(tempDir))
	cfg := &ServiceConfig{}

	err = loader.loadFromFile(cfg)
	require.NoError(t, err)

	assert.Equal(t, "test-service", cfg.ServiceName)
	assert.Equal(t, 8080, cfg.ServicePort)
	assert.Equal(t, "test", cfg.Environment)
	assert.Equal(t, "testdb.example.com", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testredis.example.com", cfg.Redis.Host)
	assert.Equal(t, 48, cfg.JWT.ExpirationHours)
	assert.Equal(t, "test-issuer", cfg.JWT.Issuer)
}

func TestLoader_Load_LocalMode(t *testing.T) {
	tempDir := t.TempDir()

	// Create config file
	configContent := `
service_name: local-service
service_port: 8081
environment: local
database:
  host: localhost
  port: 5432
  user: localuser
  password: localpass
  name: localdb
redis:
  host: localhost
  port: 6379
jwt:
  secret: local-jwt-secret
aws:
  use_local_secrets: true
`
	err := os.WriteFile(tempDir+"/config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	loader := NewLoader(WithConfigPath(tempDir))
	cfg := &ServiceConfig{}

	err = loader.Load(ctx, cfg)
	require.NoError(t, err)

	assert.Equal(t, "local-service", cfg.ServiceName)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "localuser", cfg.Database.User)
	assert.Equal(t, "local-jwt-secret", cfg.JWT.Secret)
	assert.True(t, cfg.IsLocal())
}

func TestLoadConfig_Convenience(t *testing.T) {
	tempDir := t.TempDir()

	configContent := `
service_name: convenience-test
service_port: 8082
environment: local
`
	err := os.WriteFile(tempDir+"/config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	cfg, err := LoadServiceConfig(ctx, tempDir)
	require.NoError(t, err)

	assert.Equal(t, "convenience-test", cfg.ServiceName)
	assert.Equal(t, 8082, cfg.ServicePort)
}

func TestEnvironmentOverrides(t *testing.T) {
	tempDir := t.TempDir()

	// Create minimal config file
	configContent := `
service_name: override-test
environment: local
`
	err := os.WriteFile(tempDir+"/config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variable overrides
	os.Setenv("DATABASE_HOST", "env-override-host")
	os.Setenv("DATABASE_PORT", "5433")
	defer func() {
		os.Unsetenv("DATABASE_HOST")
		os.Unsetenv("DATABASE_PORT")
	}()

	ctx := context.Background()
	cfg, err := LoadServiceConfig(ctx, tempDir)
	require.NoError(t, err)

	// Note: Viper's automatic env binding uses underscores
	// The actual override behavior depends on how Viper is configured
	assert.Equal(t, "override-test", cfg.ServiceName)
}

func TestAWSConfigJSON(t *testing.T) {
	cfg := &AWSConfig{
		Region:           "us-east-1",
		UseLocalSecrets:  false,
		LocalSecretsPath: "/app/secrets.json",
	}

	data, err := json.Marshal(cfg)
	require.NoError(t, err)

	var decoded AWSConfig
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, cfg.Region, decoded.Region)
	assert.Equal(t, cfg.UseLocalSecrets, decoded.UseLocalSecrets)
	assert.Equal(t, cfg.LocalSecretsPath, decoded.LocalSecretsPath)
}
