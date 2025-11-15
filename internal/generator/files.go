package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	// File permissions using syscall constants
	// Regular file: rw-r--r-- (owner: read+write, group: read, other: read)
	filePermRegular = syscall.S_IRUSR | syscall.S_IWUSR | syscall.S_IRGRP | syscall.S_IROTH
	// Executable file: rwxr-xr-x (owner: read+write+execute, group: read+execute, other: read+execute)
	filePermExecutable = syscall.S_IRUSR | syscall.S_IWUSR | syscall.S_IXUSR | syscall.S_IRGRP | syscall.S_IXGRP | syscall.S_IROTH | syscall.S_IXOTH
)

// PlaceholderModulePath is a placeholder used in template files
// that will be replaced with the actual module path during generation
const PlaceholderModulePath = "github.com/acme/postservice"

// StaticModulePath is the module path used in static files for type checking
// This will be replaced with the actual module path during generation
const StaticModulePath = "github.com/andrewho/create-go-api/internal/generator/static"

// PlaceholderProjectName is a placeholder used in template files
// that will be replaced with the actual project name during generation
const PlaceholderProjectName = "postservice"

// FileSystem interface for file operations (for testing)
type FileSystem interface {
	MkdirAll(path string, perm os.FileMode) error
	WriteFile(name string, data []byte, perm os.FileMode) error
}

// OSFileSystem implements FileSystem using the OS
type OSFileSystem struct{}

func (f *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *OSFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// replaceModulePath replaces the placeholder module path with the actual module path
func replaceModulePath(content, modulePath string) string {
	// Replace the static module path used for type checking
	// Handle internal packages
	content = strings.ReplaceAll(content, StaticModulePath+"/internal", modulePath+"/internal")
	// Handle protos packages (for generated protobuf code)
	content = strings.ReplaceAll(content, StaticModulePath+"/internal/protos/gen", modulePath+"/internal/protos/gen")
	content = strings.ReplaceAll(content, StaticModulePath+"/protos", modulePath+"/protos")
	// Replace the placeholder module path (for templates)
	content = strings.ReplaceAll(content, PlaceholderModulePath, modulePath)
	return content
}

// replaceProjectName replaces the placeholder project name with the actual project name
func replaceProjectName(content, projectName string) string {
	content = strings.ReplaceAll(content, PlaceholderProjectName, projectName)
	return content
}

// copyFile copies a static file and replaces placeholders
func (g *Generator) copyFile(outputPath, sourcePath string) error {
	// Read from embedded filesystem
	staticFS := GetStaticFS()
	// sourcePath already includes "static/" prefix from project.go
	content, err := staticFS.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", sourcePath, err)
	}

	// Replace placeholders
	contentStr := string(content)
	contentStr = replaceModulePath(contentStr, g.config.ModulePath)
	contentStr = replaceProjectName(contentStr, g.config.ProjectName)
	
	// Remove build tags from generated files (they're only needed in templates directory)
	if strings.Contains(sourcePath, ".go") {
		// Remove //go:build ignore lines
		contentStr = strings.ReplaceAll(contentStr, "//go:build ignore\n", "")
		contentStr = strings.ReplaceAll(contentStr, "// +build ignore\n", "")
		// Remove extra newlines
		contentStr = strings.TrimPrefix(contentStr, "\n")
	}
	
	content = []byte(contentStr)

	// Create full output path
	outputFullPath := filepath.Join(g.config.OutputDir, outputPath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputFullPath)
	if err := g.fs.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Set executable permissions for shell scripts
	perm := os.FileMode(filePermRegular)
	if strings.HasSuffix(outputPath, ".sh") {
		perm = os.FileMode(filePermExecutable)
	}

	if err := g.fs.WriteFile(outputFullPath, content, perm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// generateFile generates a file from a template and replaces placeholders
func (g *Generator) generateFile(outputPath, templatePath string, data interface{}) error {
	tmpl, err := g.templateLoader.LoadTemplate(templatePath)
	if err != nil {
		return fmt.Errorf("failed to load template %s: %w", templatePath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	// Replace placeholders
	content := buf.Bytes()
	contentStr := string(content)
	contentStr = replaceModulePath(contentStr, g.config.ModulePath)
	contentStr = replaceProjectName(contentStr, g.config.ProjectName)
	content = []byte(contentStr)

	// Create full output path
	outputFullPath := filepath.Join(g.config.OutputDir, outputPath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputFullPath)
	if err := g.fs.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Set executable permissions for shell scripts
	perm := os.FileMode(filePermRegular)
	if strings.HasSuffix(outputPath, ".sh") {
		perm = os.FileMode(filePermExecutable)
	}

	if err := g.fs.WriteFile(outputFullPath, content, perm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

