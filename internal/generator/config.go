package generator

// DatabaseType represents the database type
type DatabaseType string

const (
	DatabaseTypePostgres DatabaseType = "postgres"
	DatabaseTypeDynamoDB DatabaseType = "dynamodb"
)

// FrameworkType represents the API framework type
type FrameworkType string

const (
	FrameworkTypeChi        FrameworkType = "chi"
	FrameworkTypeConnectRPC FrameworkType = "connectrpc"
)

// ProjectConfig holds all project configuration
type ProjectConfig struct {
	ProjectName string
	ModulePath  string
	OutputDir   string
	Database    DatabaseConfig
	Framework   FrameworkType
	Deploy      bool
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Type            DatabaseType
	AWSAccessKeyID  string // For DynamoDB
	AWSSecretKey    string // For DynamoDB
	AWSRegion       string // For DynamoDB
}

