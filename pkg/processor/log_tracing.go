package processor

import (
	"regexp"
)

func extractTracingInfo(logData string) (string, string) {
	tracePattern := regexp.MustCompile(`(?i)(traceID|trace_id)[=:"']?([a-f0-9\-]{8,})`)
	spanPattern := regexp.MustCompile(`(?i)(spanID|span_id)[=:"']?([a-f0-9\-]{8,})`)

	var traceID, spanID string

	if match := tracePattern.FindStringSubmatch(logData); len(match) == 3 {
		traceID = match[2]
	}
	if match := spanPattern.FindStringSubmatch(logData); len(match) == 3 {
		spanID = match[2]
	}

	return traceID, spanID
}
