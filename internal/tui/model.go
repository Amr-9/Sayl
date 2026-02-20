package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Amr-9/sayl/internal/attacker"
	"github.com/Amr-9/sayl/internal/stats"
	"github.com/Amr-9/sayl/pkg/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type State int

const (
	StateSetup State = iota
	StateRunning
	StateSummary
)

type MainModel struct {
	state    State
	config   models.Config
	report   models.Report
	results  chan models.Result
	drainDone chan struct{} // closed by processResults when the channel is fully drained
	quitting bool

	// Phases
	setupModel tea.Model
	dashModel  tea.Model
	sumModel   tea.Model

	monitor *stats.Monitor
}

func NewModel(cfg *models.Config, startRunning bool) MainModel {
	if cfg == nil {
		cfg = &models.Config{
			Method: "GET",
			// Defaults
			Rate:         100,
			Duration:     10 * time.Second,
			Concurrency:  10,
			SuccessCodes: map[int]bool{200: true},
			HTTP2:        true,
			KeepAlive:    true,
		}
	} else if len(cfg.SuccessCodes) == 0 {
		cfg.SuccessCodes = map[int]bool{200: true}
	}

	initialState := StateSetup
	if startRunning {
		initialState = StateRunning
	}

	m := MainModel{
		state:      initialState,
		config:     *cfg,
		setupModel: NewSetupModel(cfg),
	}

	if startRunning {
		// If starting immediately, skip setup and initialize stats/dashboard

		// Recalculate duration from stages if not explicitly set
		if m.config.Duration == 0 && len(m.config.Stages) > 0 {
			for _, s := range m.config.Stages {
				m.config.Duration += s.Duration
			}
		}

		m.results = make(chan models.Result, 10000)
		m.drainDone = make(chan struct{})
		m.monitor = stats.NewMonitor()
		// History can be empty or populated from config if we want
		m.dashModel = NewDashModel(m.config, []string{"Loaded from config/flags"})
	}

	return m
}

func (m MainModel) Init() tea.Cmd {
	if m.state == StateRunning {
		return tea.Batch(
			m.startAttacking(),
			m.processResults(),
			m.tick(),
		)
	}
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	switch m.state {
	case StateSetup:
		m.setupModel, cmd = m.setupModel.Update(msg)
		if sm, ok := m.setupModel.(*SetupModel); ok {
			if sm.current == StepDone {
				// Parse strings from form temp vars
				// Note: URL and Method are already directly bound

				rateVal, _ := strconv.Atoi(sm.tempRate)
				m.config.Rate = rateVal

				dur, _ := time.ParseDuration(sm.tempDuration)
				if dur == 0 {
					dur = 10 * time.Second
				} // fallback
				m.config.Duration = dur

				workers, _ := strconv.Atoi(sm.tempWorkers)
				if workers == 0 {
					workers = 1
				}
				m.config.Concurrency = workers

				m.config.Concurrency = workers

				// Parse success codes from user input
				successCodes := make(map[int]bool)
				if sm.tempSuccessCodes != "" {
					parts := strings.Split(sm.tempSuccessCodes, ",")
					for _, part := range parts {
						code := strings.TrimSpace(part)
						if len(code) == 0 {
							continue
						}
						var statusCode int
						if _, err := fmt.Sscanf(code, "%d", &statusCode); err == nil {
							successCodes[statusCode] = true
						}
					}
				}
				// Fallback to default if parsing failed
				if len(successCodes) == 0 {
					successCodes[200] = true
				}
				m.config.SuccessCodes = successCodes

				// CRITICAL: Copy URL and Method from the setup model's config pointer
				// because they represent the latest state from the form.
				if sm.config != nil {
					m.config.URL = sm.config.URL
					m.config.Method = sm.config.Method
				}

				// Generate history lines for dashboard
				var history []string
				for i, h := range sm.history {
					// Re-create the visual style (checkmark, label, value)
					// We need to access styles from styles.go (same package)
					line := fmt.Sprintf("%s %s %s",
						lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
						lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(fmt.Sprintf("[%d] %s", i+1, h.label)),
						lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render(h.value),
					)
					history = append(history, line)
				}

				m.state = StateRunning
				m.results = make(chan models.Result, 10000)
				m.drainDone = make(chan struct{})
				m.monitor = stats.NewMonitor()
				m.dashModel = NewDashModel(m.config, history)

				return m, tea.Batch(
					m.startAttacking(),
					m.processResults(),
					m.tick(),
				)
			}
		}
	case StateRunning:
		m.dashModel, cmd = m.dashModel.Update(msg)
		switch msg.(type) {
		case tickMsg:
			report := m.monitor.Snapshot()
			report.TargetURL = m.config.URL
			report.Method = m.config.Method
			report.Duration = m.config.Duration
			report.Concurrency = m.config.Concurrency
			m.report = report
			// Explicitly update dashboard with proper stats
			m.dashModel, _ = m.dashModel.Update(report)

			return m, m.tick()
		case finishedMsg:
			m.state = StateSummary
			m.state = StateSummary
			m.report = m.monitor.Snapshot()
			m.report.TargetURL = m.config.URL
			m.report.Method = m.config.Method
			m.report.Duration = m.config.Duration
			m.report.Concurrency = m.config.Concurrency
			m.sumModel = NewSummaryModel(m.report)
		}
	}

	return m, cmd
}

type finishedMsg struct{}

type tickMsg time.Time

func (m MainModel) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m MainModel) startAttacking() tea.Cmd {
	return func() tea.Msg {
		engine := attacker.NewEngine()
		ctx, cancel := context.WithTimeout(context.Background(), m.config.Duration)
		defer cancel()

		engine.Attack(ctx, m.config, m.results)
		// Attack() has returned and closed m.results. Wait for processResults to
		// drain every buffered result before we signal the UI to switch to summary.
		<-m.drainDone
		return finishedMsg{}
	}
}

func (m MainModel) processResults() tea.Cmd {
	return func() tea.Msg {
		// Defer ensures drainDone is always closed even if a panic occurs.
		defer close(m.drainDone)
		for res := range m.results {
			isSuccess := m.config.SuccessCodes[res.Status] && res.Error == nil
			m.monitor.Add(res, isSuccess)
		}
		return nil
	}
}

func (m MainModel) View() string {
	if m.quitting {
		return "Exiting...\n"
	}

	switch m.state {
	case StateSetup:
		return m.setupModel.View()
	case StateRunning:
		return m.dashModel.View()
	case StateSummary:
		return m.sumModel.View()
	default:
		return "Unknown state"
	}
}

func (m MainModel) Report() models.Report {
	return m.report
}
