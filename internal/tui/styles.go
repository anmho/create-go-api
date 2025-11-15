package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("39")
	secondaryColor = lipgloss.Color("86")
	successColor   = lipgloss.Color("46")
	warningColor   = lipgloss.Color("226")
	errorColor     = lipgloss.Color("196")
	grayColor      = lipgloss.Color("240")
	whiteColor     = lipgloss.Color("255")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(1, 2).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	unselectedStyle = lipgloss.NewStyle().
			Foreground(grayColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(grayColor).
			Italic(true).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	labelStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			Bold(true).
			MarginRight(2)

	valueStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)
)

