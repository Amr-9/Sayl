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

// Summary Section Styles
var (
	sumSectionStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginTop(1)

	sumLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	sumValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Bold(true)

	sumBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(0, 2).
			MarginRight(1)
)

func (m *SummaryModel) View() string {
	var s strings.Builder

	// ═══════════════════════════════════════════════════════════════
	// HEADER
	// ═══════════════════════════════════════════════════════════════

	logoLines := strings.Split(bigAsciiLogo, "\n")
	styledLogo := ""
	for _, line := range logoLines {
		if line != "" {
			styledLogo += lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(line) + "\n"
		}
	}

	headerContent := styledLogo
	headerContent += lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true).Render("  High-Performance Load Testing Tool • v0.2")

	s.WriteString(headerBoxStyle.Render(headerContent))
	s.WriteString("\n\n")

	// Test Complete Banner
	completeBanner := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		Render("✨ TEST COMPLETED SUCCESSFULLY ✨")
	s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Render(completeBanner))
	s.WriteString("\n\n")

	// ═══════════════════════════════════════════════════════════════
	// SUMMARY BOXES
	// ═══════════════════════════════════════════════════════════════

	// Section Title
	s.WriteString(lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render("📊 Test Summary"))
	s.WriteString("\n\n")

	// Calculate success/failure percentages
	totalReqs := m.report.SuccessCount + m.report.FailureCount
	var successPct, failPct float64
	if totalReqs > 0 {
		successPct = float64(m.report.SuccessCount) / float64(totalReqs) * 100
		failPct = float64(m.report.FailureCount) / float64(totalReqs) * 100
	}

	// Box 1: Traffic Summary
	trafficContent := fmt.Sprintf("%s\n\n%s  %s\n%s  %s\n%s  %s\n%s  %s",
		lipgloss.NewStyle().Foreground(purpleColor).Bold(true).Render("🚀 Traffic Summary"),
		sumLabelStyle.Width(16).Render("Total Requests:"),
		sumValueStyle.Render(fmt.Sprintf("%d", m.report.TotalRequests)),
		sumLabelStyle.Width(16).Render("RPS (Avg):"),
		sumValueStyle.Render(fmt.Sprintf("%.2f", m.report.RPS)),
		sumLabelStyle.Width(16).Render("Total Data:"),
		sumValueStyle.Render(formatBytes(m.report.TotalBytes)),
		sumLabelStyle.Width(16).Render("Throughput:"),
		sumValueStyle.Render(formatThroughput(m.report.TotalBytes, m.report.Duration.Seconds())))

	box1 := sumBoxStyle.Copy().BorderForeground(purpleColor).Width(36).Render(trafficContent)

	// Box 2: Results
	resultsContent := fmt.Sprintf("%s\n\n%s  %s\n%s  %s %s\n%s  %s %s\n%s  %s",
		lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("✅ Results"),
		sumLabelStyle.Width(12).Render("Duration:"),
		sumValueStyle.Render(m.report.Duration.String()),
		sumLabelStyle.Width(12).Render("Success:"),
		successText.Bold(true).Render(fmt.Sprintf("%d", m.report.SuccessCount)),
		successText.Render(fmt.Sprintf("(%.1f%%)", successPct)),
		sumLabelStyle.Width(12).Render("Failed:"),
		errText.Bold(true).Render(fmt.Sprintf("%d", m.report.FailureCount)),
		errText.Render(fmt.Sprintf("(%.1f%%)", failPct)),
		sumLabelStyle.Width(12).Render("Workers:"),
		sumValueStyle.Render(fmt.Sprintf("%d", m.report.Concurrency)))

	box2 := sumBoxStyle.Copy().BorderForeground(accentColor).Width(36).Render(resultsContent)

	row1 := lipgloss.JoinHorizontal(lipgloss.Top, box1, box2)
	s.WriteString(row1)
	s.WriteString("\n\n")

	// ═══════════════════════════════════════════════════════════════
	// LATENCY DISTRIBUTION
	// ═══════════════════════════════════════════════════════════════

	s.WriteString(lipgloss.NewStyle().Foreground(orangeColor).Bold(true).Render("⏱️  Latency Distribution"))
	s.WriteString("\n")

	latencyBox := sumBoxStyle.Copy().BorderForeground(orangeColor).Width(74)

	latencies := []struct {
		name  string
		value string
	}{
		{"Min", fmtDuration(m.report.Min)},
		{"P50", fmtDuration(m.report.P50)},
		{"P75", fmtDuration(m.report.P75)},
		{"P90", fmtDuration(m.report.P90)},
		{"P95", fmtDuration(m.report.P95)},
		{"P99", fmtDuration(m.report.P99)},
		{"Max", fmtDuration(m.report.Max)},
	}

	// Create latency grid (2 rows)
	var latencyContent strings.Builder
	for i, lat := range latencies {
		latencyContent.WriteString(fmt.Sprintf("%s %s",
			sumLabelStyle.Width(5).Render(lat.name+":"),
			sumValueStyle.Width(12).Render(lat.value)))
		if i < len(latencies)-1 {
			latencyContent.WriteString("  │  ")
		}
	}

	s.WriteString(latencyBox.Render(latencyContent.String()))
	s.WriteString("\n\n")

	// ═══════════════════════════════════════════════════════════════
	// STATUS CODES BAR CHART
	// ═══════════════════════════════════════════════════════════════

	if len(m.report.StatusCodes) > 0 {
		s.WriteString(lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render("📊 Status Codes"))
		s.WriteString("\n")

		// Sort codes by count
		type kv struct {
			Code  string
			Count int
		}
		var sorted []kv
		for k, v := range m.report.StatusCodes {
			sorted = append(sorted, kv{Code: k, Count: v})
		}
		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i].Count < sorted[j].Count {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		maxCount := 0
		for _, item := range sorted {
			if item.Count > maxCount {
				maxCount = item.Count
			}
		}

		barWidth := 20

		var codesContent strings.Builder
		for _, item := range sorted {
			label := item.Code
			barStyle := successText

			// Status code styling
			if label == "0" {
				label = "NetErr/Timeout"
				barStyle = errText
			} else {
				var codeInt int
				n, _ := fmt.Sscanf(label, "%d", &codeInt)
				if n > 0 {
					if codeInt >= 200 && codeInt < 300 {
						label = fmt.Sprintf("%s OK", item.Code)
						barStyle = successText
					} else if codeInt >= 300 && codeInt < 400 {
						label = fmt.Sprintf("%s Redirect", item.Code)
						barStyle = warnText
					} else if codeInt >= 400 && codeInt < 500 {
						label = fmt.Sprintf("%s Client Err", item.Code)
						barStyle = warnText
					} else if codeInt >= 500 {
						label = fmt.Sprintf("%s Server Err", item.Code)
						barStyle = errText
					}
				} else {
					barStyle = errText
				}
			}

			// Calculate bar
			barLen := 0
			if maxCount > 0 {
				barLen = (item.Count * barWidth) / maxCount
			}
			if barLen < 1 && item.Count > 0 {
				barLen = 1
			}

			// Build bar with fixed width
			bar := ""
			for i := 0; i < barLen; i++ {
				bar += "█"
			}
			for i := barLen; i < barWidth; i++ {
				bar += "░"
			}

			pctVal := float64(0)
			if m.report.TotalRequests > 0 {
				pctVal = float64(item.Count) / float64(m.report.TotalRequests) * 100
			}

			// Pad label to fixed width for alignment
			paddedLabel := label
			for len(paddedLabel) < 16 {
				paddedLabel += " "
			}

			codesContent.WriteString(fmt.Sprintf("  %s %s %6d %s\n",
				sumLabelStyle.Render(paddedLabel),
				barStyle.Render(bar),
				item.Count,
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(fmt.Sprintf("(%5.1f%%)", pctVal))))
		}

		s.WriteString(sumBoxStyle.Copy().BorderForeground(primaryColor).Width(70).Render(codesContent.String()))
		s.WriteString("\n")
	}

	// ═══════════════════════════════════════════════════════════════
	// ERROR BREAKDOWN (if any)
	// ═══════════════════════════════════════════════════════════════

	if len(m.report.Errors) > 0 {
		s.WriteString("\n")
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true).Render("❌ Error Breakdown"))
		s.WriteString("\n")

		var errContent strings.Builder
		i := 0
		for errStr, count := range m.report.Errors {
			if i >= 5 {
				errContent.WriteString(fmt.Sprintf("  ... and %d more error types\n", len(m.report.Errors)-5))
				break
			}
			i++

			cleanErr := errStr
			if len(cleanErr) > 55 {
				cleanErr = cleanErr[:52] + "..."
			}

			errContent.WriteString(fmt.Sprintf("  %s  %s\n",
				sumLabelStyle.Width(55).Render(cleanErr),
				errText.Bold(true).Render(fmt.Sprintf("×%d", count))))
		}

		s.WriteString(sumBoxStyle.Copy().BorderForeground(lipgloss.Color("#FF4444")).Width(74).Render(errContent.String()))
		s.WriteString("\n")
	}

	// ═══════════════════════════════════════════════════════════════
	// FOOTER
	// ═══════════════════════════════════════════════════════════════

	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Foreground(accentColor).Render("📁 Report saved to ") +
		lipgloss.NewStyle().Foreground(yellowColor).Bold(true).Render("report.json"))
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Press Ctrl+C to exit."))

	return s.String()
}
