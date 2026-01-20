package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Amr-9/sayl/internal/circuitbreaker"
	"github.com/Amr-9/sayl/internal/validator"
	"github.com/Amr-9/sayl/pkg/models"
	"gopkg.in/yaml.v3"
)

// YAMLAssertion represents an assertion in YAML format
type YAMLAssertion struct {
	Type    string `yaml:"type"`              // contains, regex, json_path
	Value   string `yaml:"value"`             // Expected value or pattern
	Path    string `yaml:"path,omitempty"`    // JSON path (for json_path type)
	Message string `yaml:"message,omitempty"` // Custom error message
}

// YAMLConfig represents the structure of the YAML configuration file.
type YAMLConfig struct {
	Target struct {
		URL       string            `yaml:"url"`
		Method    string            `yaml:"method,omitempty"`
		Headers   map[string]string `yaml:"headers,omitempty"`
		Body      string            `yaml:"body,omitempty"`
		BodyFile  string            `yaml:"body_file,omitempty"`
		BodyJSON  interface{}       `yaml:"body_json,omitempty"`
		Timeout   string            `yaml:"timeout,omitempty"`
		Insecure  bool              `yaml:"insecure,omitempty"`
		KeepAlive bool              `yaml:"keep_alive,omitempty"`
	} `yaml:"target"`

	Load struct {
		Duration     string `yaml:"duration,omitempty"`
		Rate         int    `yaml:"rate,omitempty"`
		Concurrency  int    `yaml:"concurrency,omitempty"`
		SuccessCodes []int  `yaml:"success_codes,omitempty"`
		StopIf       string `yaml:"stop_if,omitempty"`     // Circuit breaker: "errors > 10%"
		MinSamples   int64  `yaml:"min_samples,omitempty"` // Min samples before circuit breaker can trip
		Stages       []struct {
			Duration string `yaml:"duration"`
			Target   int    `yaml:"target"`
		} `yaml:"stages,omitempty"`
	} `yaml:"load"`
	Steps []struct {
		Name       string            `yaml:"name"`
		URL        string            `yaml:"url"`
		Method     string            `yaml:"method"`
		Headers    map[string]string `yaml:"headers,omitempty"`
		Body       string            `yaml:"body,omitempty"`
		BodyFile   string            `yaml:"body_file,omitempty"`
		BodyJSON   interface{}       `yaml:"body_json,omitempty"`
		Extract    map[string]string `yaml:"extract,omitempty"`
		Variables  map[string]string `yaml:"variables,omitempty"`
		Save       map[string]string `yaml:"save,omitempty"` // Alias for variables
		Assertions []YAMLAssertion   `yaml:"assertions,omitempty"`
	} `yaml:"steps,omitempty"`
	Data []struct {
		Name string `yaml:"name"`
		Path string `yaml:"path"`
	} `yaml:"data,omitempty"`
}

