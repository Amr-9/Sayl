package tui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Shared Styles
var (
	// Brand Colors
	primaryColor   = lipgloss.Color("#00FFFF") // Cyan/Aqua
	secondaryColor = lipgloss.Color("#FF6B9D") // Pink
	accentColor    = lipgloss.Color("#00FF88") // Green
	subColor       = lipgloss.Color("241")     // Grey

	// Global Styles
	logoStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true).
			MarginLeft(1)

	// Step/Form Styles
	highlight = lipgloss.NewStyle().Foreground(secondaryColor)
	subtext   = lipgloss.NewStyle().Foreground(subColor)
	check     = lipgloss.NewStyle().Foreground(accentColor) // Green

	questionHeader = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00AAFF")).
			Bold(true).
			MarginTop(1)

	finalValue = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	// Dashboard Specific
	successText = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88")) // Bright Green
	warnText    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")) // Gold
	errText     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")) // Red
	infoText    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")) // Cyan
)

// Smaller, cleaner ASCII logo
const asciiLogo = `⚡ SAYL`

// MakeNeonTheme creates a custom theme for huh forms
func MakeNeonTheme() *huh.Theme {
	t := huh.ThemeCharm()
	t.Focused.Title = t.Focused.Title.Foreground(primaryColor).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(subColor)
	t.Focused.Base = t.Focused.Base.BorderForeground(secondaryColor)
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(secondaryColor)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(lipgloss.Color("240"))
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(accentColor).SetString("› ")
	t.Focused.Option = t.Focused.Option.Foreground(lipgloss.Color("250"))
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(primaryColor).Bold(true)
	return t
}
