// Package config provides unified configuration loading with AWS/local fallback
package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	awspkg "github.com/femi-lawal/new_bank/backend/shared-lib/pkg/aws"
	"github.com/spf13/viper"
)

// Environment represents the deployment environment
type Environment string

const (
	EnvLocal      Environment = "local"
	EnvDev        Environment = "dev"
	EnvStaging    Environment = "staging"
	EnvProduction Environment = "prod"
)

// ServiceConfig holds common service configuration
type ServiceConfig struct {
	// Service identification
	ServiceName string `mapstructure:"service_name"`
	ServicePort int    `mapstructure:"service_port"`
	Environment string `mapstructure:"environment"`

	// Database configuration
	Database DatabaseConfig `mapstructure:"database"`

	// Redis configuration
	Redis RedisConfig `mapstructure:"redis"`

	// Kafka configuration
	Kafka KafkaConfig `mapstructure:"kafka"`

	// JWT configuration
	JWT JWTConfig `mapstructure:"jwt"`

	// Observability
	Observability ObservabilityConfig `mapstructure:"observability"`

	// AWS-specific configuration
	AWS AWSConfig `mapstructure:"aws"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	// AWS-specific
	SecretARN string `mapstructure:"secret_arn"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	TLS      bool   `mapstructure:"tls"`
	// AWS-specific
	SecretARN string `mapstructure:"secret_arn"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers  []string `mapstructure:"brokers"`
	GroupID  string   `mapstructure:"group_id"`
	TLS      bool     `mapstructure:"tls"`
	SASL     bool     `mapstructure:"sasl"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret          string `mapstructure:"secret"`
	ExpirationHours int    `mapstructure:"expiration_hours"`
	Issuer          string `mapstructure:"issuer"`
	// AWS-specific
	SecretARN string `mapstructure:"secret_arn"`
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	MetricsEnabled bool   `mapstructure:"metrics_enabled"`
	MetricsPort    int    `mapstructure:"metrics_port"`
	TracingEnabled bool   `mapstructure:"tracing_enabled"`
	OTLPEndpoint   string `mapstructure:"otlp_endpoint"`
	LogLevel       string `mapstructure:"log_level"`
	LogFormat      string `mapstructure:"log_format"`
}

// AWSConfig holds AWS-specific configuration
type AWSConfig struct {
	Region           string `mapstructure:"region"`
	UseLocalSecrets  bool   `mapstructure:"use_local_secrets"`
	LocalSecretsPath string `mapstructure:"local_secrets_path"`
}

// Loader handles configuration loading with environment awareness
type Loader struct {
	configPath      string
	secretsProvider *awspkg.SecretsProvider
	isAWS           bool
}

// LoaderOption is a functional option for the Loader
type LoaderOption func(*Loader)

// WithConfigPath sets the configuration file path
func WithConfigPath(path string) LoaderOption {
	return func(l *Loader) {
		l.configPath = path
	}
}