// LoadConfig reads a YAML file and converts it into a models.Config.
func LoadConfig(path string) (*models.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var yamlCfg YAMLConfig
	if err := yaml.Unmarshal(data, &yamlCfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	fmt.Printf("DEBUG: Loaded %d steps from config\n", len(yamlCfg.Steps))

	cfg := &models.Config{
		URL:         yamlCfg.Target.URL,
		Method:      yamlCfg.Target.Method,
		Headers:     yamlCfg.Target.Headers,
		Rate:        yamlCfg.Load.Rate,
		Concurrency: yamlCfg.Load.Concurrency,
		Insecure:    yamlCfg.Target.Insecure,
		KeepAlive:   yamlCfg.Target.KeepAlive,
	}

	// Handle Steps
	if len(yamlCfg.Steps) > 0 {
		for _, s := range yamlCfg.Steps {
			// Merge Variables and Save into one map
			vars := make(map[string]string)
			for k, v := range s.Variables {
				vars[k] = v
			}
			for k, v := range s.Save {
				vars[k] = v
			}

			// Handle Step Body (Direct vs File vs JSON)
			var bodyData []byte
			if s.BodyFile != "" {
				b, err := os.ReadFile(s.BodyFile)
				if err != nil {
					return nil, fmt.Errorf("failed to read step body file '%s': %w", s.BodyFile, err)
				}
				bodyData = b
			} else if s.Body != "" {
				bodyData = []byte(s.Body)
			} else if s.BodyJSON != nil {
				b, err := json.Marshal(s.BodyJSON)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal step body_json: %w", err)
				}
				bodyData = b
			}

			// Convert YAML assertions to model assertions
			var assertions []models.Assertion
			for _, a := range s.Assertions {
				assertion := models.Assertion{
					Type:    models.AssertionType(a.Type),
					Value:   a.Value,
					Path:    a.Path,
					Message: a.Message,
				}
				// Default to "contains" if type not specified
				if assertion.Type == "" {
					assertion.Type = models.AssertContains
				}
				assertions = append(assertions, assertion)
			}

			// Pre-compile regex patterns for performance
			if len(assertions) > 0 {
				if err := validator.CompileAssertions(assertions); err != nil {
					return nil, fmt.Errorf("step '%s': %w", s.Name, err)
				}
			}

			cfg.Steps = append(cfg.Steps, models.Step{
				Name:       s.Name,
				URL:        s.URL,
				Method:     s.Method,
				Headers:    s.Headers,
				Body:       string(bodyData),
				Extract:    s.Extract,
				Variables:  vars,
				Assertions: assertions,
			})
		}
	}

	// Handle Data Sources
	if len(yamlCfg.Data) > 0 {
		for _, d := range yamlCfg.Data {
			cfg.Data = append(cfg.Data, models.DataSource{
				Name: d.Name,
				Path: d.Path,
			})
		}
	}

	// Handle Duration
	if yamlCfg.Load.Duration != "" {
		d, err := time.ParseDuration(yamlCfg.Load.Duration)
		if err != nil {
			return nil, fmt.Errorf("invalid duration format: %w", err)
		}
		cfg.Duration = d
	}

	// Handle Timeout
	if yamlCfg.Target.Timeout != "" {
		d, err := time.ParseDuration(yamlCfg.Target.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout format: %w", err)
		}
		cfg.Timeout = d
	}

	// Handle Stages
	if len(yamlCfg.Load.Stages) > 0 {
		for _, s := range yamlCfg.Load.Stages {
			d, err := time.ParseDuration(s.Duration)
			if err != nil {
				return nil, fmt.Errorf("invalid stage duration format: %w", err)
			}
			cfg.Stages = append(cfg.Stages, models.Stage{
				Duration: d,
				Target:   s.Target,
			})
		}
	}

	// Handle Body (Direct vs File vs JSON)
	if yamlCfg.Target.BodyFile != "" {
		bodyData, err := os.ReadFile(yamlCfg.Target.BodyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read body file '%s': %w", yamlCfg.Target.BodyFile, err)
		}
		cfg.Body = bodyData
	} else if yamlCfg.Target.Body != "" {
		cfg.Body = []byte(yamlCfg.Target.Body)
	} else if yamlCfg.Target.BodyJSON != nil {
		bodyData, err := json.Marshal(yamlCfg.Target.BodyJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body_json: %w", err)
		}
		cfg.Body = bodyData
	}

	// Handle Success Codes
	if len(yamlCfg.Load.SuccessCodes) > 0 {
		cfg.SuccessCodes = make(map[int]bool)
		for _, code := range yamlCfg.Load.SuccessCodes {
			cfg.SuccessCodes[code] = true
		}
	}

	// Handle Circuit Breaker
	if yamlCfg.Load.StopIf != "" {
		cfg.CircuitBreaker = &models.CircuitBreaker{
			StopIf:     yamlCfg.Load.StopIf,
			MinSamples: yamlCfg.Load.MinSamples,
		}
		// Parse and validate the condition
		if err := circuitbreaker.ParseCondition(cfg.CircuitBreaker); err != nil {
			return nil, fmt.Errorf("invalid circuit breaker: %w", err)
		}
		// Set default min_samples if not specified
		if cfg.CircuitBreaker.MinSamples <= 0 {
			cfg.CircuitBreaker.MinSamples = 100 // Cold start protection
		}
	}

	return cfg, nil
}

