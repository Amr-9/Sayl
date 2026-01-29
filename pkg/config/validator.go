package config

import (
	"fmt"
	"strings"
)

// ValidationError represents a single validation error with context and suggestions
type ValidationError struct {
	Field      string // Field path (e.g., "load.concurrency")
	Value      string // The actual value provided (if any)
	Message    string // Error description
	Expected   string // Expected format/type
	Hint       string // Helpful suggestion
	DidYouMean string // Typo correction suggestion
}

// ValidationResult holds all validation errors
type ValidationResult struct {
	Errors []ValidationError
}

// Add adds a new validation error
func (v *ValidationResult) Add(err ValidationError) {
	v.Errors = append(v.Errors, err)
}

// HasErrors returns true if there are validation errors
func (v *ValidationResult) HasErrors() bool {
	return len(v.Errors) > 0
}

// FormatErrors formats all errors into a user-friendly string
func (v *ValidationResult) FormatErrors() string {
	if !v.HasErrors() {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n❌ Configuration Errors:\n")

	for i, err := range v.Errors {
		sb.WriteString(fmt.Sprintf("\n  %d. %s\n", i+1, err.Field))

		if err.Value != "" {
			sb.WriteString(fmt.Sprintf("     ├─ Value: %q\n", truncate(err.Value, 50)))
		}

		sb.WriteString(fmt.Sprintf("     ├─ Error: %s\n", err.Message))

		if err.Expected != "" {
			sb.WriteString(fmt.Sprintf("     ├─ Expected: %s\n", err.Expected))
		}

		if err.DidYouMean != "" {
			sb.WriteString(fmt.Sprintf("     ├─ Did you mean: %q?\n", err.DidYouMean))
		}

		if err.Hint != "" {
			sb.WriteString(fmt.Sprintf("     └─ 💡 Hint: %s\n", err.Hint))
		} else {
			// Replace last ├ with └ for cleaner output
			// This is handled by putting hint last
		}
	}

	sb.WriteString("\n📖 For documentation, see: https://github.com/Amr-9/sayl#-yaml-configuration-guide\n")

	return sb.String()
}

// Known valid field names for typo detection
var validTargetFields = []string{"url", "method", "headers", "body", "body_file", "body_json", "timeout", "insecure", "keep_alive", "http2", "http2_only", "h2c"}
var validLoadFields = []string{"duration", "rate", "concurrency", "success_codes", "stages"}
var validStepFields = []string{"name", "url", "method", "headers", "body", "body_file", "body_json", "extract", "variables", "save"}
var validHTTPMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

// Hints for common fields
var fieldHints = map[string]string{
	"target.url":         "Provide the full URL including protocol (e.g., https://api.example.com/v1/users)",
	"target.method":      "HTTP method: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS",
	"target.timeout":     "Request timeout with unit (e.g., '10s', '30s', '1m')",
	"target.http2":       "Enable HTTP/2 support (true/false, default: true for HTTPS)",
	"target.http2_only":  "Force HTTP/2 only - fail if server doesn't support it",
	"target.h2c":         "Enable HTTP/2 Cleartext for non-TLS URLs (development/testing only)",
	"load.duration":      "Test duration with unit (e.g., '30s', '2m', '1h')",
	"load.rate":          "Requests per second as a positive integer (e.g., 100)",
	"load.concurrency":   "Number of concurrent workers as a positive integer (e.g., 10)",
	"load.success_codes": "List of HTTP status codes to count as success (e.g., [200, 201])",
	"load.stages":        "List of stages with 'duration' and 'target' rate for ramping",
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(a, b string) int {
	a = strings.ToLower(a)
	b = strings.ToLower(b)

	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create matrix
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

// FindClosestMatch finds the closest matching field name from valid options
func FindClosestMatch(input string, validOptions []string) string {
	if input == "" {
		return ""
	}

	bestMatch := ""
	bestDistance := 100 // arbitrary large number

	for _, option := range validOptions {
		distance := levenshteinDistance(input, option)
		// Only suggest if distance is reasonable (less than half the word length)
		if distance < bestDistance && distance <= len(option)/2+1 {
			bestDistance = distance
			bestMatch = option
		}
	}

	// Don't return exact matches as "did you mean"
	if strings.EqualFold(input, bestMatch) {
		return ""
	}

	return bestMatch
}

// GetHint returns a helpful hint for a field
func GetHint(field string) string {
	if hint, ok := fieldHints[field]; ok {
		return hint
	}
	return ""
}

// ValidateHTTPMethod checks if a method is valid and suggests corrections
func ValidateHTTPMethod(method string) (bool, string) {
	upper := strings.ToUpper(method)
	for _, valid := range validHTTPMethods {
		if upper == valid {
			return true, ""
		}
	}

	// Try to find close match
	suggestion := FindClosestMatch(method, validHTTPMethods)
	return false, suggestion
}

// truncate shortens a string for display
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
