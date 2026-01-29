package attacker

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Amr-9/sayl/internal/validator"
	"github.com/Amr-9/sayl/pkg/models"
	"github.com/tidwall/gjson"
	"golang.org/x/net/http2"
	"golang.org/x/time/rate"
)

// RetryConfig holds retry settings
type RetryConfig struct {
	MaxRetries int
	RetryDelay time.Duration
}

// Engine implements the load testing logic
type Engine struct {
	client      *http.Client
	vp          *VariableProcessor
	retry       RetryConfig
	sessionPool *sync.Pool
}

// DefaultRetryConfig returns reasonable defaults for retries
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		RetryDelay: 100 * time.Millisecond,
	}
}

func NewEngine() *Engine {
	return &Engine{
		vp:    NewVariableProcessor(),
		retry: DefaultRetryConfig(),
		sessionPool: &sync.Pool{
			New: func() any {
				return make(map[string]string)
			},
		},
	}
}

// PreflightCheck verifies that the target is reachable before starting the load test
func (e *Engine) PreflightCheck(url string, timeout time.Duration) error {
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		// Try with GET if HEAD fails to create
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
	}
	req.Header.Set("User-Agent", "Sayl/1.0 Preflight")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("target unreachable: %w", err)
	}
	defer resp.Body.Close()

	// Drain the body to allow connection reuse
	io.Copy(io.Discard, resp.Body)

	return nil
}

// isRetryableError checks if an error is worth retrying
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	retryablePatterns := []string{
		"timeout",
		"connection reset",
		"connection refused",
		"no such host",
		"EOF",
		"i/o timeout",
		"TLS handshake timeout",
	}
	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// Attack starts the load test
func (e *Engine) Attack(ctx context.Context, cfg models.Config, results chan<- models.Result) {
	// Configure client based on config
	maxConns := cfg.Concurrency * 2
	if maxConns < 100 {
		maxConns = 100
	}

	var roundTripper http.RoundTripper

	if cfg.H2C {
		// HTTP/2 Cleartext (h2c) - for non-TLS HTTP/2 testing
		roundTripper = &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				// For h2c, we dial plain TCP (no TLS)
				return (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext(ctx, network, addr)
			},
		}
	} else {
		// Standard transport - HTTP/2 enabled by default with automatic fallback to HTTP/1.1
		transport := &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: cfg.Insecure},
			MaxIdleConns:        maxConns,
			MaxIdleConnsPerHost: maxConns,
			MaxConnsPerHost:     maxConns,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   !cfg.KeepAlive,
			ForceAttemptHTTP2:   cfg.HTTP2, // Default: true
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		}

		// Always configure HTTP/2 support (with automatic fallback to HTTP/1.1)
		// This ensures proper HTTP/2 negotiation via ALPN
		if cfg.HTTP2 {
			_ = http2.ConfigureTransport(transport) // Ignore error - fallback to HTTP/1.1
		}

		if transport.ResponseHeaderTimeout == 0 && cfg.Timeout > 0 {
			transport.ResponseHeaderTimeout = cfg.Timeout
		}

		roundTripper = transport
	}

	e.client = &http.Client{
		Timeout:   cfg.Timeout,
		Transport: roundTripper,
	}
	if e.client.Timeout == 0 {
		e.client.Timeout = 30 * time.Second
	}

	// Initialize Data Feeders
	feeders := make(map[string]*CSVFeeder)
	for _, d := range cfg.Data {
		f, err := NewCSVFeeder(d.Path)
		if err != nil {
			// Report error and exit
			// Since we can't return an error here easily without changing signature,
			// we'll print to stderr and send an error result if possible or just return.
			// Best effort: Log and stop.
			// Ideally we should probably log this better.
			// For CLI tool, printing is okay.
			// We can also send a "system error" result.
			results <- models.Result{
				Timestamp: time.Now(),
				Error:     fmt.Errorf("feeder init error: %v", err),
			}
			close(results)
			return
		}
		feeders[d.Name] = f
	}

	var wg sync.WaitGroup

	// Rate Limiter Setup (simplified for now)
	var initialLimit rate.Limit
	if len(cfg.Stages) > 0 {
		initialLimit = rate.Limit(1) // Start slow if staging
	} else {
		initialLimit = rate.Limit(cfg.Rate)
	}
	limiter := rate.NewLimiter(initialLimit, 1)

	// Stage Controller
	if len(cfg.Stages) > 0 {
		go e.runStages(ctx, cfg.Stages, limiter)
	}

	// Prepare steps
	steps := cfg.Steps
	if len(steps) == 0 {
		// Create a single step from the main config
		steps = []models.Step{{
			Name:    "Main",
			URL:     cfg.URL,
			Method:  cfg.Method,
			Headers: cfg.Headers,
			Body:    string(cfg.Body),
		}}
	}

	// Launch workers
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				// Wait for rate limit permission
				if err := limiter.Wait(ctx); err != nil {
					return // Context cancelled
				}

				select {
				case <-ctx.Done():
					return
				default:
					// CRITICAL: Reuse session map to reduce GC pressure
					session := e.sessionPool.Get().(map[string]string)
					clear(session) // Go 1.21+ built-in to clear map efficiently

					// Feed Data
					for name, f := range feeders {
						data := f.Next()
						for k, v := range data {
							session[name+"."+k] = v
						}
					}

					// Execute scenario steps
					for _, step := range steps {
						result := e.executeStepWithRetry(ctx, step, session)

						// Send result
						select {
						case results <- result:
						case <-ctx.Done():
							// Ensure we return the map even if cancelled here
							e.sessionPool.Put(session)
							return
						}

						// If step failed (and not ignored?), break scenario
						// For now, any non-2xx/3xx or error stops the chain
						if result.Error != nil || result.Status >= 400 {
							break
						}
					}

					// Return map to pool for reuse
					e.sessionPool.Put(session)
				}
			}
		}()
	}

	wg.Wait()
	close(results)
}

