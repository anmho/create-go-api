package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBOption func(*aws.Config)

// WithEndpoint sets a custom endpoint URL (optional, for local development/testing only)
// By default, uses AWS SDK default configuration which uses IAM roles when running on AWS infrastructure
func WithEndpoint(endpoint string) DynamoDBOption {
	return func(cfg *aws.Config) {
		if endpoint != "" {
			cfg.BaseEndpoint = aws.String(endpoint)
		}
	}
}

// WithRegion sets the AWS region
func WithRegion(region string) DynamoDBOption {
	return func(cfg *aws.Config) {
		cfg.Region = region
	}
}

// NewDynamoDB creates a new DynamoDB client
// Uses default AWS SDK configuration which will use IAM roles when running on AWS infrastructure
// (EC2, ECS, Lambda, etc.) or environment credentials
// When using a local endpoint (WithEndpoint), dummy credentials are used to allow local development without AWS credentials
func NewDynamoDB(ctx context.Context, opts ...DynamoDBOption) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Apply options
	for _, opt := range opts {
		opt(&cfg)
	}

	// Use dummy credentials if local endpoint is set
	if cfg.BaseEndpoint != nil && *cfg.BaseEndpoint != "" {
		cfg.Credentials = credentials.NewStaticCredentialsProvider("local", "local", "")
	}

	return dynamodb.NewFromConfig(cfg), nil
}

