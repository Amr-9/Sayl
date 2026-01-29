package debug

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Amr-9/sayl/internal/attacker"
	"github.com/Amr-9/sayl/internal/validator"
	"github.com/Amr-9/sayl/pkg/models"
	"github.com/tidwall/gjson"
	"golang.org/x/net/http2"
)

// ANSI color codes for terminal output
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
)

// RunDebugMode executes a single iteration of the scenario with detailed output
func RunDebugMode(cfg *models.Config) error {
	fmt.Println()
	fmt.Printf("%s%s🛠️  STARTING DEBUG MODE (Dry Run) 🛠️%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%sRunning 1 iteration with 1 worker...%s\n\n", colorDim, colorReset)

	// Create HTTP client with same settings as the real attacker
	var roundTripper http.RoundTripper

	if cfg.H2C {
		// HTTP/2 Cleartext (h2c) - for non-TLS HTTP/2 testing
		roundTripper = &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				return (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext(ctx, network, addr)
			},
		}
	} else {
		transport := &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: cfg.Insecure},
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   !cfg.KeepAlive,
			ForceAttemptHTTP2:   cfg.HTTP2, // Default: true
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		}

		// Always configure HTTP/2 support (with automatic fallback to HTTP/1.1)
		if cfg.HTTP2 {
			_ = http2.ConfigureTransport(transport)
		}

		roundTripper = transport
	}

	client := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: roundTripper,
	}
	if client.Timeout == 0 {
		client.Timeout = 30 * time.Second
	}

	// Initialize Variable Processor (reusing the real one for logic parity)
	vp := attacker.NewVariableProcessor()

	// Initialize Data Feeders
	feeders := make(map[string]*attacker.CSVFeeder)
	for _, d := range cfg.Data {
		f, err := attacker.NewCSVFeeder(d.Path)
		if err != nil {
			return fmt.Errorf("failed to load data feeder '%s': %w", d.Name, err)
		}
		feeders[d.Name] = f
	}

	// Prepare steps
	steps := cfg.Steps
	if len(steps) == 0 {
		// Create a single step from the main config
		steps = []models.Step{{
			Name:    "Main Request",
			URL:     cfg.URL,
			Method:  cfg.Method,
			Headers: cfg.Headers,
			Body:    string(cfg.Body),
		}}
	}

	// Initialize session (variables storage)
	session := make(map[string]string)

	// Feed Data from CSV files
	for name, f := range feeders {
		data := f.Next()
		for k, v := range data {
			session[name+"."+k] = v
		}
	}

	// Execute each step
	allSuccess := true
	for i, step := range steps {
		printStepHeader(i+1, step.Name)

		success, err := executeDebugStep(client, vp, step, session, cfg)
		if err != nil {
			fmt.Printf("\n%s❌ Error executing step: %v%s\n", colorRed, err, colorReset)
			allSuccess = false
			break
		}
		if !success {
			allSuccess = false
			break
		}
	}

	// Final summary
	printSeparator()
	if allSuccess {
		fmt.Printf("%s%s✅ DEBUG SESSION COMPLETED SUCCESSFULLY%s\n\n", colorBold, colorGreen, colorReset)
	} else {
		fmt.Printf("%s%s❌ DEBUG SESSION COMPLETED WITH ERRORS%s\n\n", colorBold, colorRed, colorReset)
	}

	return nil
}

