package attacker

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Pre-defined list of popular User-Agent strings for high performance
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (iPad; CPU OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 14; SM-S918B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.144 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.144 Mobile Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Linux; Android 13; SM-A536B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6099.144 Mobile Safari/537.36",
}

// Pre-defined list of ISO 3166-1 alpha-2 country codes
var countryCodes = []string{
	"US", "GB", "CA", "AU", "DE", "FR", "IT", "ES", "NL", "BE",
	"CH", "AT", "SE", "NO", "DK", "FI", "PL", "CZ", "RO", "HU",
	"EG", "SA", "AE", "QA", "KW", "BH", "OM", "JO", "LB", "IQ",
	"IN", "PK", "BD", "ID", "MY", "SG", "TH", "VN", "PH", "JP",
	"KR", "CN", "TW", "HK", "BR", "MX", "AR", "CL", "CO", "ZA",
}

// Character sets for password generation
const (
	lettersLower = "abcdefghijklmnopqrstuvwxyz"
	lettersUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits       = "0123456789"
	hexChars     = "0123456789abcdef"
	symbols      = "!@#$%^&*"
	alphanum     = lettersLower + lettersUpper + digits
)

// VariableProcessor handles the replacement of {{var}} in strings
type VariableProcessor struct{}

// NewVariableProcessor creates a new processor
func NewVariableProcessor() *VariableProcessor {
	return &VariableProcessor{}
}

// Process replaces placeholders in the input string using the session map and dynamic generators.
// It prioritizes session variables over dynamic ones.
func (vp *VariableProcessor) Process(input string, session map[string]string) string {
	if !strings.Contains(input, "{{") {
		return input
	}

	var sb strings.Builder
	lastIdx := 0
	inputLen := len(input)

	for i := 0; i < inputLen; {
		start := strings.Index(input[i:], "{{")
		if start == -1 {
			sb.WriteString(input[i:])
			break
		}
		start += i

		end := strings.Index(input[start:], "}}")
		if end == -1 {
			sb.WriteString(input[i:])
			break
		}
		end += start

		sb.WriteString(input[lastIdx:start])
		varName := strings.TrimSpace(input[start+2 : end])
		val := vp.getValue(varName, session)
		sb.WriteString(val)

		i = end + 2
		lastIdx = i
	}

	return sb.String()
}

func (vp *VariableProcessor) getValue(name string, session map[string]string) string {
	// 1. Check Session
	if val, ok := session[name]; ok {
		return val
	}

	// 2. Check Dynamic Generators
	switch name {
	case "uuid":
		return uuid.New().String()
	case "random_int":
		return fmt.Sprintf("%d", rand.IntN(100000))
	case "timestamp":
		return fmt.Sprintf("%d", time.Now().Unix())
	case "timestamp_ms":
		return fmt.Sprintf("%d", time.Now().UnixMilli())
	case "random_email":
		return fmt.Sprintf("user%d@example.com", rand.IntN(1000000))
	case "random_name":
		names := []string{"Alice", "Bob", "Charlie", "David", "Eve", "Frank", "Grace", "Heidi"}
		return names[rand.IntN(len(names))] + fmt.Sprintf(" %d", rand.IntN(1000))
	case "random_phone":
		return fmt.Sprintf("+1-555-01%02d", rand.IntN(100))
	case "random_domain":
		sub := make([]byte, 4)
		for i := range sub {
			sub[i] = alphanum[rand.IntN(len(alphanum))]
		}
		return fmt.Sprintf("%s.example.com", string(sub))
	case "random_alphanum":
		b := make([]byte, 10)
		for i := range b {
			b[i] = alphanum[rand.IntN(len(alphanum))]
		}
		return string(b)

	// --- New Generators ---
	case "random_bool":
		if rand.IntN(2) == 0 {
			return "false"
		}
		return "true"
	case "random_float":
		return fmt.Sprintf("%.6f", rand.Float64())
	case "iso8601":
		return time.Now().UTC().Format(time.RFC3339)
	case "random_ipv4":
		return fmt.Sprintf("%d.%d.%d.%d", rand.IntN(256), rand.IntN(256), rand.IntN(256), rand.IntN(256))
	case "random_user_agent":
		return userAgents[rand.IntN(len(userAgents))]
	case "random_mac":
		return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
			rand.IntN(256), rand.IntN(256), rand.IntN(256),
			rand.IntN(256), rand.IntN(256), rand.IntN(256))
	case "random_color":
		return fmt.Sprintf("#%02x%02x%02x", rand.IntN(256), rand.IntN(256), rand.IntN(256))
	case "random_password":
		// Generate a 12-char password with mix of upper, lower, digit, symbol
		pw := make([]byte, 12)
		pw[0] = lettersUpper[rand.IntN(len(lettersUpper))]
		pw[1] = lettersLower[rand.IntN(len(lettersLower))]
		pw[2] = digits[rand.IntN(len(digits))]
		pw[3] = symbols[rand.IntN(len(symbols))]
		allChars := alphanum + symbols
		for i := 4; i < 12; i++ {
			pw[i] = allChars[rand.IntN(len(allChars))]
		}
		// Shuffle the password
		rand.Shuffle(len(pw), func(i, j int) { pw[i], pw[j] = pw[j], pw[i] })
		return string(pw)
	case "random_country":
		return countryCodes[rand.IntN(len(countryCodes))]
	}

	// 3. Check for parameterized patterns (optimized without regex)
	if strings.HasPrefix(name, "random_digits_") {
		lengthStr := name[len("random_digits_"):]
		length := parsePositiveInt(lengthStr, 10, 20)
		result := make([]byte, length)
		for i := range result {
			result[i] = digits[rand.IntN(10)]
		}
		return string(result)
	}

	if strings.HasPrefix(name, "random_hex_") {
		lengthStr := name[len("random_hex_"):]
		length := parsePositiveInt(lengthStr, 8, 64)
		result := make([]byte, length)
		for i := range result {
			result[i] = hexChars[rand.IntN(16)]
		}
		return string(result)
	}

	if strings.HasPrefix(name, "random_alphanum_") {
		lengthStr := name[len("random_alphanum_"):]
		length := parsePositiveInt(lengthStr, 10, 64)
		result := make([]byte, length)
		for i := range result {
			result[i] = alphanum[rand.IntN(len(alphanum))]
		}
		return string(result)
	}

	// Fallback: keep placeholder for debugging
	return "{{" + name + "}}"
}

// parsePositiveInt parses a string to int with a default and max value
func parsePositiveInt(s string, defaultVal, maxVal int) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			return defaultVal
		}
	}
	if n <= 0 {
		return defaultVal
	}
	if n > maxVal {
		return maxVal
	}
	return n
}
