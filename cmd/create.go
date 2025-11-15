package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/andrewho/create-go-api/cmd/flags"
	"github.com/andrewho/create-go-api/internal/generator"
	"github.com/andrewho/create-go-api/internal/tui"
	"github.com/spf13/cobra"
)

var (
	projectName string
	modulePath  string
	outputDir   string
	driver      string
	framework   string
	deploy      bool
	interactive bool
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Go API service",
	Long: `Create a new Go API service with the specified configuration.

The command supports two modes:
  - Interactive TUI mode: Run without flags or use --interactive flag
  - Non-interactive CLI mode: Provide all required flags (--name, --driver, --framework, etc.)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If interactive flag is set, use TUI
		if interactive {
			app := tui.NewApp()
			return app.Run()
		}

		// Check if any flags were provided
		flagsProvided := projectName != "" || modulePath != "" || outputDir != "" ||
			driver != "" || framework != ""

		// If flags provided, use direct CLI mode
		if flagsProvided {
			if err := validateFlags(); err != nil {
				return err
			}

			cfg := generator.ProjectConfig{
				ProjectName: projectName,
				ModulePath:  modulePath,
				OutputDir:   outputDir,
				Database:    generator.DatabaseConfig{Type: generator.DatabaseType(driver)},
				Framework:   generator.FrameworkType(framework),
				Deploy:      deploy,
			}

			gen := generator.NewGenerator(cfg)
			if err := gen.Generate(); err != nil {
				return fmt.Errorf("failed to generate project: %w", err)
			}

			fmt.Printf("✓ Project generated successfully at: %s\n", outputDir)
			fmt.Printf("  Module:  %s\n", modulePath)
			fmt.Printf("  Database: %s\n", driver)
			fmt.Printf("  Framework: %s\n", framework)
			return nil
		}

		// Otherwise, use TUI
		app := tui.NewApp()
		return app.Run()
	},
}

func init() {
	createCmd.Flags().StringVarP(&projectName, "name", "n", "", "Project name")
	createCmd.Flags().StringVarP(&modulePath, "module-path", "m", "", "Go module path")
	createCmd.Flags().StringVarP(&driver, "driver", "d", "", "Database driver (postgres, dynamodb)")
	createCmd.Flags().StringVarP(&framework, "framework", "f", "", "API framework (chi, connectrpc)")
	createCmd.Flags().BoolVar(&deploy, "deploy", false, "Enable deployment setup")
	createCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (defaults to project name)")
	createCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Use interactive TUI mode (default when no flags provided)")
}

func validateFlags() error {
	if projectName == "" {
		return fmt.Errorf("project name is required")
	}

	if !flags.IsValidDatabase(driver) {
		return fmt.Errorf("invalid database driver: %s (must be one of: %s)", driver, strings.Join(flags.AllowedDatabases, ", "))
	}

	if !flags.IsValidFramework(framework) {
		return fmt.Errorf("invalid framework: %s (must be one of: %s)", framework, strings.Join(flags.AllowedFrameworks, ", "))
	}

	if outputDir == "" {
		outputDir = projectName
	}

	// Check if directory exists and is not empty
	if info, err := os.Stat(outputDir); err == nil {
		if info.IsDir() {
			entries, err := os.ReadDir(outputDir)
			if err == nil && len(entries) > 0 {
				return fmt.Errorf("directory %s already exists and is not empty", outputDir)
			}
		}
	}

	return nil
}

