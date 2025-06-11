package pipeline

import (
	"regexp"
	"strings"
)

type SplitStrategy func(string) []string

var strategies = []SplitStrategy{
	splitByIPTimestampPattern,
	splitByMultipleJSONObjects,
	splitByHTTPVerb,
	splitByDateTimestamp,
	splitByNewline,
}

func SplitLog(raw string, logType LogType) []string {
	for _, strategy := range strategies {
		result := strategy(raw)
		if len(result) > 1 {
			return result
		}
	}
	return []string{strings.TrimSpace(raw)}
}

func splitByNewline(raw string) []string {
	if !strings.Contains(raw, "\n") {
		return []string{strings.TrimSpace(raw)}
	}
	lines := strings.Split(raw, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitByIPTimestampPattern(raw string) []string {
	pattern := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+\s+-\s+-\s+\[.+?\]`)
	return splitByRegex(raw, pattern)
}

func splitByHTTPVerb(raw string) []string {
	pattern := regexp.MustCompile(`"(GET|POST|PUT|DELETE|HEAD)\s+[^\"]+"`)
	return splitByRegex(raw, pattern)
}

func splitByDateTimestamp(raw string) []string {
	pattern := regexp.MustCompile(`\d{4}[-/]\d{2}[-/]\d{2}[ T]\d{2}:\d{2}:\d{2}`)
	return splitByRegex(raw, pattern)
}

func splitByMultipleJSONObjects(raw string) []string {
	pattern := regexp.MustCompile(`}\s*{`)
	indices := pattern.FindAllStringIndex(raw, -1)
	if len(indices) == 0 {
		return []string{strings.TrimSpace(raw)}
	}
	var result []string
	start := 0
	for _, idx := range indices {
		end := idx[0] + 1
		chunk := strings.TrimSpace(raw[start:end])
		if chunk != "" {
			result = append(result, chunk)
		}
		start = idx[1] - 1
	}
	last := strings.TrimSpace(raw[start:])
	if last != "" {
		result = append(result, last)
	}
	return result
}

func splitByRegex(raw string, pattern *regexp.Regexp) []string {
	indices := pattern.FindAllStringIndex(raw, -1)
	if len(indices) <= 1 {
		return []string{strings.TrimSpace(raw)}
	}
	var result []string
	for i := 0; i < len(indices); i++ {
		start := indices[i][0]
		end := len(raw)
		if i+1 < len(indices) {
			end = indices[i+1][0]
		}
		chunk := strings.TrimSpace(raw[start:end])
		if chunk != "" {
			result = append(result, chunk)
		}
	}
	return result
}
