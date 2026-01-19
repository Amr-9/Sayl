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

	// Grid Layout

	// Box 1: Traffic
	// Box 1: Traffic + Sparkline
	rps := fmt.Sprintf("%.1f", m.report.RPS)

	// Dynamically format throughput
	duration := time.Since(m.start).Seconds()
	// Use Report.Throughput (MB/s) if available, or calculate for dynamic unit
	// Since report.Throughput is fixed to MB/s in stats, let's recalculate for display flexibility
	tput := formatThroughput(m.report.TotalBytes, duration)

	totalData := formatBytes(m.report.TotalBytes)

	// Create sparkline from TimeSeriesData (RPS)
	var rpsHistory []int
	// Limit history to last 20 seconds for cleanliness
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

	// Box 2: Latency
	p50 := valStyle.Render(fmtDuration(m.report.P50))
	p90 := valStyle.Render(fmtDuration(m.report.P90))
	p99 := valStyle.Render(fmtDuration(m.report.P99))
	max := valStyle.Render(fmtDuration(m.report.Max))
	box2 := boxStyle.Render(fmt.Sprintf(
		"P50: %s  P90: %s\nP99: %s  Max: %s",
		p50, p90, p99, max,
	))

	// Box 3: Status Summary
	succ := successStyle.Render(fmt.Sprintf("%d", m.report.SuccessCount))
	fail := failStyle.Render(fmt.Sprintf("%d", m.report.FailureCount))
	rate := valStyle.Render(fmt.Sprintf("%.1f%%", m.report.SuccessRate))
	box3 := boxStyle.Render(fmt.Sprintf(
		"Success:  %s  (%s)\nFail:     %s",
		succ, rate,
		fail,
	))

	// Box 4: Status Codes Details
	var codesList strings.Builder
	if len(m.report.StatusCodes) > 0 {
		type kv struct {
			Code  string
			Count int
			Label string // New field for category label
		}
		var sorted []kv
		for k, v := range m.report.StatusCodes {
			sorted = append(sorted, kv{Code: k, Count: v, Label: ""})
		}
		// Sort by code string
		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i].Code > sorted[j].Code {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		// Prepare items for display
		var items []kv
		items = append(items, sorted...)

		// Classify errors (Status 0 / "Timeout" etc)
		// We already have "Timeout" in StatusCodes now, so we might duplicate if we also process m.report.Errors map?
		// But m.report.Errors are the internal errors. "Timeout" in StatusCodes is the summarized status.
		// Let's rely on StatusCodes for the main list.

		for i, item := range items {
			cStyle := valStyle
			var label string

			label = item.Code
			// Try to detect if it's a numeric status code for coloring
			var codeInt int
			n, _ := fmt.Sscanf(item.Code, "%d", &codeInt)
			if n > 0 {
				// It's a number
				if codeInt >= 200 && codeInt < 300 {
					cStyle = successStyle
				} else if codeInt >= 400 {
					cStyle = failStyle
				}
			} else {
				// It's text like "Timeout"
				cStyle = failStyle
			}

			codesList.WriteString(fmt.Sprintf("%s: %s",
				labelStyle.Render(label),
				cStyle.Render(fmt.Sprintf("%d", item.Count))))

			if i < len(items)-1 {
				codesList.WriteString("\n")
			}
		}
	} else {
		codesList.WriteString(labelStyle.Render("Waiting..."))
	}

	box4 := boxStyle.Render(codesList.String())

	// Horizontal join (Top Row)
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, box1, box2, box3)

	s.WriteString(row1)
	s.WriteString("\n")

	// Row 2: Detailed Codes
	s.WriteString(lipgloss.NewStyle().MarginTop(1).Render("Status Codes:"))
	s.WriteString("\n")
	s.WriteString(box4)
	s.WriteString("\n")

	return s.String()
}