// NewLoader creates a new configuration loader
func NewLoader(opts ...LoaderOption) *Loader {
	l := &Loader{
		configPath: ".",
		isAWS:      awspkg.IsRunningOnAWS(),
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// Load loads the configuration, automatically detecting environment
func (l *Loader) Load(ctx context.Context, cfg *ServiceConfig) error {
	// First, load from config file
	if err := l.loadFromFile(cfg); err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	// If running on AWS, load secrets from Secrets Manager
	if l.isAWS && !cfg.AWS.UseLocalSecrets {
		if err := l.loadAWSSecrets(ctx, cfg); err != nil {
			log.Printf("Warning: failed to load AWS secrets, falling back to config values: %v", err)
		}
	}

	// Apply defaults
	l.applyDefaults(cfg)

	return nil
}

func (l *Loader) loadFromFile(cfg *ServiceConfig) error {
	viper.AddConfigPath(l.configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Allow environment variables to override
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: config file not found, using defaults/env vars: %v", err)
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return err
	}

	return nil
}

func (l *Loader) loadAWSSecrets(ctx context.Context, cfg *ServiceConfig) error {
	// Initialize secrets provider if not already done
	if l.secretsProvider == nil {
		provider, err := awspkg.NewSecretsProvider(ctx, awspkg.SecretsConfig{
			Region:    cfg.AWS.Region,
			UseLocal:  cfg.AWS.UseLocalSecrets,
			LocalPath: cfg.AWS.LocalSecretsPath,
		})
		if err != nil {
			return fmt.Errorf("failed to create secrets provider: %w", err)
		}
		l.secretsProvider = provider
	}

	// Load database credentials
	if cfg.Database.SecretARN != "" {
		creds, err := l.secretsProvider.GetDatabaseCredentials(ctx, cfg.Database.SecretARN)
		if err != nil {
			return fmt.Errorf("failed to load database credentials: %w", err)
		}
		cfg.Database.Host = creds.Host
		cfg.Database.Port = creds.Port
		cfg.Database.User = creds.Username
		cfg.Database.Password = creds.Password
		cfg.Database.Name = creds.Database
		if creds.SSLMode != "" {
			cfg.Database.SSLMode = creds.SSLMode
		}
	}

	// Load Redis credentials
	if cfg.Redis.SecretARN != "" {
		creds, err := l.secretsProvider.GetRedisCredentials(ctx, cfg.Redis.SecretARN)
		if err != nil {
			return fmt.Errorf("failed to load redis credentials: %w", err)
		}
		cfg.Redis.Host = creds.Host
		cfg.Redis.Port = creds.Port
		cfg.Redis.Password = creds.Password
		cfg.Redis.TLS = creds.TLS
	}

	// Load JWT secret
	if cfg.JWT.SecretARN != "" {
		secret, err := l.secretsProvider.GetSecret(ctx, cfg.JWT.SecretARN)
		if err != nil {
			return fmt.Errorf("failed to load JWT secret: %w", err)
		}
		cfg.JWT.Secret = secret
	}

	return nil
}

func (l *Loader) applyDefaults(cfg *ServiceConfig) {
	// Database defaults
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.SSLMode == "" {
		if l.isAWS {
			cfg.Database.SSLMode = "require"
		} else {
			cfg.Database.SSLMode = "disable"
		}
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 25
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 5
	}
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = 300 // 5 minutes
	}

	// Redis defaults
	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}
	if l.isAWS && !cfg.AWS.UseLocalSecrets {
		cfg.Redis.TLS = true // Always use TLS on AWS
	}

	// JWT defaults
	if cfg.JWT.ExpirationHours == 0 {
		cfg.JWT.ExpirationHours = 24
	}
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "neobank"
	}

	// Observability defaults
	if cfg.Observability.MetricsPort == 0 {
		cfg.Observability.MetricsPort = 9090
	}
	if cfg.Observability.LogLevel == "" {
		cfg.Observability.LogLevel = "info"
	}
	if cfg.Observability.LogFormat == "" {
		cfg.Observability.LogFormat = "json"
	}

	// AWS defaults
	if cfg.AWS.Region == "" {
		cfg.AWS.Region = awspkg.GetRegion()
	}
}

// GetDSN returns the database DSN
func (cfg *ServiceConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
}

// GetRedisAddr returns the Redis address
func (cfg *ServiceConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
}

// IsProduction returns true if running in production
func (cfg *ServiceConfig) IsProduction() bool {
	env := strings.ToLower(cfg.Environment)
	return env == "prod" || env == "production"
}

// IsLocal returns true if running locally
func (cfg *ServiceConfig) IsLocal() bool {
	env := strings.ToLower(cfg.Environment)
	return env == "local" || env == "" || os.Getenv("USE_LOCAL_SECRETS") == "true"
}

// LoadServiceConfig is a convenience function for loading AWS-integrated configuration
func LoadServiceConfig(ctx context.Context, path string) (*ServiceConfig, error) {
	cfg := &ServiceConfig{}
	loader := NewLoader(WithConfigPath(path))
	if err := loader.Load(ctx, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
