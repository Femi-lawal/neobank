// Package aws provides AWS-specific utilities with local fallback support
package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// SecretsProvider provides secrets from AWS Secrets Manager with local fallback
type SecretsProvider struct {
	client    *secretsmanager.Client
	useLocal  bool
	localPath string
	cache     map[string]*cachedSecret
}

type cachedSecret struct {
	value     string
	expiresAt time.Time
}

// SecretsConfig configures the secrets provider
type SecretsConfig struct {
	// UseLocal forces local file-based secrets (for development)
	UseLocal bool
	// LocalPath is the path to local secrets file (default: ./secrets.json)
	LocalPath string
	// Region is the AWS region
	Region string
	// CacheTTL is how long to cache secrets (default: 5 minutes)
	CacheTTL time.Duration
}

// NewSecretsProvider creates a new secrets provider
func NewSecretsProvider(ctx context.Context, cfg SecretsConfig) (*SecretsProvider, error) {
	provider := &SecretsProvider{
		useLocal:  cfg.UseLocal,
		localPath: cfg.LocalPath,
		cache:     make(map[string]*cachedSecret),
	}

	if cfg.LocalPath == "" {
		provider.localPath = "./secrets.json"
	}

	// Check if we should use local mode
	if cfg.UseLocal || os.Getenv("USE_LOCAL_SECRETS") == "true" {
		provider.useLocal = true
		return provider, nil
	}

	// Try to create AWS client
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		// Fall back to local if AWS config fails
		provider.useLocal = true
		return provider, nil
	}

	provider.client = secretsmanager.NewFromConfig(awsCfg)
	return provider, nil
}

// GetSecret retrieves a secret value
func (p *SecretsProvider) GetSecret(ctx context.Context, secretName string) (string, error) {
	// Check cache first
	if cached, ok := p.cache[secretName]; ok {
		if time.Now().Before(cached.expiresAt) {
			return cached.value, nil
		}
		delete(p.cache, secretName)
	}

	var value string
	var err error

	if p.useLocal {
		value, err = p.getLocalSecret(secretName)
	} else {
		value, err = p.getAWSSecret(ctx, secretName)
	}

	if err != nil {
		return "", err
	}

	// Cache the secret
	p.cache[secretName] = &cachedSecret{
		value:     value,
		expiresAt: time.Now().Add(5 * time.Minute),
	}

	return value, nil
}

// GetSecretJSON retrieves and unmarshals a JSON secret
func (p *SecretsProvider) GetSecretJSON(ctx context.Context, secretName string, target interface{}) error {
	value, err := p.GetSecret(ctx, secretName)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), target)
}

func (p *SecretsProvider) getAWSSecret(ctx context.Context, secretName string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := p.client.GetSecretValue(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	if result.SecretString != nil {
		return *result.SecretString, nil
	}

	return "", fmt.Errorf("secret %s has no string value", secretName)
}

func (p *SecretsProvider) getLocalSecret(secretName string) (string, error) {
	data, err := os.ReadFile(p.localPath)
	if err != nil {
		return "", fmt.Errorf("failed to read local secrets file: %w", err)
	}

	var secrets map[string]interface{}
	if err := json.Unmarshal(data, &secrets); err != nil {
		return "", fmt.Errorf("failed to parse local secrets: %w", err)
	}

	value, ok := secrets[secretName]
	if !ok {
		return "", fmt.Errorf("secret %s not found in local file", secretName)
	}

	// If value is a map/object, marshal it back to JSON
	switch v := value.(type) {
	case string:
		return v, nil
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal secret value: %w", err)
		}
		return string(jsonBytes), nil
	}
}

// IsLocal returns true if using local secrets
func (p *SecretsProvider) IsLocal() bool {
	return p.useLocal
}

// DatabaseCredentials represents database connection credentials
type DatabaseCredentials struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"sslmode"`
}

// GetDatabaseCredentials retrieves database credentials from the secret
func (p *SecretsProvider) GetDatabaseCredentials(ctx context.Context, secretName string) (*DatabaseCredentials, error) {
	var creds DatabaseCredentials
	if err := p.GetSecretJSON(ctx, secretName, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// BuildDSN builds a PostgreSQL DSN from credentials
func (c *DatabaseCredentials) BuildDSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "require"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, sslMode)
}

// RedisCredentials represents Redis connection credentials
type RedisCredentials struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	TLS      bool   `json:"tls"`
}

// GetRedisCredentials retrieves Redis credentials from the secret
func (p *SecretsProvider) GetRedisCredentials(ctx context.Context, secretName string) (*RedisCredentials, error) {
	var creds RedisCredentials
	if err := p.GetSecretJSON(ctx, secretName, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// BuildAddr builds a Redis address string
func (c *RedisCredentials) BuildAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
