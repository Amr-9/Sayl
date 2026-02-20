package attacker

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lucasjones/reggen"
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
type VariableProcessor struct {
	funcMap map[string]func([]string) string
}

// NewVariableProcessor creates a new processor with built-in functions
func NewVariableProcessor() *VariableProcessor {
	vp := &VariableProcessor{}
	vp.initFuncMap()
	return vp
}

func (vp *VariableProcessor) initFuncMap() {
	vp.funcMap = map[string]func([]string) string{
		// --- Crypto & Encoding ---
		"hmac_sha256": func(args []string) string {
			if len(args) != 2 {
				return "ERROR:hmac_sha256_needs_2_args"
			}
			key := []byte(args[0])
			data := []byte(args[1])
			h := hmac.New(sha256.New, key)
			h.Write(data)
			return hex.EncodeToString(h.Sum(nil))
		},
		"base64_encode": func(args []string) string {
			if len(args) != 1 {
				return "ERROR:base64_encode_needs_1_arg"
			}
			return base64.StdEncoding.EncodeToString([]byte(args[0]))
		},
		"md5": func(args []string) string {
			if len(args) != 1 {
				return "ERROR:md5_needs_1_arg"
			}
			hash := md5.Sum([]byte(args[0]))
			return hex.EncodeToString(hash[:])
		},
		"sha256": func(args []string) string {
			if len(args) != 1 {
				return "ERROR:sha256_needs_1_arg"
			}
			hash := sha256.Sum256([]byte(args[0]))
			return hex.EncodeToString(hash[:])
		},

		// --- Advanced Time Travel ---
		"time_future": func(args []string) string {
			// args[0]: duration (e.g. "24h"), args[1]: format (optional, default RFC3339)
			if len(args) < 1 {
				return "ERROR:time_future_needs_duration"
			}
			dur, err := time.ParseDuration(args[0])
			if err != nil {
				return "ERROR:invalid_duration"
			}
			layout := time.RFC3339
			if len(args) >= 2 {
				layout = args[1]
			}
			return time.Now().Add(dur).Format(layout)
		},
		"time_past": func(args []string) string {
			if len(args) < 1 {
				return "ERROR:time_past_needs_duration"
			}
			dur, err := time.ParseDuration(args[0])
			if err != nil {
				return "ERROR:invalid_duration"
			}
			layout := time.RFC3339
			if len(args) >= 2 {
				layout = args[1]
			}
			return time.Now().Add(-dur).Format(layout)
		},

		// --- Logic & Selection ---
		"random_choice": func(args []string) string {
			if len(args) == 0 {
				return ""
			}
			return args[rand.IntN(len(args))]
		},
		"random_int_range": func(args []string) string {
			if len(args) != 2 {
				return "ERROR:random_int_range_needs_min_max"
			}
			min, _ := strconv.Atoi(strings.TrimSpace(args[0]))
			max, _ := strconv.Atoi(strings.TrimSpace(args[1]))
			if max <= min {
				return strconv.Itoa(min)
			}
			return strconv.Itoa(rand.IntN(max-min) + min)
		},
		"random_float_range": func(args []string) string {
			// min, max, decimals (optional)
			if len(args) < 2 {
				return "ERROR:random_float_range_needs_min_max"
			}
			min, _ := strconv.ParseFloat(strings.TrimSpace(args[0]), 64)
			max, _ := strconv.ParseFloat(strings.TrimSpace(args[1]), 64)
			decimals := 2
			if len(args) >= 3 {
				d, err := strconv.Atoi(strings.TrimSpace(args[2]))
				if err == nil {
					decimals = d
				}
			}
			val := min + rand.Float64()*(max-min)
			format := fmt.Sprintf("%%.%df", decimals)
			return fmt.Sprintf(format, val)
		},

		// --- String Manipulation ---
		"random_string": func(args []string) string {
			// length, charset (optional)
			length := 10
			if len(args) >= 1 {
				if l, err := strconv.Atoi(args[0]); err == nil {
					length = l
				}
			}
			chars := alphanum
			if len(args) >= 2 {
				chars = args[1]
			}
			b := make([]byte, length)
			for i := range b {
				b[i] = chars[rand.IntN(len(chars))]
			}
			return string(b)
		},
		"regex_gen": func(args []string) string {
			if len(args) != 1 {
				return "ERROR:regex_gen_needs_pattern"
			}
			res, err := reggen.Generate(args[0], 10) // 10 is max length for repeats
			if err != nil {
				return "ERROR:regex_gen_failed"
			}
			return res
		},
	}
}

// Process replaces placeholders in the input string using the session map and dynamic generators.
// It prioritizes session variables over dynamic ones.
func (vp *VariableProcessor) Process(input string, session map[string]string) string {
	if strings.IndexByte(input, '{') == -1 {
		return input
	}
	if !strings.Contains(input, "{{") {
		return input
	}

	var sb strings.Builder
	sb.Grow(len(input)) // pre-allocate to avoid reallocs during template substitution
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

		// Append text before {{
		sb.WriteString(input[lastIdx:start])

		content := strings.TrimSpace(input[start+2 : end])

		// 1. Check Function Calls: e.g. "func(arg1, arg2)"
		if idx := strings.IndexByte(content, '('); idx != -1 && strings.HasSuffix(content, ")") {
			funcName := strings.TrimSpace(content[:idx])
			argStr := content[idx+1 : len(content)-1]

			// Simple argument parser (splits by comma, handles basic quotes)
			args := parseArgs(argStr)

			if f, ok := vp.funcMap[funcName]; ok {
				sb.WriteString(f(args))
			} else {
				// Function not found, keep literal
				sb.WriteString(input[start : end+2])
			}
		} else {
			// 2. Variable or Legacy Generator
			val := vp.getValue(content, session)
			sb.WriteString(val)
		}

		i = end + 2
		lastIdx = i
	}

	return sb.String()
}

// parseArgs splits a string by comma, respecting quotes (simple implementation)
func parseArgs(s string) []string {
	var args []string
	var current strings.Builder
	inQuote := false

	for _, r := range s {
		switch r {
		case '"':
			inQuote = !inQuote
		case ',':
			if !inQuote {
				args = append(args, strings.TrimSpace(current.String()))
				current.Reset()
				continue
			}
			current.WriteRune(r)
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, strings.TrimSpace(current.String()))
	} else if len(args) > 0 || (len(s) > 0 && s[len(s)-1] == ',') {
		// Handle trailing comma or empty last arg case if needed,
		// but for now just pushing the last buffer if non-empty
		// Logic above handles "a,b" -> "a" then buffer="b". Pushing "b".
	}

	// Edge case for single empty arg? "func()" -> argStr="" -> loop doesn't run -> args nil. Correct.

	// Post-process to remove surrounding quotes if present
	for i, arg := range args {
		if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") && len(arg) >= 2 {
			args[i] = arg[1 : len(arg)-1]
		}
	}

	return args
}

func (vp *VariableProcessor) getValue(name string, session map[string]string) string {
	// 1. Check Session
	if val, ok := session[name]; ok {
		return val
	}

	// 2. Check Legacy Dynamic Generators (Keep for backward compatibility)
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
		rand.Shuffle(len(pw), func(i, j int) { pw[i], pw[j] = pw[j], pw[i] })
		return string(pw)
	case "random_country":
		return countryCodes[rand.IntN(len(countryCodes))]
	}

	// 3. Backward Compatibility Pattern Parsing (Legacy)
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
