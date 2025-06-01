package processor

import (
	"regexp"
	"time"
)

type timestampPattern struct {
	regex  *regexp.Regexp
	layout string
}

var timestampPatterns = []timestampPattern{
	{regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3}`), "2006-01-02 15:04:05.000"},
	{regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2},\d{3}`), "2006-01-02 15:04:05,000"},
	{regexp.MustCompile(`\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`), "2006/01/02 15:04:05"},
	{regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3} UTC`), "2006-01-02 15:04:05.000 UTC"},
	{regexp.MustCompile(`\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2} [+-]\d{4}`), "02/Jan/2006:15:04:05 -0700"},
	{regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z`), time.RFC3339Nano},
}

func TryExtractTimestamp(logLine string) (string, bool, bool) {
	for _, pattern := range timestampPatterns {
		match := pattern.regex.FindString(logLine)
		if match != "" {
			if t, err := time.Parse(pattern.layout, match); err == nil {
				return t.UTC().Format(time.RFC3339Nano), true, false
			}
		}
	}
	return time.Now().UTC().Format(time.RFC3339Nano), true, true
}
