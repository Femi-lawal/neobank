// Package aws provides AWS-specific utilities
package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// Config holds AWS configuration
type Config struct {
	Region   string
	IsLocal  bool
	Endpoint string // Custom endpoint for LocalStack testing
}

// NewConfig creates a new AWS configuration
func NewConfig(ctx context.Context, cfg Config) (aws.Config, error) {
	// Check if running locally
	if cfg.IsLocal || os.Getenv("AWS_ENVIRONMENT") == "local" {
		return aws.Config{}, nil
	}

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}

	// Support custom endpoint (for LocalStack)
	if cfg.Endpoint != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           cfg.Endpoint,
					SigningRegion: region,
				}, nil
			},
		)
		opts = append(opts, config.WithEndpointResolverWithOptions(customResolver))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return awsCfg, nil
}

// IsRunningOnAWS checks if the application is running on AWS infrastructure
func IsRunningOnAWS() bool {
	// Check for EKS-specific environment variables
	if os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE") != "" {
		return true
	}
	if os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI") != "" {
		return true
	}
	// Check for explicit environment setting
	if os.Getenv("AWS_ENVIRONMENT") == "aws" {
		return true
	}
	return false
}

// GetRegion returns the AWS region from environment or default
func GetRegion() string {
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region
	}
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		return region
	}
	return "us-east-1"
}
