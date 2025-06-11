package pipeline

import (
	"regexp"
	"strings"
)

var levelPatterns = map[string]*regexp.Regexp{
	"ERROR": regexp.MustCompile(`(?i)\b(error|fatal|fail|exception)\b`),
	"WARN":  regexp.MustCompile(`(?i)\b(warn|warning)\b`),
	"INFO":  regexp.MustCompile(`(?i)\b(info|started|running)\b`),
	"DEBUG": regexp.MustCompile(`(?i)\b(debug|trace)\b`),
}

func DetectLogLevel(msg string) string {
	msg = strings.ToLower(msg)

	for level, pattern := range levelPatterns {
		if pattern.MatchString(msg) {
			return level
		}
	}
	return "INFO"
}
