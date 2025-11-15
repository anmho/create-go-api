package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/anmho/create-go-api/internal/generator"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type App struct {
	model *Model
}

func NewApp() *App {
	return &App{
		model: NewModel(),
	}
}

func (a *App) Run() error {
	p := tea.NewProgram(a.model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
	return nil
}

type Model struct {
	step            Step
	projectName     textInputModel
	modulePath      textInputModel
	outputDir       textInputModel
	databaseSelect  singleSelectModel
	awsProfileSelect singleSelectModel
	awsAccessKeyID  textInputModel
	awsSecretKey    textInputModel
	awsRegion       textInputModel
	awsProfileName  string
	frameworkSelect singleSelectModel
	deployConfirm   confirmModel
	spinner       spinner.Model
	err           error
	generating    bool
	deploying     bool
	deployEnabled bool
}

type Step int

const (
	StepWelcome Step = iota
	StepProjectName
	StepModulePath
	StepOutputDir
	StepDatabaseSelection
	StepAWSProfileSelection
	StepAWSAccessKeyID
	StepAWSSecretKey
	StepAWSRegion
	StepFrameworkSelection
	StepDeploySelection
	StepReview
	StepGenerating
	StepComplete
)

func NewModel() *Model {
	databaseOptions := []list.Item{
		listItem{title: "DynamoDB", description: "NoSQL database on AWS"},
		listItem{title: "PostgreSQL", description: "Relational database with Atlas migrations"},
	}

	frameworkOptions := []list.Item{
		listItem{title: "ConnectRPC", description: "gRPC-compatible framework"},
		listItem{title: "Chi", description: "Lightweight HTTP router"},
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	// Load AWS profiles for selection
	awsProfiles := loadAWSProfiles()
	awsProfileOptions := []list.Item{
		listItem{title: "default", description: "Default AWS profile"},
	}
	for _, profile := range awsProfiles {
		if profile != "default" {
			awsProfileOptions = append(awsProfileOptions, listItem{title: profile, description: fmt.Sprintf("AWS profile: %s", profile)})
		}
	}

	return &Model{
		step:            StepWelcome,
		projectName:     newTextInput("Project name:", "postservice"),
		modulePath:      newTextInput("Go module path:", "github.com/user/postservice"),
		outputDir:       newTextInput("Output directory:", "./postservice"),
		databaseSelect:  newSingleSelect("Select database:", databaseOptions),
		awsProfileSelect: newSingleSelect("Select AWS profile:", awsProfileOptions),
		awsAccessKeyID:  newTextInputWithSensitivity("AWS Access Key ID:", "", true),
		awsSecretKey:    newTextInputWithSensitivity("AWS Secret Access Key:", "", true),
		awsRegion:       newTextInput("AWS Region:", "us-east-1"),
		frameworkSelect: newSingleSelect("Select framework:", frameworkOptions),
		deployConfirm:   newConfirmWithDefault("Deploy to Fly.io immediately after generation?", false),
		spinner:         s,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

// loadAWSProfiles reads available AWS profiles from ~/.aws/credentials and ~/.aws/config
func loadAWSProfiles() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return []string{"default"}
	}

	profiles := make(map[string]bool)
	profiles["default"] = true

	// Read profiles from credentials file
	credentialsPath := fmt.Sprintf("%s/.aws/credentials", homeDir)
	if data, err := os.ReadFile(credentialsPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				profile := strings.Trim(line, "[]")
				if profile != "" {
					profiles[profile] = true
				}
			}
		}
	}

	// Read profiles from config file (profiles are defined as [profile profile-name])
	configPath := fmt.Sprintf("%s/.aws/config", homeDir)
	if data, err := os.ReadFile(configPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "[profile ") && strings.HasSuffix(line, "]") {
				// Extract profile name from [profile profile-name]
				profile := strings.TrimPrefix(line, "[profile ")
				profile = strings.TrimSuffix(profile, "]")
				profile = strings.TrimSpace(profile)
				if profile != "" {
					profiles[profile] = true
				}
			} else if strings.HasPrefix(line, "[default]") {
				profiles["default"] = true
			}
		}
	}

	// Convert map to sorted slice
	result := []string{"default"}
	for profile := range profiles {
		if profile != "default" {
			result = append(result, profile)
		}
	}
	// Sort non-default profiles
	if len(result) > 1 {
		// Simple alphabetical sort for non-default profiles
		for i := 1; i < len(result); i++ {
			for j := i + 1; j < len(result); j++ {
				if result[i] > result[j] {
					result[i], result[j] = result[j], result[i]
				}
			}
		}
	}

	return result
}