func (e *Engine) runStages(ctx context.Context, stages []models.Stage, limiter *rate.Limiter) {
	for _, stage := range stages {
		startLimit := float64(limiter.Limit())
		targetLimit := float64(stage.Target)
		if targetLimit == 0 {
			targetLimit = 1
		}
		duration := stage.Duration
		ticker := time.NewTicker(100 * time.Millisecond)
		startTime := time.Now()

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case t := <-ticker.C:
				elapsed := t.Sub(startTime)
				if elapsed >= duration {
					limiter.SetLimit(rate.Limit(targetLimit))
					goto NextStage
				}
				progress := float64(elapsed) / float64(duration)
				currentRate := startLimit + (targetLimit-startLimit)*progress
				limiter.SetLimit(rate.Limit(currentRate))
			}
		}
	NextStage:
		ticker.Stop()
	}
}

func (e *Engine) executeStep(ctx context.Context, step models.Step, session map[string]string) models.Result {
	start := time.Now()

	// 0. Pre-process Variables (Save/Persist)
	for k, v := range step.Variables {
		session[k] = e.vp.Process(v, session)
	}

	// 1. Process Templates (URL, Body, Headers)
	url := e.vp.Process(step.URL, session)
	method := step.Method
	bodyStr := e.vp.Process(step.Body, session)

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBufferString(bodyStr))
	if err != nil {
		return models.Result{Timestamp: start, Latency: time.Since(start), Error: err, StepName: step.Name}
	}

	// Set default and custom headers
	req.Header.Set("User-Agent", "Sayl/1.0")
	req.Header.Set("Accept", "*/*")
	for k, v := range step.Headers {
		req.Header.Set(k, e.vp.Process(v, session))
	}

	// 2. Execute Request
	resp, err := e.client.Do(req)
	latency := time.Since(start)
	if err != nil {
		return models.Result{Timestamp: start, Latency: latency, Error: err, StepName: step.Name, Protocol: ""}
	}
	defer resp.Body.Close()

	// Capture the actual protocol used (HTTP/1.1, HTTP/2.0, etc.)
	protocol := resp.Proto

	// 3. Read Body (for size & extraction & assertions)
	// If we need to extract OR validate assertions, we must read the whole body.
	var bodyBytes []byte
	var written int64

	needBody := len(step.Extract) > 0 || len(step.Assertions) > 0
	if needBody {
		bodyBytes, err = io.ReadAll(resp.Body)
		written = int64(len(bodyBytes))
	} else {
		written, _ = io.Copy(io.Discard, resp.Body)
	}

	// 4. Extract Variables
	if err == nil && len(step.Extract) > 0 && len(bodyBytes) > 0 {
		// Use gjson for fast extraction

		for varName, path := range step.Extract {
			// path format: "json:data.token" or just "data.token"
			// New support for "header:Header-Name"
			if strings.HasPrefix(path, "header:") {
				headerName := strings.TrimPrefix(path, "header:")
				val := resp.Header.Get(headerName)
				if val != "" {
					session[varName] = val
				}
				continue
			}

			// Default: JSON extraction from body
			val := gjson.GetBytes(bodyBytes, path).String()
			if val != "" {
				session[varName] = val
			}
		}
	}

	// 5. Validate Assertions (NEW!)
	var assertionErr error
	if len(step.Assertions) > 0 && len(bodyBytes) > 0 {
		assertionErr = validator.ValidateAssertions(bodyBytes, step.Assertions)
	}

	return models.Result{
		Timestamp:      start,
		Latency:        latency,
		Status:         resp.StatusCode,
		Bytes:          written,
		AssertionError: assertionErr, // Separate from network errors!
		StepName:       step.Name,
		Protocol:       protocol,
	}
}

// executeStepWithRetry wraps executeStep with automatic retry logic for transient errors
func (e *Engine) executeStepWithRetry(ctx context.Context, step models.Step, session map[string]string) models.Result {
	var result models.Result

	for attempt := 0; attempt <= e.retry.MaxRetries; attempt++ {
		// Check if context is cancelled before attempt
		select {
		case <-ctx.Done():
			return models.Result{
				Timestamp: time.Now(),
				Error:     ctx.Err(),
			}
		default:
		}

		result = e.executeStep(ctx, step, session)

		// If successful or non-retryable error, return immediately
		if result.Error == nil || !isRetryableError(result.Error) {
			return result
		}

		// Don't sleep on the last attempt
		if attempt < e.retry.MaxRetries {
			// Exponential backoff: delay * 2^attempt
			backoff := e.retry.RetryDelay * time.Duration(1<<attempt)
			select {
			case <-ctx.Done():
				return result
			case <-time.After(backoff):
			}
		}
	}

	return result
}
