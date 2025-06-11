package pipeline

import (
	"regexp"
	"time"
)

var timestampPatterns = []string{
	time.RFC3339,
	time.RFC3339Nano,
	time.RFC1123,
	time.RFC1123Z,
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05.000Z07:00",
	"2006-01-02T15:04:05.000Z",
	"2006-01-02 15:04:05",
	"02/Jan/2006:15:04:05 -0700",
	"Jan 2 15:04:05",
	"01/02/2006 15:04:05",
	"2006/01/02 15:04:05",
	"2006-01-02 15:04:05.999999999 -0700 MST",
}

var extractRegex = regexp.MustCompile(`(?i)(\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:?\d{2}| UTC)?|\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2} [-+]\d{4}|[A-Za-z]{3} \d{1,2} \d{2}:\d{2}:\d{2}|\d{2}/\d{2}/\d{4} \d{2}:\d{2}:\d{2})`)

func TryExtractTimestamp(msg string) (time.Time, bool) {
	match := extractRegex.FindString(msg)
	if match == "" {
		return time.Now(), true
	}
	for _, layout := range timestampPatterns {
		if ts, err := time.Parse(layout, match); err == nil {
			return ts, false
		}
	}
	return time.Now(), true
}
