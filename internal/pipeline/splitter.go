package pipeline

import (
	"regexp"
	"strings"
)

type SplitStrategy func(string) []string

var strategies = []SplitStrategy{
	splitByIPTimestampPattern,
	splitByMultipleJSONObjects,
	splitByVerboHTTP,
	splitByNewline,
}

func SplitLog(raw string, logType LogType) []string {
	for _, strategy := range strategies {
		result := strategy(raw)
		if len(result) > 1 {
			return result
		}
	}
	return []string{raw}
}

func splitByNewline(raw string) []string {
	if strings.Contains(raw, "\n") {
		lines := strings.Split(raw, "\n")
		var cleaned []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
		return cleaned
	}
	return []string{raw}
}

func splitByIPTimestampPattern(raw string) []string {
	pattern := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+\s+-\s+-\s+\[.+?\])`)
	indices := pattern.FindAllStringIndex(raw, -1)
	if len(indices) <= 1 {
		return []string{raw}
	}
	var parts []string
	for i := 0; i < len(indices); i++ {
		start := indices[i][0]
		end := len(raw)
		if i+1 < len(indices) {
			end = indices[i+1][0]
		}
		chunk := strings.TrimSpace(raw[start:end])
		if chunk != "" {
			parts = append(parts, chunk)
		}
	}
	return parts
}

func splitByVerboHTTP(raw string) []string {
	pattern := regexp.MustCompile(`\"(GET|POST|PUT|DELETE|HEAD)\s+[^\"]+\"`)
	indices := pattern.FindAllStringIndex(raw, -1)
	if len(indices) <= 1 {
		return []string{raw}
	}
	var parts []string
	for i := 0; i < len(indices); i++ {
		start := indices[i][0]
		end := len(raw)
		if i+1 < len(indices) {
			end = indices[i+1][0]
		}
		chunk := strings.TrimSpace(raw[start:end])
		if chunk != "" {
			parts = append(parts, chunk)
		}
	}
	return parts
}

func splitByMultipleJSONObjects(raw string) []string {
	pattern := regexp.MustCompile(`}\s*{`)
	indices := pattern.FindAllStringIndex(raw, -1)
	if len(indices) == 0 {
		return []string{raw}
	}

	parts := []string{}
	start := 0
	for _, idx := range indices {
		end := idx[0] + 1
		chunk := strings.TrimSpace(raw[start:end])
		if chunk != "" {
			parts = append(parts, chunk)
		}
		start = idx[1] - 1
	}
	last := strings.TrimSpace(raw[start:])
	if last != "" {
		parts = append(parts, last)
	}
	return parts
}