// executeDebugStep runs a single step with detailed output
func executeDebugStep(client *http.Client, vp *attacker.VariableProcessor, step models.Step, session map[string]string, cfg *models.Config) (bool, error) {
	// 0. Pre-process Variables (Save/Persist) - same as real attacker
	for k, v := range step.Variables {
		session[k] = vp.Process(v, session)
	}

	// 1. Process Templates (URL, Body, Headers) - same as real attacker
	url := vp.Process(step.URL, session)
	method := step.Method
	if method == "" {
		method = "GET"
	}
	bodyStr := vp.Process(step.Body, session)

	// Create request
	req, err := http.NewRequest(method, url, bytes.NewBufferString(bodyStr))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default and custom headers - same as real attacker
	req.Header.Set("User-Agent", "Sayl/1.0 (Debug Mode)")
	req.Header.Set("Accept", "*/*")
	for k, v := range step.Headers {
		req.Header.Set(k, vp.Process(v, session))
	}

	// Print Request
	printRequest(req, bodyStr)

	// 2. Execute Request
	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		printResponseError(err, latency)
		return false, nil // Not a fatal error, just failed request
	}
	defer resp.Body.Close()

	// 3. Read Response Body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	// Print Response
	printResponse(resp, bodyBytes, latency)

	// 4. Extract Variables - same logic as real attacker
	extractedVars := make(map[string]string)
	if len(step.Extract) > 0 && len(bodyBytes) > 0 {
		for varName, path := range step.Extract {
			// Handle header extraction
			if strings.HasPrefix(path, "header:") {
				headerName := strings.TrimPrefix(path, "header:")
				val := resp.Header.Get(headerName)
				if val != "" {
					session[varName] = val
					extractedVars[varName] = val
				}
				continue
			}

			// Default: JSON extraction from body
			val := gjson.GetBytes(bodyBytes, path).String()
			if val != "" {
				session[varName] = val
				extractedVars[varName] = val
			}
		}

		printExtractedVariables(extractedVars, step.Extract)
	}

	// 5. Validate Assertions - same as real attacker
	if len(step.Assertions) > 0 {
		printAssertions(bodyBytes, step.Assertions, resp.StatusCode, cfg.SuccessCodes)
	} else {
		// Still print status code assertion
		printStatusAssertion(resp.StatusCode, cfg.SuccessCodes)
	}

	// Check if step failed (non-2xx/3xx status)
	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 400

	// If custom success codes defined, check against those
	if len(cfg.SuccessCodes) > 0 {
		isSuccess = cfg.SuccessCodes[resp.StatusCode]
	}

	return isSuccess, nil
}

// printStepHeader prints the step header
func printStepHeader(stepNum int, name string) {
	printSeparator()
	fmt.Printf("%s%s📍 STEP %d: %s%s\n", colorBold, colorMagenta, stepNum, name, colorReset)
	printSeparator()
}

// printSeparator prints a visual separator
func printSeparator() {
	fmt.Printf("%s----------------------------------------------------%s\n", colorDim, colorReset)
}

// printRequest prints the HTTP request details
func printRequest(req *http.Request, body string) {
	fmt.Printf("\n%s[REQUEST]%s\n", colorBold, colorReset)
	fmt.Printf("%s%s%s %s%s%s\n", colorBold, colorGreen, req.Method, colorCyan, req.URL.String(), colorReset)

	// Print headers
	if len(req.Header) > 0 {
		fmt.Printf("%sHeaders:%s\n", colorDim, colorReset)
		// Sort headers for consistent output
		var keys []string
		for k := range req.Header {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			for _, v := range req.Header[k] {
				fmt.Printf("  %s%s:%s %s\n", colorYellow, k, colorReset, v)
			}
		}
	}

	// Print body
	if body != "" {
		fmt.Printf("%sBody:%s\n", colorDim, colorReset)
		printFormattedJSON(body, "  ")
	}
}

// printResponse prints the HTTP response details
func printResponse(resp *http.Response, body []byte, latency time.Duration) {
	fmt.Printf("\n%s[RESPONSE]%s\n", colorBold, colorReset)

	// Protocol with color coding (HTTP/2 in green, HTTP/1.1 in cyan)
	protoColor := colorCyan
	if resp.Proto == "HTTP/2.0" {
		protoColor = colorGreen
	}
	fmt.Printf("%sProtocol:%s %s%s%s\n",
		colorDim, colorReset,
		protoColor, resp.Proto, colorReset)

	// Status with color coding
	statusColor := colorGreen
	if resp.StatusCode >= 400 {
		statusColor = colorRed
	} else if resp.StatusCode >= 300 {
		statusColor = colorYellow
	}
	fmt.Printf("%sStatus:%s %s%d %s%s %s(Time: %s)%s\n",
		colorDim, colorReset,
		statusColor, resp.StatusCode, resp.Status[4:], colorReset,
		colorDim, latency.Round(time.Millisecond), colorReset)

	// Print important headers
	importantHeaders := []string{"Content-Type", "Set-Cookie", "Authorization", "X-Request-Id", "Location"}
	var foundHeaders []string
	for _, h := range importantHeaders {
		if val := resp.Header.Get(h); val != "" {
			foundHeaders = append(foundHeaders, h)
		}
	}
	if len(foundHeaders) > 0 {
		fmt.Printf("%sHeaders:%s\n", colorDim, colorReset)
		for _, h := range foundHeaders {
			val := resp.Header.Get(h)
			// Truncate long values
			if len(val) > 80 {
				val = val[:77] + "..."
			}
			fmt.Printf("  %s%s:%s %s\n", colorYellow, h, colorReset, val)
		}
	}

	// Print body (truncated if too long)
	if len(body) > 0 {
		fmt.Printf("%sBody:%s\n", colorDim, colorReset)
		bodyStr := string(body)
		if len(bodyStr) > 2000 {
			bodyStr = bodyStr[:2000] + "\n  ... (truncated, " + fmt.Sprintf("%d", len(body)) + " bytes total)"
		}
		printFormattedJSON(bodyStr, "  ")
	}
}

