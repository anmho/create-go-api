package generator

import (
	"path/filepath"
)

// fileMapping represents a source template to output file mapping
type fileMapping struct {
	outputPath   string
	templatePath string
}

// fileGenerationRule defines when and what files to generate
type fileGenerationRule struct {
	files     []fileMapping
	condition func(*Generator) bool
}

// getFileGenerationRules returns all file generation rules based on project configuration
func (g *Generator) getFileGenerationRules() []fileGenerationRule {
	var rules []fileGenerationRule

		// Base files (always generated)
	rules = append(rules, fileGenerationRule{
		files: []fileMapping{
			{"go.mod", "templates/base/go.mod.tmpl"},
			{"README.md", "templates/base/README.md.tmpl"},
			{"Makefile", "templates/Makefile.tmpl"},
			{".gitignore", "static/.gitignore"},
			{".dockerignore", "static/.dockerignore"},
			{".env", "templates/.env.tmpl"},
			{".mockery.yaml", "templates/.mockery.yaml.tmpl"},
			// docker-compose.yml is generated in database-specific rules
			{"prometheus.yml", "templates/prometheus.yml.tmpl"},
			{"grafana/provisioning/datasources/prometheus.yml", "templates/grafana/provisioning/datasources/prometheus.yml.tmpl"},
		},
	})

	// Config files (always generated)
	rules = append(rules, fileGenerationRule{
		files: []fileMapping{
			{"internal/config/stage.go", "static/internal/config/stage.go"},
			{"internal/config/config.go", "static/internal/config/config.go"},
			{"internal/config/local.yaml", "static/internal/config/local.yaml"},
			{"internal/config/production.yaml", "static/internal/config/production.yaml"},
		},
	})

	// Posts domain files (always generated)
	rules = append(rules, fileGenerationRule{
		files: []fileMapping{
			{"internal/posts/post.go", "static/internal/posts/post.go"},
			{"internal/posts/errors.go", "static/internal/posts/errors.go"},
			{"internal/posts/table.go", "static/internal/posts/table.go"},
			{"internal/posts/service.go", "static/internal/posts/service.go"},
			{"internal/posts/service_test.go", "static/internal/posts/service_test.go"},
		},
	})

	// Scripts (always generated)
	rules = append(rules, fileGenerationRule{
		files: []fileMapping{
			{"scripts/check-deps.sh", "templates/scripts/check-deps.sh.tmpl"},
			{"scripts/generate.sh", "templates/scripts/generate.sh.tmpl"},
			{"scripts/migrate.sh", "static/scripts/migrate.sh"},
		},
	})

	// Deploy scripts (only if deploy option is selected)
	if g.config.Deploy {
		rules = append(rules, fileGenerationRule{
			files: []fileMapping{
				{"scripts/deploy.sh", "templates/scripts/deploy.sh.tmpl"},
				{"scripts/destroy.sh", "templates/scripts/destroy.sh.tmpl"},
			},
		})
	}

	// Database type-specific files
	switch g.config.Database.Type {
	case DatabaseTypePostgres:
		rules = append(rules, fileGenerationRule{
			files: []fileMapping{
				{"internal/database/postgres.go", "static/internal/database/postgres.go"},
				{"internal/posts/postgres_table.go", "static/internal/posts/postgres_table.go"},
				{"internal/posts/postgres_table_test.go", "static/internal/posts/postgres_table_test.go"},
				{".env.local", "static/.env.local.postgres"},
				{"docker-compose.yml", "static/docker-compose.yml.postgres"},
				{"schema.sql", "static/schema.sql"},
			},
		})
	case DatabaseTypeDynamoDB:
		rules = append(rules, fileGenerationRule{
			files: []fileMapping{
				{"internal/database/dynamodb.go", "static/internal/database/dynamodb.go"},
				{"internal/posts/dynamodb_table.go", "static/internal/posts/dynamodb_table.go"},
				{"internal/posts/dynamodb_table_test.go", "static/internal/posts/dynamodb_table_test.go"},
				{"internal/posts/dynamodb_converters.go", "static/internal/posts/dynamodb_converters.go"},
				{".env.local", "templates/.env.local.dynamodb.tmpl"},
				{"docker-compose.yml", "static/docker-compose.yml.dynamodb"},
			},
		})
	}

	// Framework type-specific files
	switch g.config.Framework {
	case FrameworkTypeChi:
		rules = append(rules, fileGenerationRule{
			files: []fileMapping{
				{"cmd/api/main.go", "templates/cmd/api/main_chi.go.tmpl"},
				{"internal/posts/routes.go", "static/internal/posts/routes.go"},
			},
		})
	case FrameworkTypeConnectRPC:
		rules = append(rules, fileGenerationRule{
			files: []fileMapping{
				{"cmd/api/main.go", "templates/cmd/api/main_connectrpc.go.tmpl"},
				{"internal/api/posts_handler.go", "static/internal/api/posts_handler_connectrpc.go"},
				{"internal/posts/converters.go", "templates/internal/posts/converters.go.tmpl"},
				{"internal/protos/posts/v1/posts.proto", "static/protos/posts/v1/posts.proto"},
				{"buf.yaml", "static/buf.yaml"},
				{"buf.gen.yaml", "templates/buf.gen.yaml.tmpl"},
			},
		})
	}

	// Deployment files
	if g.config.Deploy {
		rules = append(rules, fileGenerationRule{
			files: []fileMapping{
				{"fly.toml", "templates/deploy/fly.toml.tmpl"},
				{"Dockerfile", "static/Dockerfile"},
				{".github/workflows/deploy.yml", "templates/deploy/github/workflows/deploy.yml.tmpl"},
			},
			condition: func(g *Generator) bool {
				// Create .github/workflows directory if needed
				_ = g.fs.MkdirAll(filepath.Join(g.config.OutputDir, ".github", "workflows"), 0755)
				return true
			},
		})
	}

	return rules
}

