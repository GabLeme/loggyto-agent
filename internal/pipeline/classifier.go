package pipeline

import (
	"encoding/json"
	"regexp"
	"strings"
)

type LogType string

const (
	LogTypePlain     LogType = "PLAIN"
	LogTypeJSON      LogType = "JSON"
	LogTypeAccess    LogType = "ACCESS_LOG"
	LogTypeException LogType = "EXCEPTION"
)

type Classification struct {
	Type       LogType
	Confidence int
	Matches    []string
}

func ClassifyLog(raw string) Classification {
	matches := []string{}

	if isValidJSON(raw) {
		matches = append(matches, "valid JSON")
		return Classification{Type: LogTypeJSON, Confidence: 95, Matches: matches}
	}

	if hasAccessLogPattern(raw) {
		matches = append(matches, "access log pattern")
		return Classification{Type: LogTypeAccess, Confidence: 85, Matches: matches}
	}

	if hasExceptionIndicators(raw) {
		matches = append(matches, "exception keywords")
		return Classification{Type: LogTypeException, Confidence: 90, Matches: matches}
	}

	return Classification{Type: LogTypePlain, Confidence: 60, Matches: []string{"no strong pattern"}}
}

func isValidJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func hasAccessLogPattern(s string) bool {
	accessPattern := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+\s+-\s+-\s+\[.+\]\s+"(GET|POST|PUT|DELETE|HEAD)\s+.+?"\s+\d{3}`)
	return accessPattern.MatchString(s)
}

func hasExceptionIndicators(s string) bool {
	exceptionKeywords := []string{"Exception", "Traceback", "Error", "Caused by", "panic:"}
	for _, kw := range exceptionKeywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}
