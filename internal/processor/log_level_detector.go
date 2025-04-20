package processor

import (
	"regexp"
	"strings"
)

func detectLogLevel(logData string) string {
	logData = strings.ToLower(logData)

	levelPatterns := map[string]*regexp.Regexp{
		"ERROR": regexp.MustCompile(`(?i)\b(error|fatal|critical)\b`),
		"WARN":  regexp.MustCompile(`(?i)\b(warn|warning)\b`),
		"DEBUG": regexp.MustCompile(`(?i)\b(debug|trace)\b`),
		"INFO":  regexp.MustCompile(`(?i)\b(info|notice)\b`),
	}

	for level, pattern := range levelPatterns {
		if pattern.MatchString(logData) {
			return level
		}
	}

	return "INFO"
}