// loadAWSCredentialsFromProfile loads AWS credentials and region from a specific AWS profile
func loadAWSCredentialsFromProfile(profileName string) (accessKeyID, secretKey, region string) {
	ctx := context.Background()
	
	// Load config with specific profile
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profileName),
	)
	if err != nil {
		return "", "", ""
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return "", "", ""
	}

	// Get region from config, fallback to us-east-1 if not set
	region = cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	return creds.AccessKeyID, creds.SecretAccessKey, region
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.step > StepWelcome {
				m.step--
			}
		}

		switch m.step {
		case StepWelcome:
			if msg.String() == "enter" {
				m.step = StepProjectName
			}
			return m, nil
		case StepProjectName:
			var cmd tea.Cmd
			m.projectName, cmd = m.projectName.Update(msg)
			if msg.String() == "enter" && m.projectName.value != "" {
				// Set default output dir if empty
				if m.outputDir.value == "" || m.outputDir.value == "./postservice" {
					m.outputDir.SetValue("./" + m.projectName.value)
				}
				// Set default module path if empty
				if m.modulePath.value == "" || m.modulePath.value == "github.com/user/postservice" {
					m.modulePath.SetValue("github.com/user/" + m.projectName.value)
				}
				m.step = StepModulePath
			}
			return m, cmd
		case StepModulePath:
			var cmd tea.Cmd
			m.modulePath, cmd = m.modulePath.Update(msg)
			if msg.String() == "enter" && m.modulePath.value != "" {
				m.step = StepOutputDir
			}
			return m, cmd
		case StepOutputDir:
			var cmd tea.Cmd
			m.outputDir, cmd = m.outputDir.Update(msg)
			if msg.String() == "enter" && m.outputDir.value != "" {
				m.step = StepDatabaseSelection
			}
			return m, cmd
		case StepDatabaseSelection:
			var cmd tea.Cmd
			m.databaseSelect, cmd = m.databaseSelect.Update(msg)
			if msg.String() == "enter" && m.databaseSelect.GetSelected() != "" {
				// If DynamoDB is selected, show AWS profile selection
				if strings.Contains(m.databaseSelect.GetSelected(), "DynamoDB") {
					m.step = StepAWSProfileSelection
				} else {
					// Skip AWS credentials for PostgreSQL
					m.step = StepFrameworkSelection
				}
			}
			return m, cmd
		case StepAWSProfileSelection:
			var cmd tea.Cmd
			m.awsProfileSelect, cmd = m.awsProfileSelect.Update(msg)
			if msg.String() == "enter" && m.awsProfileSelect.GetSelected() != "" {
				// Load credentials and region from selected profile
				selectedProfile := m.awsProfileSelect.GetSelected()
				m.awsProfileName = selectedProfile
				accessKeyID, secretKey, region := loadAWSCredentialsFromProfile(selectedProfile)
				if accessKeyID != "" {
					m.awsAccessKeyID.SetValue(accessKeyID)
				}
				if secretKey != "" {
					m.awsSecretKey.SetValue(secretKey)
				}
				if region != "" {
					m.awsRegion.SetValue(region)
				}
				m.step = StepAWSAccessKeyID
			}
			return m, cmd
		case StepAWSAccessKeyID:
			var cmd tea.Cmd
			m.awsAccessKeyID, cmd = m.awsAccessKeyID.Update(msg)
			if msg.String() == "enter" && m.awsAccessKeyID.value != "" {
				m.step = StepAWSSecretKey
			}
			return m, cmd
		case StepAWSSecretKey:
			var cmd tea.Cmd
			m.awsSecretKey, cmd = m.awsSecretKey.Update(msg)
			if msg.String() == "enter" && m.awsSecretKey.value != "" {
				m.step = StepAWSRegion
			}
			return m, cmd
		case StepAWSRegion:
			var cmd tea.Cmd
			m.awsRegion, cmd = m.awsRegion.Update(msg)
			if msg.String() == "enter" && m.awsRegion.value != "" {
				m.step = StepFrameworkSelection
			}
			return m, cmd
		case StepFrameworkSelection:
			var cmd tea.Cmd
			m.frameworkSelect, cmd = m.frameworkSelect.Update(msg)
			if msg.String() == "enter" && m.frameworkSelect.GetSelected() != "" {
				m.step = StepDeploySelection
			}
			return m, cmd
		case StepDeploySelection:
			var cmd tea.Cmd
			m.deployConfirm, cmd = m.deployConfirm.Update(msg)
			if msg.String() == "enter" {
				m.step = StepReview
			}
			return m, cmd
		case StepReview:
			if msg.String() == "enter" {
				m.step = StepGenerating
				m.generating = true
				return m, tea.Batch(m.spinner.Tick, m.generate())
			}
		case StepComplete:
			if msg.String() == "enter" {
				return m, tea.Quit
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case GenerationCompleteMsg:
		if msg.ShouldDeploy {
			m.step = StepGenerating
			m.generating = true
			m.deploying = true
			return m, tea.Batch(m.spinner.Tick, m.deploy(msg.OutputDir, msg.ProjectName))
		}
		m.step = StepComplete
		m.generating = false
		return m, nil
	case DeploymentCompleteMsg:
		m.step = StepComplete
		m.generating = false
		m.deploying = false
		if msg.Error != nil {
			m.err = msg.Error
		}
		return m, nil
	case GenerationErrorMsg:
		m.err = msg.Err
		m.generating = false
		return m, nil
	}

	return m, nil
}

