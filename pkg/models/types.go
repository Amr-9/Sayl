package models

import "time"

// Config defines the load test parameters
type Config struct {
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	Body         []byte            `json:"body,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	Timeout      time.Duration     `json:"timeout"`
	Insecure     bool              `json:"insecure"`   // Skip TLS verification
	KeepAlive    bool              `json:"keep_alive"` // Use keep-alive connections
	Duration     time.Duration     `json:"duration"`
	Rate         int               `json:"rate"`        // Requests per second
	Concurrency  int               `json:"concurrency"` // Number of workers
	SuccessCodes map[int]bool      `json:"success_codes"`
	Stages       []Stage           `json:"stages,omitempty"`
	Steps        []Step            `json:"steps,omitempty"` // For chained scenarios
	Data         []DataSource      `json:"data,omitempty"`  // distinct CSV data sources
}

// DataSource defines a source of external data (e.g. CSV file)
type DataSource struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Step represents a single step in a chained scenario
type Step struct {
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers,omitempty"`
	Body      string            `json:"body,omitempty"`      // Raw string for templating
	Extract   map[string]string `json:"extract,omitempty"`   // Extraction rules: "var_name": "json_path"
	Variables map[string]string `json:"variables,omitempty"` // Variables to pre-calculate and store in session
}

// Stage represents a load test stage
type Stage struct {
	Duration time.Duration `json:"duration"`
	Target   int           `json:"target"` // Target requests per second
}

// Result represents a single HTTP request outcome
type Result struct {
	Timestamp time.Time
	Latency   time.Duration
	Status    int
	Bytes     int64
	Error     error
}

// SecondStats captures metrics for a single second of the test
type SecondStats struct {
	Second      int            `json:"second"`
	Requests    int64          `json:"requests"`
	Success     int64          `json:"success"`
	Failures    int64          `json:"failures"`
	AvgLatency  float64        `json:"avg_latency_ms"`
	P50         time.Duration  `json:"p50"`
	P75         time.Duration  `json:"p75"`
	P90         time.Duration  `json:"p90"`
	P95         time.Duration  `json:"p95"`
	P99         time.Duration  `json:"p99"`
	StatusCodes map[string]int `json:"status_codes"`
}

// Report is the final summary of the load test
type Report struct {
	TargetURL      string         `json:"target_url"`
	Method         string         `json:"method"`
	Duration       time.Duration  `json:"duration"` // Configured duration
	Concurrency    int            `json:"concurrency"`
	TotalRequests  int64          `json:"total_requests"`
	SuccessCount   int64          `json:"success_count"`
	FailureCount   int64          `json:"failure_count"`
	SuccessRate    float64        `json:"success_rate"`
	TotalBytes     int64          `json:"total_bytes"`
	Throughput     float64        `json:"throughput"` // MB/s
	RPS            float64        `json:"rps"`
	P50            time.Duration  `json:"p50"`
	P75            time.Duration  `json:"p75"`
	P90            time.Duration  `json:"p90"`
	P95            time.Duration  `json:"p95"`
	P99            time.Duration  `json:"p99"`
	Max            time.Duration  `json:"max"`
	Min            time.Duration  `json:"min"`
	StatusCodes    map[string]int `json:"status_codes"`
	Errors         map[string]int `json:"errors"`
	TimeSeriesData []SecondStats  `json:"time_series_data"`
}