// printResponseError prints an error response
func printResponseError(err error, latency time.Duration) {
	fmt.Printf("\n%s[RESPONSE]%s\n", colorBold, colorReset)
	fmt.Printf("%s❌ Request Failed%s %s(Time: %s)%s\n",
		colorRed, colorReset,
		colorDim, latency.Round(time.Millisecond), colorReset)
	fmt.Printf("  %sError:%s %v\n", colorRed, colorReset, err)
}

// printExtractedVariables prints the extracted variables
func printExtractedVariables(vars map[string]string, extractRules map[string]string) {
	fmt.Printf("\n%s[🔍 VARIABLES EXTRACTED]%s\n", colorBold, colorReset)

	if len(vars) == 0 {
		fmt.Printf("  %s⚠️  No variables extracted (paths may not match response)%s\n", colorYellow, colorReset)
		return
	}

	// Sort for consistent output
	var keys []string
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := vars[k]
		source := extractRules[k]
		// Truncate long values
		displayVal := v
		if len(displayVal) > 60 {
			displayVal = displayVal[:57] + "..."
		}
		fmt.Printf("  %s✅ %s%s = %s\"%s\"%s  %s(Source: %s)%s\n",
			colorGreen, colorBold, k, colorCyan, displayVal, colorReset,
			colorDim, source, colorReset)
	}
}

// printAssertions validates and prints assertion results
func printAssertions(body []byte, assertions []models.Assertion, statusCode int, successCodes map[int]bool) {
	fmt.Printf("\n%s[🛡️ ASSERTIONS]%s\n", colorBold, colorReset)

	// Status code assertion
	printStatusAssertion(statusCode, successCodes)

	// User-defined assertions
	for _, assertion := range assertions {
		err := validator.ValidateAssertions(body, []models.Assertion{assertion})

		var assertionDesc string
		switch assertion.Type {
		case models.AssertContains:
			assertionDesc = fmt.Sprintf("Contains \"%s\"", truncate(assertion.Value, 40))
		case models.AssertRegex:
			assertionDesc = fmt.Sprintf("Regex \"%s\"", truncate(assertion.Value, 40))
		case models.AssertJSONPath:
			if assertion.Value != "" {
				assertionDesc = fmt.Sprintf("JSON Path \"%s\" = \"%s\"", assertion.Path, truncate(assertion.Value, 30))
			} else {
				assertionDesc = fmt.Sprintf("JSON Path \"%s\" exists", assertion.Path)
			}
		}

		if err != nil {
			fmt.Printf("  %s❌ %s: FAILED%s\n", colorRed, assertionDesc, colorReset)
			fmt.Printf("     %s└─ %v%s\n", colorDim, err, colorReset)
		} else {
			// For JSON path, show the actual value
			if assertion.Type == models.AssertJSONPath && assertion.Path != "" {
				actual := gjson.GetBytes(body, assertion.Path).String()
				fmt.Printf("  %s✅ %s:%s Passed (Value: \"%s\")\n",
					colorGreen, assertionDesc, colorReset, truncate(actual, 40))
			} else {
				fmt.Printf("  %s✅ %s:%s Passed\n", colorGreen, assertionDesc, colorReset)
			}
		}
	}
}

// printStatusAssertion prints the status code assertion result
func printStatusAssertion(statusCode int, successCodes map[int]bool) {
	isSuccess := statusCode >= 200 && statusCode < 400
	if len(successCodes) > 0 {
		isSuccess = successCodes[statusCode]
	}

	if isSuccess {
		fmt.Printf("  %s✅ Status Code: %d OK%s\n", colorGreen, statusCode, colorReset)
	} else {
		fmt.Printf("  %s❌ Status Code: %d (Expected 2xx/3xx)%s\n", colorRed, statusCode, colorReset)
	}
}

// printFormattedJSON attempts to pretty-print JSON, falls back to raw output
func printFormattedJSON(s string, prefix string) {
	// Try to parse as JSON for pretty printing
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(s), &jsonObj); err == nil {
		// It's valid JSON, pretty print it
		pretty, err := json.MarshalIndent(jsonObj, prefix, "  ")
		if err == nil {
			fmt.Printf("%s%s\n", prefix, string(pretty))
			return
		}
	}

	// Not JSON or failed to format, print as-is with prefix
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		fmt.Printf("%s%s\n", prefix, line)
	}
}

// truncate truncates a string to the given length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
