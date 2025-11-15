package config

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/Oudwins/zog"
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

//go:embed production.yaml local.yaml
var configFS embed.FS

var _ = env.Parse // Imported for secrets parsing when needed

type Config struct {
	Server  ServerConfig   `yaml:"server"`
	Auth    *AuthConfig    `yaml:"auth,omitempty"`
	Metrics *MetricsConfig `yaml:"metrics,omitempty"`
	PostHog *PostHogConfig `yaml:"posthog,omitempty"`
	Secrets SecretsConfig  `yaml:"-"`
}

type ServerConfig struct {
	Port  string `yaml:"port"`
	Stage Stage  `yaml:"stage"`
}

type AuthConfig struct {
	TokenExpiry string `yaml:"token_expiry"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type PostHogConfig struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
}

type SecretsConfig struct {
	// DynamoDB configuration (from environment variables)
	AWSRegion          string `env:"AWS_REGION"`
	TableName          string `env:"TABLE_NAME"`
	EndpointURL        string `env:"DYNAMODB_ENDPOINT_URL"` // Optional: for local DynamoDB (e.g., http://localhost:8000)
	AWSAccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`

	// Postgres configuration (from environment variables)
	DatabaseURL string `env:"DATABASE_URL"`

	// Auth secrets
	JWTSecret string `env:"JWT_SECRET"`

	// PostHog secrets
	PostHogAPIKey string `env:"POSTHOG_API_KEY"`
}

// Load reads configuration from stage-specific YAML file and secrets from environment variables
// All config files (local.yaml, production.yaml) are bundled in the Docker image
// The STAGE environment variable selects which config file to use at runtime
// YAML file is the source of truth - no overrides
// Secrets (AWS credentials, database URLs, JWT secrets) are loaded from environment variables only
// Defaults to "production" if STAGE is not set
func Load() (*Config, error) {
	cfg := &Config{}

	// First, check STAGE from environment to know which .env file to load
	// We need to load the .env file BEFORE parsing all environment variables
	stageStr := os.Getenv("STAGE")
	var stage Stage
	if stageStr == "" {
		slog.Info("STAGE environment variable not set, running in production mode")
		stage = StageProduction
	} else {
		var err error
		stage, err = ParseStage(stageStr)
		if err != nil {
			return nil, err
		}
	}

	// Load .env file based on stage BEFORE parsing environment variables
	// This ensures values from .env files are available when parsing
	switch stage {
	case StageLocal:
		// Load .env.local file from current working directory
		// STAGE=local means Configuration will godotenv load .env.local and use local.yaml
		if err := godotenv.Load(".env.local"); err != nil {
			return nil, fmt.Errorf("failed to load .env.local file for local stage: %w. The file should exist in the project root", err)
		}
	case StageProduction:
		// STAGE=production means it will not load any environment file and use production.yaml
		// Secrets are set via deployment platform environment variables only
		// Do not load any .env file for production
	}

	// Load config from embedded filesystem (all config files are bundled in binary)
	// Both local.yaml and production.yaml are embedded, STAGE selects which to use
	// This allows the application to run in any mode without filesystem access
	configFileName := fmt.Sprintf("%s.yaml", stage)
	data, err := configFS.ReadFile(configFileName)
	if err != nil {
		return nil, fmt.Errorf("config file %s not found in embedded filesystem for STAGE=%s", configFileName, stage)
	}

	// Parse YAML file - this is the source of truth for all non-secret configuration
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file for stage %s: %w", stage, err)
	}

	// Parse secrets from environment variables (already loaded from .env files above)
	// Note: AWS credentials are optional when using local DynamoDB (endpoint_url is set)
	if err := env.Parse(&cfg.Secrets); err != nil {
		return nil, fmt.Errorf("failed to parse secrets from environment variables: %w. Please ensure all required secrets are set (e.g., DATABASE_URL, JWT_SECRET). AWS credentials are optional for local DynamoDB", err)
	}

	return cfg, nil
}

// configSchema defines the declarative validation schema for Config using zog
var configSchema = zog.Struct(zog.Shape{
	"Server": zog.Struct(zog.Shape{
		"Port": zog.String().Min(1).Required(zog.Message("server.port is required")),
		// Stage is a custom type, validated in TestFunc below
	}).TestFunc(func(server any, ctx zog.Ctx) bool {
		s, ok := server.(ServerConfig)
		if !ok {
			return false
		}
		return s.Stage.IsValid()
	}, zog.Message("server.stage must be one of: local, production")),
	"Secrets": zog.Struct(zog.Shape{
		"AWSRegion":          zog.String(),
		"TableName":          zog.String(),
		"EndpointURL":        zog.String(),
		"AWSAccessKeyID":     zog.String(),
		"AWSSecretAccessKey": zog.String(),
		"DatabaseURL":        zog.String(),
		"JWTSecret":          zog.String(),
		"PostHogAPIKey":      zog.String(),
	}).TestFunc(func(secrets any, ctx zog.Ctx) bool {
		s, ok := secrets.(SecretsConfig)
		if !ok {
			return false
		}
		hasDynamoDB := s.AWSRegion != "" || s.TableName != ""
		hasPostgres := s.DatabaseURL != ""

	if !hasDynamoDB && !hasPostgres {
			return false
	}
	if hasDynamoDB && hasPostgres {
			return false
	}

	if hasDynamoDB {
			if s.AWSRegion == "" || s.TableName == "" {
				return false
		}
			if s.EndpointURL == "" {
				if s.AWSAccessKeyID == "" || s.AWSSecretAccessKey == "" {
					return false
				}
			}
		}

		return true
	}, zog.Message("database configuration is invalid: must set either DynamoDB (AWS_REGION, TABLE_NAME) or Postgres (DATABASE_URL), but not both")),
	"Auth": zog.Ptr(zog.Struct(zog.Shape{
		"TokenExpiry": zog.String(),
	})),
	"PostHog": zog.Ptr(zog.Struct(zog.Shape{
		"Enabled": zog.Bool(),
		"Host":    zog.String(),
	})),
}).TestFunc(func(cfg any, ctx zog.Ctx) bool {
	c, ok := cfg.(*Config)
	if !ok {
		return false
	}

	// Validate auth config if present
	if c.Auth != nil {
		if c.Secrets.JWTSecret == "" {
			return false
		}
	}

	// Validate PostHog config if present
	if c.PostHog != nil && c.PostHog.Enabled {
		if c.Secrets.PostHogAPIKey == "" {
			return false
		}
	}

	return true
}, zog.Message("JWT_SECRET is required when auth is enabled; POSTHOG_API_KEY is required when posthog is enabled"))

// Validate validates the loaded configuration using declarative zog schema
// Returns an error if required fields are missing or invalid
func (c *Config) Validate() error {
	issues := configSchema.Validate(c)
	if len(issues) > 0 {
		// Convert zog issues to error messages
		var messages []string
		for path, issueList := range issues {
			for _, issue := range issueList {
				msg := path
				if issue.Code != "" {
					msg += fmt.Sprintf(": %s", issue.Code)
				}
				if issue.Message != "" {
					msg += fmt.Sprintf(" - %s", issue.Message)
				}
				if issue.Value != nil {
					msg += fmt.Sprintf(" (value: %v)", issue.Value)
				}
				messages = append(messages, msg)
			}
		}
		return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
	}
	return nil
}