// awsRegionToFlyRegion maps AWS regions to Fly.io regions
// This ensures DynamoDB tables are in the same region as the Fly.io deployment
func awsRegionToFlyRegion(awsRegion string) string {
	// Map AWS regions to Fly.io regions
	// Fly.io regions are typically 3-letter codes
	regionMap := map[string]string{
		// US East
		"us-east-1":      "iad", // Washington, DC
		"us-east-2":      "ord", // Chicago
		// US West
		"us-west-1":      "sea", // Seattle
		"us-west-2":      "dfw", // Dallas
		// Europe
		"eu-west-1":      "ams", // Amsterdam
		"eu-west-2":      "lhr", // London
		"eu-west-3":      "cdg", // Paris
		"eu-central-1":   "fra", // Frankfurt
		"eu-north-1":     "arn", // Stockholm
		// Asia Pacific
		"ap-southeast-1": "sin", // Singapore
		"ap-southeast-2": "syd", // Sydney
		"ap-northeast-1": "nrt", // Tokyo
		"ap-south-1":     "bom", // Mumbai
		// South America
		"sa-east-1":      "gru", // São Paulo
	}

	if flyRegion, ok := regionMap[awsRegion]; ok {
		return flyRegion
	}

	// Default to iad (Washington, DC) if region not found
	return "iad"
}

// getTemplateData returns the data structure for template execution
func (g *Generator) getTemplateData() map[string]interface{} {
	// Determine Fly.io region based on AWS region if DynamoDB is selected
	flyRegion := "iad" // Default
	if g.config.Database.Type == DatabaseTypeDynamoDB && g.config.Database.AWSRegion != "" {
		flyRegion = awsRegionToFlyRegion(g.config.Database.AWSRegion)
	}

	return map[string]interface{}{
		"ProjectName": g.config.ProjectName,
		"ModulePath":  g.config.ModulePath,
		"Database": map[string]interface{}{
			"Type":           string(g.config.Database.Type),
			"AWSAccessKeyID": g.config.Database.AWSAccessKeyID,
			"AWSSecretKey":   g.config.Database.AWSSecretKey,
			"AWSRegion":      g.config.Database.AWSRegion,
		},
		"Framework":    string(g.config.Framework),
		"HasPostgres":  g.config.Database.Type == DatabaseTypePostgres,
		"HasDynamoDB":  g.config.Database.Type == DatabaseTypeDynamoDB,
		"HasChi":       g.config.Framework == FrameworkTypeChi,
		"HasConnectRPC": g.config.Framework == FrameworkTypeConnectRPC,
		"HasGRPC":      g.config.Framework == FrameworkTypeConnectRPC,
		"Deploy":       g.config.Deploy,
		"FlyRegion":    flyRegion,
	}
}

// generateFiles generates multiple files from a list of mappings
func (g *Generator) generateFiles(files []fileMapping, data interface{}) error {
	for _, file := range files {
		// Check if this is a static file (no .tmpl extension) that should be copied without template processing
		if filepath.Ext(file.templatePath) != ".tmpl" {
			if err := g.copyFile(file.outputPath, file.templatePath); err != nil {
				return err
			}
		} else {
			// Regular template file - process with template engine
			if err := g.generateFile(file.outputPath, file.templatePath, data); err != nil {
				return err
			}
		}
	}
	return nil
}

