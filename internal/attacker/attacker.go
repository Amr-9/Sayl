package attacker

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand/v2"
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

// warmConnections pre-establishes TCP/TLS connections so the first real requests
// don't pay the full handshake cost, eliminating the cold-start latency spike.
func (e *Engine) warmConnections(ctx context.Context, url string, count int) {
	if url == "" || count <= 0 {
		return
	}
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
			if err != nil {
				return
			}
			req.Header.Set("User-Agent", "Sayl/1.0 Warmup")
			resp, err := e.client.Do(req)
			if err != nil {
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}()
	}
	wg.Wait()
}

// retryablePatterns holds lowercase patterns for retryable errors — allocated once at startup.
var retryablePatterns = []string{
	"timeout",
	"connection reset",
	"connection refused",
	"no such host",
	"eof",
	"i/o timeout",
	"tls handshake timeout",
}

// isRetryableError checks if an error is worth retrying
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error()) // lowercase once, not per pattern
	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
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
			AllowHTTP:        true,
			MaxHeaderListSize: 16 * 1024,       // reject unexpectedly large response headers
			WriteByteTimeout: 10 * time.Second, // prevent stalled connections from blocking workers
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

	// Pre-warm connections to avoid cold-start latency spikes in the first seconds.
	targetURL := cfg.URL
	if len(cfg.Steps) > 0 {
		targetURL = cfg.Steps[0].URL
	}
	warmCount := cfg.Concurrency / 4
	if warmCount < 2 {
		warmCount = 2
	}
	if warmCount > 32 {
		warmCount = 32
	}
	e.warmConnections(ctx, targetURL, warmCount)

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

	// Pre-compile all step templates so workers avoid repeated string scanning.
	compiled := make([]compiledStep, len(steps))
	for i, step := range steps {
		cs := compiledStep{
			url:     CompileTemplate(step.URL),
			body:    CompileTemplate(step.Body),
			headers: make(map[string]*CompiledTemplate, len(step.Headers)),
			vars:    make(map[string]*CompiledTemplate, len(step.Variables)),
		}
		for k, v := range step.Headers {
			cs.headers[k] = CompileTemplate(v)
		}
		for k, v := range step.Variables {
			cs.vars[k] = CompileTemplate(v)
		}
		compiled[i] = cs
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

					// Execute scenario steps using pre-compiled templates
					for j, step := range steps {
						result := e.executeCompiledStepWithRetry(ctx, step, compiled[j], session)

						// Send result
						select {
						case results <- result:
						case <-ctx.Done():
							// Ensure we return the map even if cancelled here
							e.sessionPool.Put(session)
							return
						}

						// If step failed, break scenario
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


// executeCompiledStep is identical to executeStep but uses pre-compiled templates
// to avoid repeated string scanning on every request.
func (e *Engine) executeCompiledStep(ctx context.Context, step models.Step, cs compiledStep, session map[string]string) models.Result {
	start := time.Now()

	// 0. Pre-process Variables using compiled templates
	for k, ct := range cs.vars {
		session[k] = ct.Execute(e.vp, session)
	}

	// 1. Process Templates via compiled parts (no scanning overhead)
	url := cs.url.Execute(e.vp, session)
	method := step.Method
	bodyStr := cs.body.Execute(e.vp, session)

	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(bodyStr))
	if err != nil {
		return models.Result{Timestamp: start, Latency: time.Since(start), Error: err, StepName: step.Name}
	}

	req.Header.Set("User-Agent", "Sayl/1.0")
	req.Header.Set("Accept", "*/*")
	for k, ct := range cs.headers {
		req.Header.Set(k, ct.Execute(e.vp, session))
	}

	// 2. Execute Request
	resp, err := e.client.Do(req)
	latency := time.Since(start)
	if err != nil {
		return models.Result{Timestamp: start, Latency: latency, Error: err, StepName: step.Name, Protocol: ""}
	}
	defer resp.Body.Close()

	protocol := resp.Proto

	// 3. Read Body
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
		for varName, path := range step.Extract {
			if strings.HasPrefix(path, "header:") {
				headerName := strings.TrimPrefix(path, "header:")
				if val := resp.Header.Get(headerName); val != "" {
					session[varName] = val
				}
				continue
			}
			if val := gjson.GetBytes(bodyBytes, path).String(); val != "" {
				session[varName] = val
			}
		}
	}

	// 5. Validate Assertions
	var assertionErr error
	if len(step.Assertions) > 0 && len(bodyBytes) > 0 {
		assertionErr = validator.ValidateAssertions(bodyBytes, step.Assertions)
	}

	return models.Result{
		Timestamp:      start,
		Latency:        latency,
		Status:         resp.StatusCode,
		Bytes:          written,
		AssertionError: assertionErr,
		StepName:       step.Name,
		Protocol:       protocol,
	}
}

// executeCompiledStepWithRetry wraps executeCompiledStep with the same retry logic.
func (e *Engine) executeCompiledStepWithRetry(ctx context.Context, step models.Step, cs compiledStep, session map[string]string) models.Result {
	var result models.Result

	for attempt := 0; attempt <= e.retry.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return models.Result{Timestamp: time.Now(), Error: ctx.Err()}
		default:
		}

		result = e.executeCompiledStep(ctx, step, cs, session)

		if result.Error == nil || !isRetryableError(result.Error) {
			return result
		}

		if attempt < e.retry.MaxRetries {
			backoff := e.retry.RetryDelay * time.Duration(1<<attempt)
			backoff = time.Duration(float64(backoff) * (0.75 + rand.Float64()*0.5))
			select {
			case <-ctx.Done():
				return result
			case <-time.After(backoff):
			}
		}
	}

	return result
}

