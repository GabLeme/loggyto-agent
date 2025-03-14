package processor

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
	"unicode"
)

func cleanLogMessage(logData string) string {
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, logData)

	return strings.TrimSpace(cleaned)
}

func parseJSONMessage(logData string) (interface{}, bool) {
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(logData), &jsonData); err == nil {
		return jsonData, true
	}
	return logData, false
}

func normalizeTimestamp(logData string) string {
	timePatterns := []string{
		`(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})`,
		`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z)`,
	}

	for _, pattern := range timePatterns {
		re := regexp.MustCompile(pattern)
		match := re.FindString(logData)
		if match != "" {
			parsedTime, err := time.Parse("2006/01/02 15:04:05", match)
			if err == nil {
				return parsedTime.UTC().Format(time.RFC3339)
			}
		}
	}

	return time.Now().UTC().Format(time.RFC3339)
}
