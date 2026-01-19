package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Amr-9/sayl/pkg/models"
	"gopkg.in/yaml.v3"
)

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
		Stages       []struct {
			Duration string `yaml:"duration"`
			Target   int    `yaml:"target"`
		} `yaml:"stages,omitempty"`
	} `yaml:"load"`
	Steps []struct {
		Name      string            `yaml:"name"`
		URL       string            `yaml:"url"`
		Method    string            `yaml:"method"`
		Headers   map[string]string `yaml:"headers,omitempty"`
		Body      string            `yaml:"body,omitempty"`
		BodyFile  string            `yaml:"body_file,omitempty"`
		BodyJSON  interface{}       `yaml:"body_json,omitempty"`
		Extract   map[string]string `yaml:"extract,omitempty"`
		Variables map[string]string `yaml:"variables,omitempty"`
		Save      map[string]string `yaml:"save,omitempty"` // Alias for variables
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

			cfg.Steps = append(cfg.Steps, models.Step{
				Name:      s.Name,
				URL:       s.URL,
				Method:    s.Method,
				Headers:   s.Headers,
				Body:      string(bodyData),
				Extract:   s.Extract,
				Variables: vars,
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

	return cfg, nil
}

// Validate checks if the configuration is valid so we can start running immediately.
func Validate(cfg *models.Config) error {
	var errors []string

	// Target Validation
	if cfg.URL == "" && len(cfg.Steps) == 0 {
		errors = append(errors, "missing target URL or scenario steps (target.url or steps)")
	}

	if cfg.Method == "" {
		if len(cfg.Steps) == 0 {
			cfg.Method = "GET" // Default for single target
		}
	} else {
		// specialized validation if needed for method
	}

	// Load Profile Validation
	if len(cfg.Stages) > 0 {
		// Stages validation
		for i, stage := range cfg.Stages {
			if stage.Duration <= 0 {
				errors = append(errors, fmt.Sprintf("stage %d duration must be > 0", i+1))
			}
			if stage.Target < 0 {
				errors = append(errors, fmt.Sprintf("stage %d target rate must be >= 0", i+1))
			}
		}
	} else {
		// Fixed rate validation
		if cfg.Rate <= 0 {
			errors = append(errors, "rate must be greater than 0 (load.rate)")
		}
		if cfg.Duration <= 0 {
			errors = append(errors, "duration must be greater than 0 (load.duration)")
		}
	}

	if cfg.Concurrency <= 0 {
		errors = append(errors, "concurrency must be greater than 0 (load.concurrency)")
	}

	// Set default success code if none provided
	if len(cfg.SuccessCodes) == 0 {
		cfg.SuccessCodes = map[int]bool{200: true}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration errors:\n- %s", dumpErrors(errors))
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
