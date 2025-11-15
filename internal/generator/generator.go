package generator

import (
	"fmt"
	"path/filepath"
)

type Generator struct {
	config         ProjectConfig
	fs             FileSystem
	templateLoader TemplateLoader
}

// NewGenerator creates a new generator with default dependencies
func NewGenerator(config ProjectConfig) *Generator {
	return &Generator{
		config:         config,
		fs:             &OSFileSystem{},
		templateLoader: NewEmbeddedTemplateLoader(),
	}
}

// Generate generates the complete project structure
func (g *Generator) Generate() error {
	// Create output directory
	if err := g.fs.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create directory structure
	if err := g.createDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Get all file generation rules based on project configuration
	rules := g.getFileGenerationRules()
	data := g.getTemplateData()

	// Generate all files based on rules
	for _, rule := range rules {
		if rule.condition == nil || rule.condition(g) {
			if err := g.generateFiles(rule.files, data); err != nil {
				return err
			}
		}
	}

	return nil
}

// createDirectoryStructure creates the necessary directory structure
func (g *Generator) createDirectoryStructure() error {
	dirs := []string{
		"cmd/api",
		"internal/config",
		"internal/database",
		"internal/posts",
		"internal/metrics",
	}

	// Add framework-specific directories
	if g.config.Framework == FrameworkTypeConnectRPC {
		dirs = append(dirs, "internal/protos/posts/v1", "internal/protos/gen/posts/v1")
	}

	// Add migrations directory if using PostgreSQL
	if g.config.Database.Type == DatabaseTypePostgres {
		dirs = append(dirs, "migrations")
	}

	// Add scripts directory
	dirs = append(dirs, "scripts")

	// Add terraform directory if using DynamoDB
	if g.config.Database.Type == DatabaseTypeDynamoDB {
		dirs = append(dirs, "terraform")
	}

	for _, dir := range dirs {
		path := filepath.Join(g.config.OutputDir, dir)
		if err := g.fs.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

