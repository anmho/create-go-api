package generator

import (
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static
//go:embed static/.gitignore
//go:embed static/.dockerignore
//go:embed static/.env
//go:embed static/.env.local.postgres
//go:embed static/.env.local.dynamodb
var staticFS embed.FS

// GetTemplatesFS returns the embedded templates filesystem
func GetTemplatesFS() embed.FS {
	return templatesFS
}

// GetStaticFS returns the embedded static filesystem
func GetStaticFS() embed.FS {
	return staticFS
}

// TemplateLoader interface for loading templates
type TemplateLoader interface {
	LoadTemplate(path string) (*template.Template, error)
}

// EmbeddedTemplateLoader loads templates from embedded filesystem
type EmbeddedTemplateLoader struct{}

func NewEmbeddedTemplateLoader() *EmbeddedTemplateLoader {
	return &EmbeddedTemplateLoader{}
}

func (l *EmbeddedTemplateLoader) LoadTemplate(path string) (*template.Template, error) {
	// Read template from embedded filesystem
	// Path already includes "templates/" prefix from project.go
	data, err := templatesFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", path, err)
	}

	// Parse template
	tmpl, err := template.New(path).Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", path, err)
	}

	return tmpl, nil
}
