package models

import (
	"regexp"
	"time"
)

// AssertionType defines the type of assertion to perform
type AssertionType string

const (
	AssertContains AssertionType = "contains"
	AssertRegex    AssertionType = "regex"
	AssertJSONPath AssertionType = "json_path"
)

// Assertion defines a validation rule for response bodies
type Assertion struct {
	Type    AssertionType  `json:"type"`              // contains, regex, json_path
	Value   string         `json:"value"`             // Expected value or pattern
	Path    string         `json:"path,omitempty"`    // JSON path (for json_path type)
	Regex   *regexp.Regexp `json:"-"`                 // Pre-compiled regex (set at config load)
	Message string         `json:"message,omitempty"` // Custom error message
}

// CircuitBreaker defines conditions to stop a test automatically
type CircuitBreaker struct {
	// StopIf is the raw condition string, e.g., "errors > 10%"
	StopIf string `json:"stop_if"`
	// MinSamples is the minimum number of requests before the breaker can trip (cold start protection)
	MinSamples int64 `json:"min_samples"`
	// Parsed condition fields (set during config load)
	Metric    string  `json:"-"` // "errors", "error_rate", "failures"
	Operator  string  `json:"-"` // ">", "<", ">=", "<="
	Threshold float64 `json:"-"` // Threshold value (e.g., 10 for 10%, or 0.1 for rate)
	IsPercent bool    `json:"-"` // Whether threshold is a percentage
}

// Config defines the load test parameters
type Config struct {
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Body           []byte            `json:"body,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	Timeout        time.Duration     `json:"timeout"`
	Insecure       bool              `json:"insecure"`   // Skip TLS verification
	KeepAlive      bool              `json:"keep_alive"` // Use keep-alive connections
	Duration       time.Duration     `json:"duration"`
	Rate           int               `json:"rate"`        // Requests per second
	Concurrency    int               `json:"concurrency"` // Number of workers
	SuccessCodes   map[int]bool      `json:"success_codes"`
	Stages         []Stage           `json:"stages,omitempty"`
	Steps          []Step            `json:"steps,omitempty"` // For chained scenarios
	Data           []DataSource      `json:"data,omitempty"`  // distinct CSV data sources
	CircuitBreaker *CircuitBreaker   `json:"circuit_breaker,omitempty"`
	Debug          bool              `json:"-"` // Debug mode - run single iteration with detailed output
}

// DataSource defines a source of external data (e.g. CSV file)
type DataSource struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Step represents a single step in a chained scenario
type Step struct {
	Name       string            `json:"name"`
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`      // Raw string for templating
	Extract    map[string]string `json:"extract,omitempty"`   // Extraction rules: "var_name": "json_path"
	Variables  map[string]string `json:"variables,omitempty"` // Variables to pre-calculate and store in session
	Assertions []Assertion       `json:"assertions,omitempty"`
}

// Stage represents a load test stage
type Stage struct {
	Duration time.Duration `json:"duration"`
	Target   int           `json:"target"` // Target requests per second
}

// Result represents a single HTTP request outcome
type Result struct {
	Timestamp      time.Time
	Latency        time.Duration
	Status         int
	Bytes          int64
	Error          error  // Network/server error
	AssertionError error  // Assertion failure (classified separately)
	StepName       string // Name of the step for reporting
}

// SecondStats captures metrics for a single second of the test
type SecondStats struct {
	Second            int            `json:"second"`
	Requests          int64          `json:"requests"`
	Success           int64          `json:"success"`
	Failures          int64          `json:"failures"`
	AssertionFailures int64          `json:"assertion_failures"`
	AvgLatency        float64        `json:"avg_latency_ms"`
	P50               time.Duration  `json:"p50"`
	P75               time.Duration  `json:"p75"`
	P90               time.Duration  `json:"p90"`
	P95               time.Duration  `json:"p95"`
	P99               time.Duration  `json:"p99"`
	StatusCodes       map[string]int `json:"status_codes"`
}

// Report is the final summary of the load test
type Report struct {
	TargetURL          string         `json:"target_url"`
	Method             string         `json:"method"`
	Duration           time.Duration  `json:"duration"` // Configured duration
	Concurrency        int            `json:"concurrency"`
	TotalRequests      int64          `json:"total_requests"`
	SuccessCount       int64          `json:"success_count"`
	FailureCount       int64          `json:"failure_count"`
	AssertionFailures  int64          `json:"assertion_failures"` // Separate from network failures
	SuccessRate        float64        `json:"success_rate"`
	TotalBytes         int64          `json:"total_bytes"`
	Throughput         float64        `json:"throughput"` // MB/s
	RPS                float64        `json:"rps"`
	P50                time.Duration  `json:"p50"`
	P75                time.Duration  `json:"p75"`
	P90                time.Duration  `json:"p90"`
	P95                time.Duration  `json:"p95"`
	P99                time.Duration  `json:"p99"`
	Max                time.Duration  `json:"max"`
	Min                time.Duration  `json:"min"`
	StatusCodes        map[string]int `json:"status_codes"`
	Errors             map[string]int `json:"errors"`
	AssertionErrors    map[string]int `json:"assertion_errors,omitempty"` // Assertion failures by message
	TimeSeriesData     []SecondStats  `json:"time_series_data"`
	CircuitBroken      bool           `json:"circuit_broken,omitempty"`
	CircuitBreakReason string         `json:"circuit_break_reason,omitempty"`
}
