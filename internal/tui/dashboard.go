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
}

func NewDashModel(cfg models.Config, history []string) *DashModel {
	return &DashModel{
		config:   cfg,
		report:   models.Report{},
		start:    time.Now(),
		progress: progress.New(progress.WithDefaultGradient()),
		history:  history,
	}
}

func (m *DashModel) Init() tea.Cmd {
	return nil
}

func (m *DashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case models.Report:
		m.report = msg
	}
	// no separate cmd for progress, it's just visual
	return m, nil
}

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			MarginBottom(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1).
			MarginRight(1)

	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	valStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true)

	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	failStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

func (m *DashModel) View() string {
	var s strings.Builder

	// Render History first (preserved from setup)
	if len(m.history) > 0 {
		for _, line := range m.history {
			s.WriteString(line + "\n")
		}
		s.WriteString("\n")
	}

	s.WriteString(headerStyle.Render(fmt.Sprintf("⚡ ATTACKING %s", m.config.URL)))
	s.WriteString("\n\n")

	// Calc progress
	elapsed := time.Since(m.start)
	pct := float64(elapsed) / float64(m.config.Duration)
	if pct > 1.0 {
		pct = 1.0
	}

	s.WriteString(m.progress.ViewAs(pct))
	s.WriteString(fmt.Sprintf("\n %s / %s\n\n", elapsed.Round(time.Second), m.config.Duration))

	// --- Visual Dashboard Grid ---

	// 1. Throughput & Data Box
	rps := fmt.Sprintf("%.1f", m.report.RPS)
	tput := formatThroughput(m.report.TotalBytes, elapsed.Seconds())
	totalData := formatBytes(m.report.TotalBytes)

	// Sparkline Logic
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

	box1 := boxStyle.Render(fmt.Sprintf(
		"RPS:      %s\nFlow:     %s\nData:     %s\n%s",
		valStyle.Render(rps),
		valStyle.Render(tput),
		valStyle.Render(totalData),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Render(spark),
	))

	// 2. Latency Box
	p50 := valStyle.Render(fmtDuration(m.report.P50))
	p90 := valStyle.Render(fmtDuration(m.report.P90))
	p99 := valStyle.Render(fmtDuration(m.report.P99))
	max := valStyle.Render(fmtDuration(m.report.Max))
	box2 := boxStyle.Render(fmt.Sprintf(
		"Latency (P50): %s\nLatency (P90): %s\nLatency (P99): %s\nMax Latency:   %s",
		p50, p90, p99, max,
	))

	// 3. Success/Fail Box (Enhanced)
	totalReqs := m.report.SuccessCount + m.report.FailureCount
	var failPct float64
	if totalReqs > 0 {
		failPct = (float64(m.report.FailureCount) / float64(totalReqs)) * 100.0
	}

	// Dynamic coloring for failure rate
	failColor := successText
	if failPct > 0 {
		failColor = warnText
	}
	if failPct > 5.0 {
		failColor = errText
	}

	succ := successText.Render(fmt.Sprintf("%d", m.report.SuccessCount))
	fail := errText.Render(fmt.Sprintf("%d", m.report.FailureCount))
	// success rate (green)
	sRate := successText.Render(fmt.Sprintf("%.1f%%", m.report.SuccessRate))
	// failure rate (dynamic)
	fRate := failColor.Render(fmt.Sprintf("%.1f%%", failPct))

	box3 := boxStyle.Render(fmt.Sprintf(
		"Total Reqs: %d\nSuccess:    %s (%s)\nFailures:   %s (%s)",
		totalReqs, succ, sRate, fail, fRate,
	))

	// 4. Status Codes (Detailed & Friendly)
	var codesList strings.Builder
	if len(m.report.StatusCodes) > 0 {
		type kv struct {
			Code  string
			Count int
		}
		var sorted []kv
		for k, v := range m.report.StatusCodes {
			sorted = append(sorted, kv{Code: k, Count: v})
		}
		// Sort by code string
		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i].Code > sorted[j].Code {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		for i, item := range sorted {
			cStyle := valStyle
			label := item.Code

			// --- Status Code Re-labeling ---
			if label == "0" {
				label = "NetErr/Timeout"
				cStyle = errText
			} else {
				// Try parsing int
				var codeInt int
				n, _ := fmt.Sscanf(label, "%d", &codeInt)
				if n > 0 {
					if codeInt >= 200 && codeInt < 300 {
						cStyle = successText
						label = fmt.Sprintf("%s OK", item.Code)
					} else if codeInt >= 300 && codeInt < 400 {
						cStyle = warnText
						label = fmt.Sprintf("%s Redirect", item.Code)
					} else if codeInt >= 400 && codeInt < 500 {
						cStyle = warnText
						label = fmt.Sprintf("%s Client Err", item.Code)
					} else if codeInt >= 500 {
						cStyle = errText
						label = fmt.Sprintf("%s Server Err", item.Code)
					}
				} else {
					// Fallback for non-numeric (e.g. "Timeout")
					cStyle = errText
				}
			}

			// Format: "200 OK:        150"
			// Fixed width for alignment
			codesList.WriteString(fmt.Sprintf("%-16s %s",
				labelStyle.Render(label+":"),
				cStyle.Render(fmt.Sprintf("%d", item.Count))))

			if i < len(sorted)-1 {
				codesList.WriteString("\n")
			}
		}
	} else {
		codesList.WriteString(labelStyle.Render("Waiting for data..."))
	}

	box4 := boxStyle.Render(codesList.String())

	// Layout Composition
	// Row 1: Traffic | Latency | Success
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, box1, box2, box3)
	s.WriteString(row1)
	s.WriteString("\n")

	// Row 2: Status Codes details
	s.WriteString(lipgloss.NewStyle().MarginTop(1).Foreground(lipgloss.Color("241")).Render("Status Details:"))
	s.WriteString("\n")
	s.WriteString(box4)
	s.WriteString("\n")

	return s.String()
}
