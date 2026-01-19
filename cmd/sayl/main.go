package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Amr-9/sayl/internal/report"
	"github.com/Amr-9/sayl/internal/tui"
	"github.com/Amr-9/sayl/pkg/config"
	"github.com/Amr-9/sayl/pkg/models"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Use all available CPU cores for maximum performance
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Define command-line flags
	var (
		configPath  string
		url         string
		method      string
		rate        int
		durationStr string
		concurrency int
		successStr  string
	)

	flag.StringVar(&configPath, "config", "", "Path to YAML configuration file")
	flag.StringVar(&configPath, "f", "", "Path to YAML configuration file (shorthand)")
	flag.StringVar(&url, "url", "", "Target URL")
	flag.StringVar(&method, "method", "", "HTTP Method (GET, POST, etc.)")
	flag.IntVar(&rate, "rate", 0, "Requests per second")
	flag.StringVar(&durationStr, "duration", "", "Duration of the test (e.g., 10s, 1m)")
	flag.IntVar(&concurrency, "concurrency", 0, "Number of concurrent workers")
	flag.StringVar(&successStr, "success", "", "Comma-separated list of success status codes (e.g., 200,201)")

	flag.Parse()

	var cfg *models.Config

	// 1. Load from Config File if provided
	if configPath != "" {
		loadedCfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Error loading config file: %v\n", err)
			os.Exit(1)
		}
		cfg = loadedCfg
	} else {
		// Initialize empty config if no file
		cfg = &models.Config{}
	}

	// 2. Override with Flags (Precedence: Flag > File)
	if url != "" {
		cfg.URL = url
	}
	if method != "" {
		cfg.Method = method
	}
	if rate > 0 {
		cfg.Rate = rate
	}
	if durationStr != "" {
		d, err := time.ParseDuration(durationStr)
		if err != nil {
			fmt.Printf("Invalid duration flag: %v\n", err)
			os.Exit(1)
		}
		cfg.Duration = d
	}
	if concurrency > 0 {
		cfg.Concurrency = concurrency
	}
	if successStr != "" {
		codes := make(map[int]bool)
		parts := strings.Split(successStr, ",")
		for _, part := range parts {
			var code int
			if _, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &code); err == nil {
				codes[code] = true
			}
		}
		if len(codes) > 0 {
			cfg.SuccessCodes = codes
		}
	}

	// 3. Defaults are handled inside config.Validate or TUI Setup
	// Check if we have enough info to run immediately (Skip Setup)
	startRunning := false
	if err := config.Validate(cfg); err == nil {
		startRunning = true
	} else {
		// If validation fails but user provided some flags, might want to warn
		// But usually we just fall back to TUI setup with prepopulated values
		// However, TUI SetupModel uses *models.Config so we can pass what we have
	}

	p := tea.NewProgram(tui.NewModel(cfg, startRunning))
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	if finalModel, ok := m.(tui.MainModel); ok {
		// Only save if we actually ran a test
		if finalModel.Report().TotalRequests > 0 {
			rep := finalModel.Report()

			// Save JSON report
			saveReport("report.json", rep)
			fmt.Println("\n📊 Report saved to report.json")

			// Generate HTML report with charts
			if err := report.GenerateHTML(rep, "report.html"); err != nil {
				fmt.Printf("⚠️  Failed to generate HTML report: %v\n", err)
			} else {
				fmt.Println("📈 Interactive HTML report saved to report.html")
			}
		}
	}
}

func saveReport(path string, report models.Report) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
