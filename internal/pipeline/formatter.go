package pipeline

import (
	"regexp"
	"strings"
)

func CleanLogMessage(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	s = removeAnsiCodes(s)
	s = removeRedundantTimestamps(s)
	s = collapseSpaces(s)

	return s
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func removeAnsiCodes(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

var tsPrefixRegex = regexp.MustCompile(`^(\[?\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}:\d{2}(\.\d+)?(Z| UTC)?\]?[-\s]*)+`)

func removeRedundantTimestamps(s string) string {
	return tsPrefixRegex.ReplaceAllString(s, "")
}

func collapseSpaces(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}
