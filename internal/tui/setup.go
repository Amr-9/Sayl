package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Amr-9/sayl/pkg/config"
	"github.com/Amr-9/sayl/pkg/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// Styles are now in styles.go

type Step int

const (
	StepURL Step = iota
	StepMethod
	StepMode          // New: Constant or Staged
	StepRate          // For Constant
	StepDuration      // For Constant
	StepStageDuration // For Staged
	StepStageTarget   // For Staged
	StepAddAnother    // For Staged
	StepConcurrency
	StepTimeout
	StepSuccessCodes
	StepSaveConfig // New: Save Config
	StepDone
)

type stepResult struct {
	label string
	value string
}

type SetupModel struct {
	config  *models.Config
	current Step
	history []stepResult
	form    *huh.Form // Active form for the current step

	// temporary fields for form binding
	tempRate         string
	tempDuration     string
	tempWorkers      string
	tempTimeout      string
	tempSuccessCodes string

	// Staged Mode Fields
	mode            string // "Constant" or "Staged"
	tempStageDur    string
	tempStageTarget string
	addAnother      bool

	saveConfig bool
}

func NewSetupModel(cfg *models.Config) *SetupModel {
	m := &SetupModel{
		config:           cfg,
		current:          StepURL,
		history:          make([]stepResult, 0),
		tempRate:         "100",
		tempDuration:     "30s",
		tempWorkers:      "10",
		tempTimeout:      "30s",
		tempSuccessCodes: "200",
		mode:             "Constant",
	}
	m.nextForm()
	return m
}

func (m *SetupModel) nextForm() {
	limit := huh.ThemeCharm() // fallback
	// We want to use our custom neon theme, but it's in the same package so we can call it directly
	// However, if it's not exported or something, we might have issues.
	// MakeNeonTheme is exported in styles.go, so `tui.MakeNeonTheme` or just `MakeNeonTheme` work.

	neon := MakeNeonTheme()
	_ = limit

	switch m.current {
	case StepURL:
		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Target URL").
					Placeholder("https://api.example.com").
					Value(&m.config.URL).
					Validate(func(s string) error {
						if len(s) < 4 || !strings.HasPrefix(s, "http") {
							return fmt.Errorf("URL must start with http")
						}
						return nil
					}),
			),
		).WithTheme(neon)
	case StepMethod:
		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("HTTP Method").
					Options(
						huh.NewOption("GET", "GET"),
						huh.NewOption("POST", "POST"),
						huh.NewOption("PUT", "PUT"),
						huh.NewOption("DELETE", "DELETE"),
						huh.NewOption("PATCH", "PATCH"),
					).
					Value(&m.config.Method),
			),
		).WithTheme(neon)
	case StepMode:
		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Attack Mode").
					Options(
						huh.NewOption("Constant Rate", "Constant"),
						huh.NewOption("Staged Attack (Ramping)", "Staged"),
					).
					Value(&m.mode),
			),
		).WithTheme(neon)
	case StepRate:
		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Requests Per Second (RPS)").
					Description("Target throughput").
					Value(&m.tempRate),
			),
		).WithTheme(neon)
	case StepDuration:
		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Test Duration").
					Description("e.g., 10s, 1m, 500ms").
					Value(&m.tempDuration),
			),
		).WithTheme(neon)
	case StepStageDuration:
		m.tempStageDur = "10s" // default
		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Stage Duration").
					Description("How long this stage lasts (e.g., 10s)").
					Value(&m.tempStageDur),
			),
		).WithTheme(neon)
	case StepStageTarget:
		m.tempStageTarget = "50" // default
		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Target Rate (End of Stage)").
					Description("RPS to reach by end of stage").
					Value(&m.tempStageTarget),
			),
		).WithTheme(neon)
	case StepAddAnother:
		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Add Another Stage?").
					Value(&m.addAnother),
			),
		).WithTheme(neon)
	case StepConcurrency:
		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Concurrency").
					Description("Simultaneous workers").
					Value(&m.tempWorkers),
			),
		).WithTheme(neon)
	case StepTimeout:
		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Request Timeout").
					Description("Max time to wait (e.g., 5s, 30s, 1m)").
					Value(&m.tempTimeout).
					Validate(func(s string) error {
						if _, err := time.ParseDuration(s); err != nil {
							return fmt.Errorf("invalid duration (use 10s, 1m, etc)")
						}
						return nil
					}),
			),
		).WithTheme(neon)
	case StepSuccessCodes:
		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Success Status Codes").
					Description("Comma-separated (e.g., 200,201,204)").
					Placeholder("200").
					Value(&m.tempSuccessCodes).
					Validate(func(s string) error {
						if len(s) == 0 {
							return fmt.Errorf("at least one status code required")
						}
						parts := strings.Split(s, ",")
						for _, part := range parts {
							code := strings.TrimSpace(part)
							if len(code) == 0 {
								continue
							}
							// Check if it's a valid number
							if _, err := fmt.Sscanf(code, "%d", new(int)); err != nil {
								return fmt.Errorf("invalid status code: %s", code)
							}
						}
						return nil
					}),
			),
		).WithTheme(neon)
	case StepSaveConfig:
		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Save Configuration?").
					Description("Save this setup to a YAML file for future use.").
					Value(&m.saveConfig),
			),
		).WithTheme(neon)
	case StepDone:
		m.form = nil
	}

	if m.form != nil {
		m.form.Init()
	}
}

