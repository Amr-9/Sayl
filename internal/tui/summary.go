package tui

import (
	"fmt"
	"strings"

	"github.com/Amr-9/sayl/pkg/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SummaryModel struct {
	report models.Report
}

func NewSummaryModel(report models.Report) *SummaryModel {
	return &SummaryModel{
		report: report,
	}
}

func (m *SummaryModel) Init() tea.Cmd {
	return nil
}

func (m *SummaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// Styles
var (
	sumHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			MarginBottom(1)

	sumStatStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginRight(2)

	sumValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Bold(true)
)

func (m *SummaryModel) View() string {
	var s strings.Builder

	// Header
	logo := logoStyle.Render(asciiLogo)
	s.WriteString(borderStyle.Render(logo))
	s.WriteString("\n")
	s.WriteString(subtitleStyle.Render("High-Performance Load Testing Tool • v1.0"))
	s.WriteString("\n\n")

	s.WriteString(sumHeaderStyle.Render("📊 Test Summary"))
	s.WriteString("\n\n")

	// Stats
	// Section 1: Traffic & Throughput
	s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("🚀 Traffic & Throughput"))
	s.WriteString("\n")

	tData := [][]string{
		{"Total Requests", fmt.Sprintf("%d", m.report.TotalRequests)},
		{"Success Rate", fmt.Sprintf("%.2f%%", m.report.SuccessRate)},
		{"RPS (Avg)", fmt.Sprintf("%.2f", m.report.RPS)},
		{"Total Data", formatBytes(m.report.TotalBytes)},
		{"Throughput", formatThroughput(m.report.TotalBytes, m.report.Duration.Seconds())},
		{"Duration", m.report.Duration.String()},
	}

	for _, row := range tData {
		s.WriteString(fmt.Sprintf("  %s %s\n", sumStatStyle.Render(fmt.Sprintf("%-15s", row[0]+":")), sumValueStyle.Render(row[1])))
	}
	s.WriteString("\n")

	s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render("Latency Distribution:"))
	s.WriteString("\n")

	lData := [][]string{
		{"Min", fmtDuration(m.report.Min)},
		{"P50", fmtDuration(m.report.P50)},
		{"P75", fmtDuration(m.report.P75)},
		{"P90", fmtDuration(m.report.P90)},
		{"P95", fmtDuration(m.report.P95)},
		{"P99", fmtDuration(m.report.P99)},
		{"Max", fmtDuration(m.report.Max)},
	}

	// 2 columns layout for latency
	for i := 0; i < len(lData); i += 2 {
		r1 := lData[i]
		s.WriteString(fmt.Sprintf("  %s %s", sumStatStyle.Render(fmt.Sprintf("%-5s", r1[0]+":")), sumValueStyle.Render(fmt.Sprintf("%-12s", r1[1]))))
		if i+1 < len(lData) {
			r2 := lData[i+1]
			s.WriteString(fmt.Sprintf("  %s %s", sumStatStyle.Render(fmt.Sprintf("%-5s", r2[0]+":")), sumValueStyle.Render(r2[1])))
		}
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// Section 3: Status Codes
	if len(m.report.StatusCodes) > 0 {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("📊 Status Codes"))
		s.WriteString("\n")

		// Sort codes
		var codes []string
		for k := range m.report.StatusCodes {
			codes = append(codes, k)
		}
		// Bubblesort for simplicity
		for i := 0; i < len(codes); i++ {
			for j := i + 1; j < len(codes); j++ {
				if codes[i] > codes[j] {
					codes[i], codes[j] = codes[j], codes[i]
				}
			}
		}

		for _, code := range codes {
			count := m.report.StatusCodes[code]
			label := fmt.Sprintf("Code %s", code)
			style := sumValueStyle

			// Detect status type
			var codeInt int
			n, _ := fmt.Sscanf(code, "%d", &codeInt)

			if n > 0 {
				// Numeric Code
				if codeInt >= 400 {
					style = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
				} else {
					style = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
				}
			} else {
				// Text (e.g. "Timeout")
				label = code
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
			}

			s.WriteString(fmt.Sprintf("  %s %s\n", sumStatStyle.Render(fmt.Sprintf("%-15s", label+":")), style.Render(fmt.Sprintf("%d", count))))
		}
		s.WriteString("\n")

		// If there are detailed errors (code 0), show breakdown
		if len(m.report.Errors) > 0 {
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render("❌ Error Breakdown"))
			s.WriteString("\n")
			for errStr, count := range m.report.Errors {
				// Shorten error if too long?
				cleanErr := errStr
				if len(cleanErr) > 50 {
					cleanErr = cleanErr[:47] + "..."
				}
				s.WriteString(fmt.Sprintf("  %s %s\n", sumStatStyle.Render(fmt.Sprintf("%-30s", cleanErr+":")), sumValueStyle.Render(fmt.Sprintf("%d", count))))
			}
		}
	}

	s.WriteString("\n")
	s.WriteString(highlight.Render("✨ Report saved to report.json"))
	s.WriteString("\n" + subtext.Render("Press Ctrl+C to exit."))

	return s.String()
}
