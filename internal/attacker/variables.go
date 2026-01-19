package attacker

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
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

	// We'll do a simple loop to find and replace {{...}}.
	// For high performance with many variables, a more complex parser might be needed,
	// but strings.Replace is quite optimized for simple cases.
	// However, we need to identify *which* variables are present to avoid blindly trying to replace everything.
	// A better approach for "hot loop" is to find indices of {{ and }}.

	var sb strings.Builder
	lastIdx := 0
	inputLen := len(input)

	for i := 0; i < inputLen; {
		// Find start of {{
		start := strings.Index(input[i:], "{{")
		if start == -1 {
			sb.WriteString(input[i:])
			break
		}
		start += i // Adjust relative index to absolute

		// Find end of }}
		end := strings.Index(input[start:], "}}")
		if end == -1 {
			// Malformed, just write the rest
			sb.WriteString(input[i:])
			break
		}
		end += start // Adjust relative index

		// Write text before {{
		sb.WriteString(input[lastIdx:start])

		// Extract variable name
		varName := strings.TrimSpace(input[start+2 : end])

		// Get value
		val := vp.getValue(varName, session)
		sb.WriteString(val)

		// Advance indices
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
		return fmt.Sprintf("%d", rand.Intn(100000))
	case "timestamp":
		return fmt.Sprintf("%d", time.Now().Unix())
	case "timestamp_ms":
		return fmt.Sprintf("%d", time.Now().UnixMilli())
	case "random_email":
		return fmt.Sprintf("user%d@example.com", rand.Intn(1000000))
	case "random_name":
		names := []string{"Alice", "Bob", "Charlie", "David", "Eve", "Frank", "Grace", "Heidi"}
		return names[rand.Intn(len(names))] + fmt.Sprintf(" %d", rand.Intn(1000))
	case "random_phone":
		return fmt.Sprintf("+1-555-01%02d", rand.Intn(100))
	case "random_domain":
		sub := make([]byte, 4)
		const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
		for i := range sub {
			sub[i] = letters[rand.Intn(len(letters))]
		}
		return fmt.Sprintf("%s.example.com", string(sub))
	case "random_alphanum":
		const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		b := make([]byte, 10)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		return string(b)
	}

	// 4. Check for dynamic patterns
	if strings.HasPrefix(name, "random_digits_") {
		var length int
		if _, err := fmt.Sscanf(name, "random_digits_%d", &length); err == nil && length > 0 {
			if length > 20 { // Cap length to avoid abuse
				length = 20
			}
			digits := make([]byte, length)
			for i := range digits {
				digits[i] = byte(rand.Intn(10) + '0')
			}
			return string(digits)
		}
	}

	// 3. Fallback (keep original or empty?) -> Let's keep a placeholder if not found for debugging,
	// or return empty string? Usually keeping it makes it obvious something failed.
	// But standard behavior is often empty string. Let's return the placeholder for visibility.
	return "{{" + name + "}}"
}
