package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Amr-9/sayl/pkg/models"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DashModel struct {
	config   models.Config
	report   models.Report
	start    time.Time
	progress progress.Model
	history  []string // Rendered history lines
	tick     int      // For animations
}

func NewDashModel(cfg models.Config, history []string) *DashModel {
	p := progress.New(
		progress.WithScaledGradient("#00FFFF", "#FF6B9D"),
		progress.WithoutPercentage(),
	)
	return &DashModel{
		config:   cfg,
		report:   models.Report{},
		start:    time.Now(),
		progress: p,
		history:  history,
		tick:     0,
	}
}

func (m *DashModel) Init() tea.Cmd {
	return nil
}

func (m *DashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case models.Report:
		m.report = msg
		m.tick++
	}
	return m, nil
}

func (m *DashModel) View() string {
	var s strings.Builder

	// ═══════════════════════════════════════════════════════════════
	// HEADER SECTION
	// ═══════════════════════════════════════════════════════════════

	// Render enhanced header with logo
	logoLines := strings.Split(bigAsciiLogo, "\n")
	styledLogo := ""
	for _, line := range logoLines {
		if line != "" {
			styledLogo += lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(line) + "\n"
		}
	}

	headerContent := styledLogo
	headerContent += lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true).Render("  High-Performance Load Testing Tool")

	s.WriteString(headerBoxStyle.Render(headerContent))
	s.WriteString("\n\n")

	// Target Info Line
	timeoutDisplay := m.config.Timeout
	if timeoutDisplay == 0 {
		timeoutDisplay = 10 * time.Second // Default timeout
	}
	targetLine := fmt.Sprintf("🎯 %s  %s",
		targetStyle.Render(m.config.URL),
		metaStyle.Render(fmt.Sprintf("│ %s │ %d workers │ %v timeout",
			m.config.Method, m.config.Concurrency, timeoutDisplay)))
	s.WriteString(targetLine)
	s.WriteString("\n\n")

	// ═══════════════════════════════════════════════════════════════
	// PROGRESS BAR SECTION
	// ═══════════════════════════════════════════════════════════════

	elapsed := time.Since(m.start)
	pct := float64(elapsed) / float64(m.config.Duration)
	if pct > 1.0 {
		pct = 1.0
	}

	remaining := m.config.Duration - elapsed
	if remaining < 0 {
		remaining = 0
	}

	// Divider
	s.WriteString(dividerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	s.WriteString("\n")

	// Progress bar with spinner
	spinner := GetSpinnerFrame(m.tick)
	progressBar := m.progress.ViewAs(pct)
	timeInfo := fmt.Sprintf("%s  %s / %s  (remaining: %s)",
		lipgloss.NewStyle().Foreground(accentColor).Render(spinner),
		lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(elapsed.Round(time.Second).String()),
		m.config.Duration.String(),
		lipgloss.NewStyle().Foreground(orangeColor).Render(remaining.Round(time.Second).String()))

	s.WriteString(progressBar)
	s.WriteString("\n")
	s.WriteString(timeInfo)
	s.WriteString("\n")
	s.WriteString(dividerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	s.WriteString("\n\n")

	// ═══════════════════════════════════════════════════════════════
	// METRICS BOXES
	// ═══════════════════════════════════════════════════════════════

	// Calculate values
	rps := fmt.Sprintf("%.1f", m.report.RPS)
	tput := formatThroughput(m.report.TotalBytes, elapsed.Seconds())
	totalData := formatBytes(m.report.TotalBytes)

	// Sparkline
	var rpsHistory []int
	maxLen := 20
	startIdx := 0
	if len(m.report.TimeSeriesData) > maxLen {
		startIdx = len(m.report.TimeSeriesData) - maxLen
	}
	for i := startIdx; i < len(m.report.TimeSeriesData); i++ {
		rpsHistory = append(rpsHistory, int(m.report.TimeSeriesData[i].Requests))
	}
	spark := renderSparkline(rpsHistory)

	// BOX 1: Performance
	box1Content := fmt.Sprintf("%s\n%s %s\n%s %s\n%s %s\n%s",
		lipgloss.NewStyle().Foreground(purpleColor).Bold(true).Render("📈 Performance"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("RPS:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true).Render(rps),
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Flow:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true).Render(tput),
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Data:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true).Render(totalData),
		sparklineStyle.Render(spark))

	box1 := dashBoxStyle.Copy().BorderForeground(purpleColor).Width(24).Render(box1Content)

	// BOX 2: Latency
	p50 := fmtDuration(m.report.P50)
	p90 := fmtDuration(m.report.P90)
	p99 := fmtDuration(m.report.P99)
	maxLat := fmtDuration(m.report.Max)

	box2Content := fmt.Sprintf("%s\n%s %s\n%s %s\n%s %s\n%s %s",
		lipgloss.NewStyle().Foreground(orangeColor).Bold(true).Render("⏱️  Latency"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("P50:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true).Render(p50),
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("P90:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true).Render(p90),
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("P99:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true).Render(p99),
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Max:"),
		lipgloss.NewStyle().Foreground(yellowColor).Bold(true).Render(maxLat))

	box2 := dashBoxStyle.Copy().BorderForeground(orangeColor).Width(24).Render(box2Content)

	// BOX 3: Results
	totalReqs := m.report.SuccessCount + m.report.FailureCount
	var successPct, failPct float64
	if totalReqs > 0 {
		successPct = (float64(m.report.SuccessCount) / float64(totalReqs)) * 100.0
		failPct = (float64(m.report.FailureCount) / float64(totalReqs)) * 100.0
	}

	// Color coding for failure rate
	failColor := successText
	if failPct > 0 {
		failColor = warnText
	}
	if failPct > 5.0 {
		failColor = errText
	}

	box3Content := fmt.Sprintf("%s\n%s %s\n%s %s %s\n%s %s %s",
		lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("✅ Results"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Total:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true).Render(fmt.Sprintf("%d", totalReqs)),
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Success:"),
		successText.Bold(true).Render(fmt.Sprintf("%d", m.report.SuccessCount)),
		successText.Render(fmt.Sprintf("(%.1f%%)", successPct)),
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Failed:"),
		failColor.Bold(true).Render(fmt.Sprintf("%d", m.report.FailureCount)),
		failColor.Render(fmt.Sprintf("(%.1f%%)", failPct)))

	box3 := dashBoxStyle.Copy().BorderForeground(accentColor).Width(26).Render(box3Content)

	// Join boxes horizontally
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, box1, box2, box3)
	s.WriteString(row1)
	s.WriteString("\n\n")

	// ═══════════════════════════════════════════════════════════════
	// STATUS CODES SECTION (Bar Chart Style)
	// ═══════════════════════════════════════════════════════════════

	s.WriteString(lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render("📊 Status Codes"))
	s.WriteString("\n")

	if len(m.report.StatusCodes) > 0 {
		// Sort codes
		type kv struct {
			Code  string
			Count int
		}
		var sorted []kv
		for k, v := range m.report.StatusCodes {
			sorted = append(sorted, kv{Code: k, Count: v})
		}
		// Sort by count descending
		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i].Count < sorted[j].Count {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		// Find max for bar scaling
		maxCount := 0
		for _, item := range sorted {
			if item.Count > maxCount {
				maxCount = item.Count
			}
		}

		barWidth := 20

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

			// Calculate bar length
			barLen := 0
			if maxCount > 0 {
				barLen = (item.Count * barWidth) / maxCount
			}
			if barLen > barWidth {
				barLen = barWidth
			}
			if barLen < 1 && item.Count > 0 {
				barLen = 1 // At least 1 block for visibility
			}

			// Build the bar with fixed width
			bar := ""
			for i := 0; i < barLen; i++ {
				bar += "█"
			}
			for i := barLen; i < barWidth; i++ {
				bar += "░"
			}

			// Calculate percentage
			pctVal := float64(0)
			if totalReqs > 0 {
				pctVal = float64(item.Count) / float64(totalReqs) * 100
			}

			// Pad label to fixed width for alignment
			paddedLabel := label
			for len(paddedLabel) < 16 {
				paddedLabel += " "
			}

			// Render line with fixed widths
			line := fmt.Sprintf("  %s %s %6d %s",
				lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(paddedLabel),
				barStyle.Render(bar),
				item.Count,
				lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(fmt.Sprintf("(%5.1f%%)", pctVal)))

			s.WriteString(line + "\n")
		}

	} else {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true).Render("  Waiting for responses...") + "\n")
	}

	return s.String()
}