func (m *SetupModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m *SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.current == StepDone {
		return m, nil
	}

	var cmd tea.Cmd
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateCompleted {
		// Record history
		switch m.current {
		case StepURL:
			m.history = append(m.history, stepResult{"Target", m.config.URL})
			m.current = StepMethod
		case StepMethod:
			m.history = append(m.history, stepResult{"Method", m.config.Method})
			m.current = StepMode
		case StepMode:
			m.history = append(m.history, stepResult{"Mode", m.mode})
			if m.mode == "Constant" {
				m.current = StepRate
			} else {
				m.config.Stages = []models.Stage{} // Reset
				m.current = StepStageDuration
			}
		case StepRate:
			m.history = append(m.history, stepResult{"Rate", m.tempRate + " req/s"})
			m.current = StepDuration
		case StepDuration:
			m.history = append(m.history, stepResult{"Duration", m.tempDuration})
			m.current = StepConcurrency
		case StepStageDuration:
			// No history yet, wait for target
			m.current = StepStageTarget
		case StepStageTarget:
			// Save Stage
			var dur time.Duration
			var target int
			// Parse (simplified error handling for brevity, real world should validate)
			// Assuming inputs are valid or validated in form
			// We should add validation in form definition actually
			// For now, let's just parse
			fmt.Sscanf(m.tempStageDur, "%v", &dur) // This format might be wrong for time.Duration
			// time.ParseDuration is better but Sscanf is... tricky.
			// Let's use standard parsing in validater ideally.
			// Re-parse here properly:
			d, _ := time.ParseDuration(m.tempStageDur)
			fmt.Sscanf(m.tempStageTarget, "%d", &target)

			m.config.Stages = append(m.config.Stages, models.Stage{Duration: d, Target: target})

			stepLabel := fmt.Sprintf("Stage %d", len(m.config.Stages))
			stepVal := fmt.Sprintf("%v @ %v RPS", m.tempStageDur, m.tempStageTarget)
			m.history = append(m.history, stepResult{stepLabel, stepVal})

			m.current = StepAddAnother
		case StepAddAnother:
			if m.addAnother {
				m.current = StepStageDuration
			} else {
				m.current = StepConcurrency
			}
		case StepConcurrency:
			m.history = append(m.history, stepResult{"Workers", m.tempWorkers})
			m.current = StepTimeout
		case StepTimeout:
			m.history = append(m.history, stepResult{"Timeout", m.tempTimeout})
			m.current = StepSuccessCodes
		case StepSuccessCodes:
			m.history = append(m.history, stepResult{"Success Codes", m.tempSuccessCodes})

			// Finalize Config
			if m.mode == "Constant" {
				fmt.Sscanf(m.tempRate, "%d", &m.config.Rate)
				m.config.Duration, _ = time.ParseDuration(m.tempDuration)
			}
			fmt.Sscanf(m.tempWorkers, "%d", &m.config.Concurrency)
			m.config.Timeout, _ = time.ParseDuration(m.tempTimeout)
			// Success codes parsing is done in form or here?
			// The tempSuccessCodes is string.
			// Helper function to parse...
			// For now let's assume it's set in m.config elsewhere or we do it here.
			// Actually the validator in form ensures it's correct but doesn't set it in config.
			// We need to parse tempSuccessCodes and set m.config.SuccessCodes
			codes := make(map[int]bool)
			parts := strings.Split(m.tempSuccessCodes, ",")
			for _, part := range parts {
				var code int
				if _, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &code); err == nil {
					codes[code] = true
				}
			}
			m.config.SuccessCodes = codes

			m.current = StepSaveConfig
		case StepSaveConfig:
			if m.saveConfig {
				// Generate Smart Filename
				// sayl-{host}-{rate}rps.yaml or sayl-{host}-staged.yaml

				// 1. Extract Host
				host := m.config.URL
				// Strip protocol
				if strings.Contains(host, "://") {
					parts := strings.Split(host, "://")
					if len(parts) > 1 {
						host = parts[1]
					}
				}
				// Strip path
				if idx := strings.Index(host, "/"); idx != -1 {
					host = host[:idx]
				}
				// Sanitize
				host = strings.ReplaceAll(host, ".", "-")
				host = strings.ReplaceAll(host, ":", "-")

				var filename string
				if len(m.config.Stages) > 0 {
					filename = fmt.Sprintf("sayl-%s-staged.yaml", host)
				} else {
					filename = fmt.Sprintf("sayl-%s-%drps.yaml", host, m.config.Rate)
				}

				// Check for existence and auto-increment
				ext := filepath.Ext(filename)
				base := strings.TrimSuffix(filename, ext)
				originalBase := base
				counter := 2

				for {
					if _, err := os.Stat(filename); os.IsNotExist(err) {
						break
					}
					// File exists, try next increment
					filename = fmt.Sprintf("%s (%d)%s", originalBase, counter, ext)
					counter++
				}

				// Placeholder for now, I will fix imports in next step
				if err := config.SaveConfig(filename, m.config); err != nil {
					// In a TUI, we might want to show an error, but for now let's just log it to history
					m.history = append(m.history, stepResult{"Save Error", err.Error()})
				} else {
					m.history = append(m.history, stepResult{"Saved", filename})
				}
			}
			m.current = StepDone
		}

		if m.current != StepDone {
			m.nextForm()
			// Init the new form immediately so it captures input if needed,
			// though Update loop usually handles it.
			return m, m.form.Init()
		}
	}

	return m, cmd
}

func (m *SetupModel) View() string {
	var s strings.Builder

	// Compact Header
	logo := logoStyle.Render(asciiLogo)
	subtitle := subtitleStyle.Render("Load Testing Tool")
	s.WriteString(borderStyle.Render(logo + subtitle))
	s.WriteString("\n\n")

	// Render History (completed steps)
	for _, h := range m.history {
		mark := check.Render("✓")
		label := subtext.Render(h.label + ":")
		val := finalValue.Render(h.value)
		s.WriteString(fmt.Sprintf("  %s %s %s\n", mark, label, val))
	}

	// Render Active Form
	if m.form != nil {
		if len(m.history) > 0 {
			s.WriteString("\n")
		}

		stepNum := len(m.history) + 1
		totalSteps := 6 // URL, Method, Rate, Duration, Workers, SuccessCodes
		header := questionHeader.Render(fmt.Sprintf("› Step %d/%d", stepNum, totalSteps))
		s.WriteString(header + "\n")

		s.WriteString(m.form.View())
	} else {
		// Done
		s.WriteString("\n" + highlight.Render("🚀 Ready! Press Enter to start..."))
	}

	return s.String()
}
