package circuitbreaker

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Amr-9/sayl/pkg/models"
)

// Breaker monitors error rates and trips when thresholds are exceeded
type Breaker struct {
	config  *models.CircuitBreaker
	tripped int32 // atomic: 0 = closed, 1 = open
	reason  string
	mu      sync.Mutex
}

// NewBreaker creates a circuit breaker from config
func NewBreaker(cfg *models.CircuitBreaker) (*Breaker, error) {
	if cfg == nil {
		return nil, nil
	}

	// Parse the condition
	if err := ParseCondition(cfg); err != nil {
		return nil, err
	}

	// Set default min_samples if not specified (cold start protection)
	if cfg.MinSamples <= 0 {
		cfg.MinSamples = 100 // Default: need at least 100 samples before tripping
	}

	return &Breaker{
		config: cfg,
	}, nil
}

// conditionPattern matches expressions like "errors > 10%" or "error_rate > 0.1"
var conditionPattern = regexp.MustCompile(`(?i)(errors?|error_rate|failures?)\s*([><=]+)\s*([\d.]+)(%)?`)

// ParseCondition parses the stop_if expression and populates the config fields
func ParseCondition(cfg *models.CircuitBreaker) error {
	expr := strings.TrimSpace(cfg.StopIf)
	if expr == "" {
		return fmt.Errorf("empty circuit breaker condition")
	}

	matches := conditionPattern.FindStringSubmatch(expr)
	if matches == nil {
		return fmt.Errorf("invalid circuit breaker condition '%s'. Expected format: 'errors > 10%%' or 'error_rate > 0.1'", expr)
	}

	// matches[1] = metric (errors, error_rate, failures)
	// matches[2] = operator (>, <, >=, <=)
	// matches[3] = threshold value
	// matches[4] = % sign (optional)

	cfg.Metric = strings.ToLower(matches[1])
	cfg.Operator = matches[2]

	threshold, err := strconv.ParseFloat(matches[3], 64)
	if err != nil {
		return fmt.Errorf("invalid threshold value '%s': %w", matches[3], err)
	}
	cfg.Threshold = threshold
	cfg.IsPercent = matches[4] == "%"

	// Normalize metric names
	switch cfg.Metric {
	case "error", "errors":
		cfg.Metric = "errors"
	case "failure", "failures":
		cfg.Metric = "failures"
	case "error_rate":
		cfg.Metric = "error_rate"
	}

	return nil
}

// Check evaluates whether the circuit breaker should trip based on current stats.
// Returns true if the breaker has tripped (test should stop).
func (b *Breaker) Check(totalRequests, failures, assertionFailures int64) bool {
	if b == nil || b.config == nil {
		return false
	}

	// Already tripped?
	if atomic.LoadInt32(&b.tripped) == 1 {
		return true
	}

	// Cold start protection: don't trip until we have enough samples
	if totalRequests < b.config.MinSamples {
		return false
	}

	// Calculate current error rate
	totalErrors := failures + assertionFailures
	var currentValue float64

	switch b.config.Metric {
	case "errors", "error_rate":
		if b.config.IsPercent {
			// Percentage: errors > 10%
			currentValue = float64(totalErrors) / float64(totalRequests) * 100
		} else {
			// Rate: error_rate > 0.1
			currentValue = float64(totalErrors) / float64(totalRequests)
		}
	case "failures":
		// Absolute count: failures > 100
		currentValue = float64(totalErrors)
	default:
		return false
	}

	// Evaluate condition
	shouldTrip := false
	switch b.config.Operator {
	case ">":
		shouldTrip = currentValue > b.config.Threshold
	case ">=":
		shouldTrip = currentValue >= b.config.Threshold
	case "<":
		shouldTrip = currentValue < b.config.Threshold
	case "<=":
		shouldTrip = currentValue <= b.config.Threshold
	}

	if shouldTrip {
		b.mu.Lock()
		if atomic.CompareAndSwapInt32(&b.tripped, 0, 1) {
			if b.config.IsPercent {
				b.reason = fmt.Sprintf("Circuit breaker tripped: %s (%.1f%%) exceeded threshold (%.1f%%)",
					b.config.Metric, currentValue, b.config.Threshold)
			} else {
				b.reason = fmt.Sprintf("Circuit breaker tripped: %s (%.3f) exceeded threshold (%.3f)",
					b.config.Metric, currentValue, b.config.Threshold)
			}
		}
		b.mu.Unlock()
		return true
	}

	return false
}

// IsTripped returns whether the breaker has tripped
func (b *Breaker) IsTripped() bool {
	if b == nil {
		return false
	}
	return atomic.LoadInt32(&b.tripped) == 1
}

// Reason returns the reason for tripping (empty if not tripped)
func (b *Breaker) Reason() string {
	if b == nil {
		return ""
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.reason
}

// Reset resets the circuit breaker state
func (b *Breaker) Reset() {
	if b == nil {
		return
	}
	atomic.StoreInt32(&b.tripped, 0)
	b.mu.Lock()
	b.reason = ""
	b.mu.Unlock()
}
