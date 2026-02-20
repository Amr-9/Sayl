package attacker

import "strings"

// templatePart is either a static literal or a variable/function reference.
type templatePart struct {
	isLiteral bool
	literal   string // set when isLiteral == true
	ref       string // content between {{ and }}, set when isLiteral == false
}

// CompiledTemplate is a pre-parsed template ready for fast per-request execution.
// Parsing happens once at config load time; only substitution runs per request.
type CompiledTemplate struct {
	parts   []templatePart
	hasVars bool // false → purely static string, Execute returns parts[0].literal
}

// CompileTemplate parses a template string into a CompiledTemplate.
// Must be called once per template string (URL, body, header value, variable value).
func CompileTemplate(input string) *CompiledTemplate {
	// Fast-path: no placeholders at all.
	if strings.IndexByte(input, '{') == -1 || !strings.Contains(input, "{{") {
		return &CompiledTemplate{
			parts:   []templatePart{{isLiteral: true, literal: input}},
			hasVars: false,
		}
	}

	ct := &CompiledTemplate{hasVars: true}
	remaining := input
	for {
		start := strings.Index(remaining, "{{")
		if start == -1 {
			if remaining != "" {
				ct.parts = append(ct.parts, templatePart{isLiteral: true, literal: remaining})
			}
			break
		}
		// Literal text before {{
		if start > 0 {
			ct.parts = append(ct.parts, templatePart{isLiteral: true, literal: remaining[:start]})
		}
		afterOpen := remaining[start+2:]
		end := strings.Index(afterOpen, "}}")
		if end == -1 {
			// Unterminated — treat the rest as a literal.
			ct.parts = append(ct.parts, templatePart{isLiteral: true, literal: remaining[start:]})
			break
		}
		ref := strings.TrimSpace(afterOpen[:end])
		ct.parts = append(ct.parts, templatePart{isLiteral: false, ref: ref})
		remaining = afterOpen[end+2:]
	}
	return ct
}

// Execute renders the compiled template using the given session map and variable processor.
// Called on every request — designed to minimise allocations.
func (ct *CompiledTemplate) Execute(vp *VariableProcessor, session map[string]string) string {
	if !ct.hasVars {
		// Static string: return the single literal directly with zero allocations.
		return ct.parts[0].literal
	}

	// Pre-size the builder with the total literal content length + headroom for vars.
	literalLen := 0
	for i := range ct.parts {
		if ct.parts[i].isLiteral {
			literalLen += len(ct.parts[i].literal)
		}
	}

	var sb strings.Builder
	sb.Grow(literalLen + 64)

	for i := range ct.parts {
		p := &ct.parts[i]
		if p.isLiteral {
			sb.WriteString(p.literal)
			continue
		}
		// Variable or function call.
		if idx := strings.IndexByte(p.ref, '('); idx != -1 && strings.HasSuffix(p.ref, ")") {
			funcName := strings.TrimSpace(p.ref[:idx])
			argStr := p.ref[idx+1 : len(p.ref)-1]
			if f, ok := vp.funcMap[funcName]; ok {
				sb.WriteString(f(parseArgs(argStr)))
			} else {
				// Unknown function — emit the original placeholder.
				sb.WriteString("{{")
				sb.WriteString(p.ref)
				sb.WriteString("}}")
			}
		} else {
			sb.WriteString(vp.getValue(p.ref, session))
		}
	}

	return sb.String()
}

// compiledStep holds pre-compiled templates for a single scenario step.
type compiledStep struct {
	url     *CompiledTemplate
	body    *CompiledTemplate
	headers map[string]*CompiledTemplate
	vars    map[string]*CompiledTemplate
}
