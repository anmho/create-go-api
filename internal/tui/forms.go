package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// item implements list.Item for single select
type listItem struct {
	title       string
	description string
}

func (i listItem) FilterValue() string { return i.title }
func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.description }

type textInputModel struct {
	textinput textinput.Model
	label     string
	value     string
	focused   bool
	sensitive bool // If true, redact the value in display
}

func newTextInput(label, placeholder string) textInputModel {
	return newTextInputWithSensitivity(label, placeholder, false)
}

func newTextInputWithSensitivity(label, placeholder string, sensitive bool) textInputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 60
	ti.PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
	ti.TextStyle = lipgloss.NewStyle().Foreground(whiteColor)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(primaryColor)
	
	// Don't use EchoPassword mode - we'll handle masking in View() to show last 4 chars

	return textInputModel{
		textinput: ti,
		label:     label,
		focused:   true,
		sensitive: sensitive,
	}
}

func (m textInputModel) Update(msg tea.Msg) (textInputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	m.value = m.textinput.Value()
	return m, cmd
}

func (m *textInputModel) SetValue(value string) {
	m.textinput.SetValue(value)
	m.value = value
}

// maskString masks a string for display (redacts all but last 4 characters)
func maskString(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) <= 4 {
		// If 4 or fewer characters, mask everything
		return strings.Repeat("*", len(s))
	}
	// Redact all characters except the last 4
	return strings.Repeat("*", len(s)-4) + s[len(s)-4:]
}

func (m textInputModel) View() string {
	var inputView string
	if m.sensitive {
		// For sensitive fields, show masked value with last 4 characters visible
		maskedValue := maskString(m.value)
		// Get the prompt from the textinput
		prompt := m.textinput.Prompt
		
		// Build the input view with cursor at end (most common case when typing)
		// The cursor position matches the actual value length
		promptStyle := lipgloss.NewStyle().Foreground(primaryColor)
		textStyle := lipgloss.NewStyle().Foreground(whiteColor)
		cursorStyle := lipgloss.NewStyle().Foreground(primaryColor)
		
		// Cursor is typically at the end when typing
		displayValue := maskedValue + cursorStyle.Render("█")
		
		inputView = promptStyle.Render(prompt) + textStyle.Render(displayValue)
	} else {
		inputView = m.textinput.View()
	}
	
	return lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render(m.label),
		"",
		inputView,
	)
}

type singleSelectModel struct {
	options  []string
	selected int
	cursor   int
	label    string
	values   []string
}

func newSingleSelect(label string, items []list.Item) singleSelectModel {
	options := make([]string, len(items))
	for i, item := range items {
		options[i] = item.(listItem).title
	}

	return singleSelectModel{
		options:  options,
		selected: 0,
		cursor:   0,
		label:    label,
		values:   options,
	}
}

func (m singleSelectModel) Update(msg tea.Msg) (singleSelectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case " ", "enter":
			m.selected = m.cursor
		}
	}
	return m, nil
}

func (m singleSelectModel) View() string {
	var items []string
	items = append(items, labelStyle.Render(m.label))
	items = append(items, "")

	for i, option := range m.options {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		radio := "( )"
		if m.selected == i {
			radio = selectedStyle.Render("(●)")
		}

		style := unselectedStyle
		if m.cursor == i {
			style = selectedStyle
		}

		items = append(items, style.Render(cursor+" "+radio+" "+option))
	}

	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func (m singleSelectModel) GetSelected() string {
	if m.selected >= 0 && m.selected < len(m.values) {
		return m.values[m.selected]
	}
	return ""
}

type confirmModel struct {
	choice string
	label  string
}

func newConfirmWithDefault(label string, defaultYes bool) confirmModel {
	choice := "no"
	if defaultYes {
		choice = "yes"
	}
	return confirmModel{
		label:  label,
		choice: choice,
	}
}

func (m confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "y", "Y":
			m.choice = "yes"
		case "n", "N":
			m.choice = "no"
		case "enter":
			// Enter confirms current choice
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	yesStyle := unselectedStyle
	noStyle := unselectedStyle
	if m.choice == "yes" {
		yesStyle = selectedStyle
	} else {
		noStyle = selectedStyle
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render(m.label),
		"",
		lipgloss.JoinHorizontal(lipgloss.Left,
			yesStyle.Render("[Y] Yes"),
			"  ",
			noStyle.Render("[N] No"),
		),
	)
}

func (m confirmModel) GetChoice() bool {
	return m.choice == "yes"
}