func (m *Model) generate() tea.Cmd {
	return func() tea.Msg {
		// Map database selection
		var dbType generator.DatabaseType
		selectedDB := m.databaseSelect.GetSelected()
		if strings.Contains(selectedDB, "PostgreSQL") {
			dbType = generator.DatabaseTypePostgres
		} else if strings.Contains(selectedDB, "DynamoDB") {
			dbType = generator.DatabaseTypeDynamoDB
		}

		// Map framework selection
		var frameworkType generator.FrameworkType
		selectedFramework := m.frameworkSelect.GetSelected()
		if strings.Contains(selectedFramework, "Chi") {
			frameworkType = generator.FrameworkTypeChi
		} else if strings.Contains(selectedFramework, "ConnectRPC") {
			frameworkType = generator.FrameworkTypeConnectRPC
		}

		cfg := generator.ProjectConfig{
			ProjectName: m.projectName.value,
			ModulePath:  m.modulePath.value,
			OutputDir:   m.outputDir.value,
			Database: generator.DatabaseConfig{
				Type:           dbType,
				AWSAccessKeyID: m.awsAccessKeyID.value,
				AWSSecretKey:   m.awsSecretKey.value,
				AWSRegion:      m.awsRegion.value,
			},
			Framework: frameworkType,
			Deploy:    true, // Always generate deployment files
		}

		gen := generator.NewGenerator(cfg)
		
		// Generate synchronously
		if err := gen.Generate(); err != nil {
			return GenerationErrorMsg{Err: err}
		}

		// Store deploy flag for completion message (whether to deploy now)
		m.deployEnabled = m.deployConfirm.GetChoice()

		// If user chose to deploy immediately, trigger deployment
		if m.deployEnabled {
			return GenerationCompleteMsg{
				ShouldDeploy: true,
				OutputDir:    cfg.OutputDir,
				ProjectName:  cfg.ProjectName,
			}
		}

		return GenerationCompleteMsg{ShouldDeploy: false}
	}
}

type GenerationCompleteMsg struct {
	ShouldDeploy bool
	OutputDir    string
	ProjectName  string
}
type GenerationErrorMsg struct {
	Err error
}
type DeploymentCompleteMsg struct {
	Success bool
	Error   error
}

