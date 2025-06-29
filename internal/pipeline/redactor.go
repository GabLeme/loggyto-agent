package pipeline

import (
	"regexp"
	"strings"
)

type Redactor struct {
	rules []redactionRule
}

type redactionRule struct {
	pattern     *regexp.Regexp
	replacement string
}

func NewRedactor() *Redactor {
	return &Redactor{
		rules: []redactionRule{
			{
				pattern:     regexp.MustCompile(`(?i)(Authorization:\s*(Bearer\s+|Token\s+))[\w\-\.=]+`),
				replacement: "$1[REDACTED]",
			},
			{
				pattern:     regexp.MustCompile(`(?i)(access_key|api_key|secret_key|token)[=:\s]*[\w\-\.]{8,}`),
				replacement: "$1=[REDACTED]",
			},
			{
				pattern:     regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`),
				replacement: "[REDACTED_EMAIL]",
			},
			{
				pattern:     regexp.MustCompile(`(?i)(password|senha)[=:\s"]+[^"\s]+`),
				replacement: "$1=[REDACTED]",
			},
			{
				pattern:     regexp.MustCompile(`\b\d{3}\.\d{3}\.\d{3}-\d{2}\b`),
				replacement: "[REDACTED_CPF]",
			},
			{
				pattern:     regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`),
				replacement: "[REDACTED_CC]",
			},
			{
				pattern:     regexp.MustCompile(`[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+`),
				replacement: "[REDACTED_JWT]",
			},
		},
	}
}

func (r *Redactor) Redact(input string) string {
	redacted := input

	for _, rule := range r.rules {
		redacted = rule.pattern.ReplaceAllString(redacted, rule.replacement)
	}

	if strings.Contains(redacted, "secret") {
		redacted = regexp.MustCompile(`(?i)(secret)[=:\s"]+[^"\s]+`).ReplaceAllString(redacted, "$1=[REDACTED]")
	}

	return redacted
}
