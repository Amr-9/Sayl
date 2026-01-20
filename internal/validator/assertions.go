package validator

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/Amr-9/sayl/pkg/models"
	"github.com/tidwall/gjson"
)

// AssertionError represents a validation failure with detailed context
type AssertionError struct {
	Type     models.AssertionType
	Expected string
	Actual   string
	Path     string
	Message  string
}

func (e *AssertionError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	switch e.Type {
	case models.AssertContains:
		return fmt.Sprintf("assertion failed: response body does not contain '%s'", e.Expected)
	case models.AssertRegex:
		return fmt.Sprintf("assertion failed: response body does not match regex '%s'", e.Expected)
	case models.AssertJSONPath:
		if e.Expected != "" {
			return fmt.Sprintf("assertion failed: json path '%s' expected '%s', got '%s'", e.Path, e.Expected, e.Actual)
		}
		return fmt.Sprintf("assertion failed: json path '%s' not found or empty", e.Path)
	default:
		return fmt.Sprintf("assertion failed: %s", e.Expected)
	}
}

// CompileAssertions pre-compiles regex patterns for assertions at config load time.
// This MUST be called during config parsing, not per-request, to ensure performance.
func CompileAssertions(assertions []models.Assertion) error {
	for i := range assertions {
		if assertions[i].Type == models.AssertRegex {
			compiled, err := regexp.Compile(assertions[i].Value)
			if err != nil {
				return fmt.Errorf("invalid regex pattern '%s': %w", assertions[i].Value, err)
			}
			assertions[i].Regex = compiled
		}
	}
	return nil
}

// ValidateAssertions checks all assertions against the response body.
// Returns nil if all pass, or the first AssertionError encountered.
// For performance, assertions are evaluated in order and fail-fast on first failure.
func ValidateAssertions(body []byte, assertions []models.Assertion) error {
	for _, assertion := range assertions {
		var err error
		switch assertion.Type {
		case models.AssertContains:
			err = validateContains(body, assertion)
		case models.AssertRegex:
			err = validateRegex(body, assertion)
		case models.AssertJSONPath:
			err = validateJSONPath(body, assertion)
		default:
			// Unknown assertion type - treat as contains by default
			err = validateContains(body, assertion)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// validateContains checks if the response body contains the expected string.
// Uses bytes.Contains for performance (no string conversion needed).
func validateContains(body []byte, assertion models.Assertion) error {
	expected := []byte(assertion.Value)
	if !bytes.Contains(body, expected) {
		return &AssertionError{
			Type:     models.AssertContains,
			Expected: assertion.Value,
			Actual:   truncateBody(body, 100),
			Message:  assertion.Message,
		}
	}
	return nil
}

// validateRegex checks if the response body matches the pre-compiled regex pattern.
// The regex MUST be pre-compiled during config load for performance.
func validateRegex(body []byte, assertion models.Assertion) error {
	if assertion.Regex == nil {
		// Fallback: compile on the fly (should not happen if config is loaded correctly)
		compiled, err := regexp.Compile(assertion.Value)
		if err != nil {
			return &AssertionError{
				Type:     models.AssertRegex,
				Expected: assertion.Value,
				Message:  fmt.Sprintf("invalid regex: %v", err),
			}
		}
		assertion.Regex = compiled
	}

	if !assertion.Regex.Match(body) {
		return &AssertionError{
			Type:     models.AssertRegex,
			Expected: assertion.Value,
			Actual:   truncateBody(body, 100),
			Message:  assertion.Message,
		}
	}
	return nil
}

// validateJSONPath uses gjson for high-performance JSON path extraction.
// gjson operates directly on []byte without full unmarshaling, making it ideal for load testing.
func validateJSONPath(body []byte, assertion models.Assertion) error {
	path := assertion.Path
	if path == "" {
		// If no path provided, use value as path for existence check
		path = assertion.Value
	}

	result := gjson.GetBytes(body, path)
	
	// Check if path exists
	if !result.Exists() {
		return &AssertionError{
			Type:     models.AssertJSONPath,
			Path:     path,
			Expected: assertion.Value,
			Actual:   "",
			Message:  assertion.Message,
		}
	}

	// If a value is expected, compare it
	if assertion.Value != "" && assertion.Path != "" {
		actual := result.String()
		// Support various comparison styles
		expected := strings.TrimSpace(assertion.Value)
		actualTrimmed := strings.TrimSpace(actual)
		
		if actualTrimmed != expected {
			return &AssertionError{
				Type:     models.AssertJSONPath,
				Path:     path,
				Expected: expected,
				Actual:   actualTrimmed,
				Message:  assertion.Message,
			}
		}
	}

	return nil
}

// truncateBody returns a truncated version of the body for error messages
func truncateBody(body []byte, maxLen int) string {
	if len(body) <= maxLen {
		return string(body)
	}
	return string(body[:maxLen]) + "..."
}