func (m *Model) View() string {
	if m.generating {
		return m.renderGenerating()
	}

	if m.err != nil {
		return m.renderError()
	}

	switch m.step {
	case StepWelcome:
		return m.renderWelcome()
	case StepProjectName:
		return m.renderProjectName()
	case StepModulePath:
		return m.renderModulePath()
	case StepOutputDir:
		return m.renderOutputDir()
		case StepDatabaseSelection:
			return m.renderDatabaseSelection()
		case StepAWSProfileSelection:
			return m.renderAWSProfileSelection()
		case StepAWSAccessKeyID:
			return m.renderAWSAccessKeyID()
		case StepAWSSecretKey:
			return m.renderAWSSecretKey()
		case StepAWSRegion:
			return m.renderAWSRegion()
		case StepFrameworkSelection:
			return m.renderFrameworkSelection()
	case StepDeploySelection:
		return m.renderDeploySelection()
	case StepReview:
		return m.renderReview()
	case StepGenerating:
		return m.renderGenerating()
	case StepComplete:
		return m.renderComplete()
	default:
		return "Unknown step"
	}
}

func (m *Model) renderWelcome() string {
	logo := titleStyle.Render("🚀 create-go-api")
	subtitle := subtitleStyle.Render("Scaffold Go API Service Boilerplate")

	description := lipgloss.NewStyle().
		Foreground(whiteColor).
		MarginTop(2).
		MarginBottom(1).
		Render(`Welcome! This tool will help you create a production-ready
Go API service with:

  • REST API (Chi) or gRPC (ConnectRPC)
  • Database integration (PostgreSQL or DynamoDB)
  • Atlas migrations for PostgreSQL
  • One-click deployment (Fly.io)
  • Container-based testing`)

	help := helpStyle.Render("\nPress Enter to continue...")

	return lipgloss.JoinVertical(lipgloss.Left, logo, subtitle, description, help)
}

func (m *Model) renderProjectName() string {
	title := titleStyle.Render("📝 Project Configuration")
	form := m.projectName.View()
	help := helpStyle.Render("\nEnter: Continue  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", form, help)
}

func (m *Model) renderModulePath() string {
	title := titleStyle.Render("📦 Module Path")
	form := m.modulePath.View()
	help := helpStyle.Render("\nEnter: Continue  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", form, help)
}

func (m *Model) renderOutputDir() string {
	title := titleStyle.Render("📁 Output Directory")
	form := m.outputDir.View()
	help := helpStyle.Render("\nEnter: Continue  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", form, help)
}

func (m *Model) renderDatabaseSelection() string {
	title := titleStyle.Render("🗄️  Database Selection")
	form := m.databaseSelect.View()
	help := helpStyle.Render("\n↑/↓: Navigate  Enter: Select  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", form, help)
}

func (m *Model) renderAWSProfileSelection() string {
	title := titleStyle.Render("🔐 AWS Profile Selection")
	note := lipgloss.NewStyle().
		Foreground(whiteColor).
		MarginTop(1).
		MarginBottom(1).
		Render("Select an AWS profile. Credentials will be pre-filled from the selected profile.")
	form := m.awsProfileSelect.View()
	help := helpStyle.Render("\n↑/↓: Navigate  Enter: Select  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", note, form, help)
}

