package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/Amr-9/sayl/internal/debug"
	"github.com/Amr-9/sayl/internal/report"
	"github.com/Amr-9/sayl/internal/tui"
	"github.com/Amr-9/sayl/pkg/config"
	"github.com/Amr-9/sayl/pkg/models"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Panic recovery - prevent crashes
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("\n❌ Fatal error: %v\n", r)
			fmt.Println("💡 Please report this issue at: https://github.com/Amr-9/sayl/issues")
			os.Exit(1)
		}
	}()

	// Use all available CPU cores for maximum performance
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Setup graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals (Ctrl+C, SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\n⚠️  Received interrupt signal, shutting down gracefully...")
		cancel()
		// Give a moment for cleanup
		time.Sleep(500 * time.Millisecond)
	}()

	// Store context for TUI to use
	_ = ctx // Will be passed to TUI in future enhancement

	// Define command-line flags
	var (
		configPath  string
		url         string
		method      string
		rate        int
		durationStr string
		concurrency int
		successStr  string
		debugMode   bool
	)

	flag.StringVar(&configPath, "config", "", "Path to YAML configuration file")
	flag.StringVar(&configPath, "f", "", "Path to YAML configuration file (shorthand)")
	flag.StringVar(&url, "url", "", "Target URL")
	flag.StringVar(&method, "method", "", "HTTP Method (GET, POST, etc.)")
	flag.IntVar(&rate, "rate", 0, "Requests per second")
	flag.StringVar(&durationStr, "duration", "", "Duration of the test (e.g., 10s, 1m)")
	flag.IntVar(&concurrency, "concurrency", 0, "Number of concurrent workers")
	flag.StringVar(&successStr, "success", "", "Comma-separated list of success status codes (e.g., 200,201)")
	flag.BoolVar(&debugMode, "debug", false, "Run in debug mode (single iteration with detailed output)")
	flag.BoolVar(&debugMode, "d", false, "Run in debug mode (shorthand)")

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
		// If a config file was explicitly provided but is invalid, we should report the error and exit
		// instead of dropping into the TUI.
		if configPath != "" {
			fmt.Printf("Configuration Error: %v\n", err)
			os.Exit(1)
		}
		// Otherwise (no config file), fall back to TUI setup
	}

	// 4. Debug Mode - Run single iteration with detailed output (bypasses TUI)
	if debugMode {
		if !startRunning {
			fmt.Println("❌ Debug mode requires a valid configuration.")
			fmt.Println("💡 Please provide a config file: sayl -config scenario.yaml --debug")
			os.Exit(1)
		}

		// Set debug flag on config
		cfg.Debug = true

		if err := debug.RunDebugMode(cfg); err != nil {
			fmt.Printf("❌ Debug mode error: %v\n", err)
			os.Exit(1)
		}
		return // Exit after debug mode completes
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

			// Print console summary
			report.PrintConsoleReport(rep)

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

func saveReport(path string, rep models.Report) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create report file '%s': %w", path, err)
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(rep); err != nil {
		f.Close()
		return fmt.Errorf("failed to encode report: %w", err)
	}

	// Sync to ensure data is written to disk
	if err := f.Sync(); err != nil {
		f.Close()
		return fmt.Errorf("failed to sync report file: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close report file: %w", err)
	}

	return nil
}
