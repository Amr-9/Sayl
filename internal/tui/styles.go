package tui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Shared Styles
var (
	// Brand Colors - Enhanced Palette
	primaryColor   = lipgloss.Color("#00FFFF") // Cyan/Aqua
	secondaryColor = lipgloss.Color("#FF6B9D") // Pink
	accentColor    = lipgloss.Color("#00FF88") // Green
	purpleColor    = lipgloss.Color("#BD93F9") // Purple
	orangeColor    = lipgloss.Color("#FFB86C") // Orange
	yellowColor    = lipgloss.Color("#F1FA8C") // Yellow
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

	// Enhanced Dashboard Styles
	boxTitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	dashBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(0, 1).
			MarginRight(1)

	headerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(primaryColor).
			Padding(0, 2).
			Align(lipgloss.Center)

	targetStyle = lipgloss.NewStyle().
			Foreground(yellowColor).
			Bold(true)

	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444"))

	// Status Code Bar Styles
	barFullStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	barEmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333"))

	sparklineStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	// Animation chars
	spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
)

// Enhanced ASCII Logo
const asciiLogo = `⚡ SAYL`

const bigAsciiLogo = `
 ███████╗ █████╗ ██╗   ██╗██╗     
 ██╔════╝██╔══██╗╚██╗ ██╔╝██║     
 ███████╗███████║ ╚████╔╝ ██║     
 ╚════██║██╔══██║  ╚██╔╝  ██║     
 ███████║██║  ██║   ██║   ███████╗
 ╚══════╝╚═╝  ╚═╝   ╚═╝   ╚══════╝`

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

// RenderBar creates a horizontal bar chart
func RenderBar(value, max int, width int, style lipgloss.Style) string {
	if max == 0 {
		return ""
	}
	filled := (value * width) / max
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := filled; i < width; i++ {
		bar += "░"
	}

	return style.Render(bar)
}

// RenderPercentBar renders a bar with percentage
func RenderPercentBar(value, total int, width int) string {
	if total == 0 {
		return ""
	}

	filled := int(float64(width) * float64(value) / float64(total))
	if filled > width {
		filled = width
	}

	// Choose color based on context (will be set by caller)
	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}

	return bar
}

// GetSpinnerFrame returns the current spinner frame based on tick
func GetSpinnerFrame(tick int) string {
	return spinnerFrames[tick%len(spinnerFrames)]
}