func (m *Model) renderAWSAccessKeyID() string {
	title := titleStyle.Render("🔑 AWS Access Key ID")
	profileNote := ""
	if m.awsProfileName != "" {
		profileNote = lipgloss.NewStyle().
			Foreground(whiteColor).
			MarginTop(1).
			Render(fmt.Sprintf("Profile: %s", m.awsProfileName))
	}
	note := lipgloss.NewStyle().
		Foreground(whiteColor).
		MarginTop(1).
		MarginBottom(1).
		Render("Pre-filled from your current AWS profile. You can override if needed.")
	form := m.awsAccessKeyID.View()
	help := helpStyle.Render("\nEnter: Continue  Esc: Back  Ctrl+C: Quit")

	content := []string{title, ""}
	if profileNote != "" {
		content = append(content, profileNote)
	}
	content = append(content, note, form, help)
	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func (m *Model) renderAWSSecretKey() string {
	title := titleStyle.Render("🔐 AWS Secret Access Key")
	profileNote := ""
	if m.awsProfileName != "" {
		profileNote = lipgloss.NewStyle().
			Foreground(whiteColor).
			MarginTop(1).
			Render(fmt.Sprintf("Profile: %s", m.awsProfileName))
	}
	note := lipgloss.NewStyle().
		Foreground(whiteColor).
		MarginTop(1).
		MarginBottom(1).
		Render("Pre-filled from your current AWS profile. You can override if needed.")
	form := m.awsSecretKey.View()
	help := helpStyle.Render("\nEnter: Continue  Esc: Back  Ctrl+C: Quit")

	content := []string{title, ""}
	if profileNote != "" {
		content = append(content, profileNote)
	}
	content = append(content, note, form, help)
	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func (m *Model) renderAWSRegion() string {
	title := titleStyle.Render("🌍 AWS Region")
	note := lipgloss.NewStyle().
		Foreground(whiteColor).
		MarginTop(1).
		MarginBottom(1).
		Render("Enter the AWS region where your DynamoDB table will be deployed (e.g., us-east-1, us-west-2, eu-west-1)")
	form := m.awsRegion.View()
	help := helpStyle.Render("\nEnter: Continue  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", note, form, help)
}

func (m *Model) renderFrameworkSelection() string {
	title := titleStyle.Render("🌐 Framework Selection")
	form := m.frameworkSelect.View()
	help := helpStyle.Render("\n↑/↓: Navigate  Enter: Select  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", form, help)
}

func (m *Model) renderDeploySelection() string {
	title := titleStyle.Render("🚀 Deployment")
	note := lipgloss.NewStyle().
		Foreground(whiteColor).
		MarginTop(1).
		MarginBottom(1).
		Render("Deployment files (Dockerfile, fly.toml, GitHub Actions) will always be generated.\nThis option controls whether to deploy immediately after generation.")
	form := m.deployConfirm.View()
	help := helpStyle.Render("\nY/N: Toggle  Enter: Continue  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", note, form, help)
}

func (m *Model) renderReview() string {
	title := titleStyle.Render("📋 Review Configuration")
	
	var deployText string
	if m.deployConfirm.GetChoice() {
		deployText = successStyle.Render("Yes")
	} else {
		deployText = unselectedStyle.Render("No")
	}

	reviewItems := []string{
		labelStyle.Render("Project Name:") + " " + valueStyle.Render(m.projectName.value),
		labelStyle.Render("Module Path:") + " " + valueStyle.Render(m.modulePath.value),
		labelStyle.Render("Output Dir:") + " " + valueStyle.Render(m.outputDir.value),
		labelStyle.Render("Database:") + " " + valueStyle.Render(m.databaseSelect.GetSelected()),
	}

	// Show AWS credentials only if DynamoDB is selected
	if strings.Contains(m.databaseSelect.GetSelected(), "DynamoDB") {
		if m.awsProfileName != "" {
			reviewItems = append(reviewItems,
				labelStyle.Render("AWS Profile:")+" "+valueStyle.Render(m.awsProfileName),
			)
		}
		reviewItems = append(reviewItems,
			labelStyle.Render("AWS Access Key ID:")+" "+valueStyle.Render(maskString(m.awsAccessKeyID.value)),
			labelStyle.Render("AWS Secret Key:")+" "+valueStyle.Render(maskString(m.awsSecretKey.value)),
			labelStyle.Render("AWS Region:")+" "+valueStyle.Render(m.awsRegion.value),
		)
	}

	reviewItems = append(reviewItems,
		labelStyle.Render("Framework:")+" "+valueStyle.Render(m.frameworkSelect.GetSelected()),
		labelStyle.Render("Deploy Now:")+" "+deployText,
		"",
		helpStyle.Render("Press Enter to generate, Esc to go back, Ctrl+C to quit"),
	)

	content := lipgloss.JoinVertical(lipgloss.Left, reviewItems...)

	return lipgloss.JoinVertical(lipgloss.Left, title, "", content)
}

func (m *Model) renderGenerating() string {
	var title, message string
	if m.deploying {
		title = titleStyle.Render("🚀 Deploying to Fly.io...")
		message = "Deploying application..."
	} else {
		title = titleStyle.Render("⚙️  Generating Project...")
		message = "Generating project files..."
	}
	spinner := m.spinner.View()
	
	content := lipgloss.JoinVertical(lipgloss.Left,
		spinner+" "+message,
		"",
		helpStyle.Render("Please wait..."),
	)

	return lipgloss.JoinVertical(lipgloss.Left, title, "", content)
}

// deploy attempts to deploy the project to Fly.io
func (m *Model) deploy(outputDir, projectName string) tea.Cmd {
	return func() tea.Msg {
		// Check if flyctl or fly command exists in PATH
		var flyCmd string
		if path, err := exec.LookPath("flyctl"); err == nil && path != "" {
			flyCmd = "flyctl"
		} else if path, err := exec.LookPath("fly"); err == nil && path != "" {
			flyCmd = "fly"
		}

		if flyCmd == "" {
			return DeploymentCompleteMsg{
				Success: false,
				Error: fmt.Errorf("flyctl or fly command not found. Please install from https://fly.io/docs/getting-started/installing-flyctl/"),
			}
		}

		// Use fly launch to create and deploy the app (non-interactive, reuse fly.toml)
		cmd := exec.Command(flyCmd, "launch", "--name", projectName, "--copy-config", "--yes")
		cmd.Dir = outputDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			return DeploymentCompleteMsg{
				Success: false,
				Error: fmt.Errorf("deployment failed: %w\nOutput: %s", err, string(output)),
			}
		}

		return DeploymentCompleteMsg{Success: true}
	}
}

func (m *Model) renderComplete() string {
	var title string
	if m.err != nil {
		// Check if it's a deployment error
		if strings.Contains(m.err.Error(), "deployment") || strings.Contains(m.err.Error(), "flyctl") || strings.Contains(m.err.Error(), "fly") {
			title = successStyle.Render("✓ Project Generated Successfully!")
			content := lipgloss.JoinVertical(lipgloss.Left,
				"",
				valueStyle.Render("Project:")+" "+m.projectName.value,
				valueStyle.Render("Module:")+" "+m.modulePath.value,
				valueStyle.Render("Output:")+" "+m.outputDir.value,
				"",
				errorStyle.Render("⚠ Deployment failed:"),
				errorStyle.Render(m.err.Error()),
				"",
				subtitleStyle.Render("Next steps:"),
				"  cd "+m.outputDir.value,
				"  make deps",
				"  make build",
				"  make deploy  # Try deploying manually",
				"",
				helpStyle.Render("Press Enter to exit"),
			)
			return lipgloss.JoinVertical(lipgloss.Left, title, content)
		}
		// Other errors handled by renderError
		return m.renderError()
	}
	
	title = successStyle.Render("✓ Project Generated Successfully!")
	
	nextSteps := []string{
		"  cd " + m.outputDir.value,
		"  make deps",
		"  make build",
		"  make deploy",
	}
	
	nextSteps = append(nextSteps, "")
	
	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		valueStyle.Render("Project:")+" "+m.projectName.value,
		valueStyle.Render("Module:")+" "+m.modulePath.value,
		valueStyle.Render("Output:")+" "+m.outputDir.value,
		"",
		subtitleStyle.Render("Next steps:"),
		strings.Join(nextSteps, "\n"),
		helpStyle.Render("Press Enter to exit"),
	)

	return lipgloss.JoinVertical(lipgloss.Left, title, content)
}

func (m *Model) renderError() string {
	title := errorStyle.Render("✗ Error")
	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		errorStyle.Render(m.err.Error()),
		"",
		helpStyle.Render("Press Ctrl+C to exit"),
	)

	return lipgloss.JoinVertical(lipgloss.Left, title, content)
}