// Validate checks if the configuration is valid so we can start running immediately.
// Returns detailed errors with suggestions for fixing issues.
func Validate(cfg *models.Config) error {
	result := &ValidationResult{}

	// Target Validation
	if cfg.URL == "" && len(cfg.Steps) == 0 {
		result.Add(ValidationError{
			Field:   "target.url",
			Message: "missing required field",
			Hint:    GetHint("target.url"),
		})
	}

	if cfg.Method == "" {
		if len(cfg.Steps) == 0 {
			cfg.Method = "GET" // Default for single target
		}
	} else {
		// Validate HTTP method
		if valid, suggestion := ValidateHTTPMethod(cfg.Method); !valid {
			err := ValidationError{
				Field:    "target.method",
				Value:    cfg.Method,
				Message:  "invalid HTTP method",
				Expected: "GET, POST, PUT, DELETE, PATCH, HEAD, or OPTIONS",
			}
			if suggestion != "" {
				err.DidYouMean = suggestion
			}
			result.Add(err)
		}
	}

	// Load Profile Validation
	if len(cfg.Stages) > 0 {
		// Stages validation
		for i, stage := range cfg.Stages {
			if stage.Duration <= 0 {
				result.Add(ValidationError{
					Field:    fmt.Sprintf("load.stages[%d].duration", i),
					Message:  "duration must be greater than 0",
					Expected: "duration string with unit (e.g., '30s', '1m')",
					Hint:     "Each stage needs a positive duration",
				})
			}
			if stage.Target < 0 {
				result.Add(ValidationError{
					Field:    fmt.Sprintf("load.stages[%d].target", i),
					Value:    fmt.Sprintf("%d", stage.Target),
					Message:  "target rate cannot be negative",
					Expected: "non-negative integer (0 or greater)",
					Hint:     "Use target: 0 to stop traffic at end of test",
				})
			}
		}
	} else {
		// Fixed rate validation
		if cfg.Rate <= 0 {
			result.Add(ValidationError{
				Field:    "load.rate",
				Value:    fmt.Sprintf("%d", cfg.Rate),
				Message:  "rate must be greater than 0",
				Expected: "positive integer (e.g., 100)",
				Hint:     GetHint("load.rate"),
			})
		}
		if cfg.Duration <= 0 {
			result.Add(ValidationError{
				Field:    "load.duration",
				Message:  "missing or invalid duration",
				Expected: "duration string with unit",
				Hint:     GetHint("load.duration"),
			})
		}
	}

	if cfg.Concurrency <= 0 {
		result.Add(ValidationError{
			Field:    "load.concurrency",
			Value:    fmt.Sprintf("%d", cfg.Concurrency),
			Message:  "concurrency must be greater than 0",
			Expected: "positive integer (e.g., 10)",
			Hint:     GetHint("load.concurrency"),
		})
	}

	// Validate Steps
	for i, step := range cfg.Steps {
		if step.URL == "" {
			result.Add(ValidationError{
				Field:   fmt.Sprintf("steps[%d].url", i),
				Message: "missing required URL",
				Hint:    "Each step must have a URL to request",
			})
		}
		if step.Method == "" {
			result.Add(ValidationError{
				Field:   fmt.Sprintf("steps[%d].method", i),
				Message: "missing required HTTP method",
				Hint:    "Specify method: GET, POST, PUT, DELETE, etc.",
			})
		} else if valid, suggestion := ValidateHTTPMethod(step.Method); !valid {
			err := ValidationError{
				Field:    fmt.Sprintf("steps[%d].method", i),
				Value:    step.Method,
				Message:  "invalid HTTP method",
				Expected: "GET, POST, PUT, DELETE, PATCH, HEAD, or OPTIONS",
			}
			if suggestion != "" {
				err.DidYouMean = suggestion
			}
			result.Add(err)
		}
	}

	// Set default success code if none provided
	if len(cfg.SuccessCodes) == 0 {
		cfg.SuccessCodes = map[int]bool{200: true}
	}

	if result.HasErrors() {
		return fmt.Errorf("%s", result.FormatErrors())
	}

	return nil
}

func dumpErrors(errs []string) string {
	var out string
	for i, e := range errs {
		if i > 0 {
			out += "\n- "
		}
		out += e
	}
	return out
}

// SaveConfig saves the current configuration to a YAML file.
func SaveConfig(path string, cfg *models.Config) error {
	var yamlCfg YAMLConfig
	yamlCfg.Target.URL = cfg.URL
	yamlCfg.Target.Method = cfg.Method
	yamlCfg.Target.Headers = cfg.Headers
	if len(cfg.Body) > 0 {
		// Try to unmarshal as JSON first to save as body_json if possible,
		// but for simplicity, let's just save as string body for now unless it's complex.
		// Actually, let's just save as Body string to be safe and simple.
		yamlCfg.Target.Body = string(cfg.Body)
	}
	if cfg.Timeout > 0 {
		yamlCfg.Target.Timeout = cfg.Timeout.String()
	}
	yamlCfg.Target.Insecure = cfg.Insecure
	yamlCfg.Target.KeepAlive = cfg.KeepAlive

	if len(cfg.Stages) > 0 {
		for _, s := range cfg.Stages {
			yamlCfg.Load.Stages = append(yamlCfg.Load.Stages, struct {
				Duration string `yaml:"duration"`
				Target   int    `yaml:"target"`
			}{
				Duration: s.Duration.String(),
				Target:   s.Target,
			})
		}
	} else {
		yamlCfg.Load.Duration = cfg.Duration.String()
		yamlCfg.Load.Rate = cfg.Rate
	}

	yamlCfg.Load.Concurrency = cfg.Concurrency
	for code := range cfg.SuccessCodes {
		yamlCfg.Load.SuccessCodes = append(yamlCfg.Load.SuccessCodes, code)
	}

	data, err := yaml.Marshal(yamlCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Append usage instructions
	comment := fmt.Sprintf("\n# Run this configuration:\n# ./sayl -config %s\n", filepath.Base(path))
	data = append(data, []byte(comment)...)

	return os.WriteFile(path, data, 0644)
}
