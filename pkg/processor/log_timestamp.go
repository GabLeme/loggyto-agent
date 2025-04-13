package processor

import (
	"regexp"
	"time"
)

var timestampRegexes = []*regexp.Regexp{
	regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3}`),          // 2025-04-02 14:45:30.123
	regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2},\d{3}`),           // 2025-04-02 14:45:30,456
	regexp.MustCompile(`\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`),                 // 2025/04/02 00:08:24
	regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3} UTC`),      // 2025-04-02 01:00:06.302 UTC
	regexp.MustCompile(`\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2} [+-]\d{4}`), // 02/Apr/2025:00:08:32 +0000
	regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z`),         // 2025-04-03T00:02:10.305Z
}

func TryExtractTimestamp(logLine string) (string, bool, bool) {
	for _, regex := range timestampRegexes {
		match := regex.FindString(logLine)
		if match != "" {
			if parsed, ok := normalizeTimestamp(match); ok {
				return parsed, true, false
			}
		}
	}

	return time.Now().UTC().Format(time.RFC3339Nano), true, true
}

func normalizeTimestamp(raw string) (string, bool) {
	layouts := []string{
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05,000",
		"2006/01/02 15:04:05",
		"2006-01-02 15:04:05.000 UTC",
		"02/Jan/2006:15:04:05 -0700",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC().Format(time.RFC3339Nano), true
		}
	}
	return "", false
}
