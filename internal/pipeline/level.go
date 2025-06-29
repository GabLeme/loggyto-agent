package pipeline

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

var (
	keyValueRegex   = regexp.MustCompile(`(?i)(level|lvl|severity|log_level|loglevel)\s*[:=]\s*(info|warn|warning|error|debug|fatal|trace)`)
	prefixRegex     = regexp.MustCompile(`(?i)^\[?(info|warn|warning|error|debug|fatal|trace)\]?[:\s-]`)
	keywordRegex    = regexp.MustCompile(`(?i)\b(error|fatal|warn|warning|info|debug|trace)\b`)
	httpCodeRegex   = regexp.MustCompile(`\s(\d{3})\s`)
	routeDebugRegex = regexp.MustCompile(`(?i)\b(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\s+\/[^ ]*debug[^ ]*`)
)

var levelMap = map[string]string{
	"error":     "ERROR",
	"err":       "ERROR",
	"fatal":     "ERROR",
	"fail":      "ERROR",
	"exception": "ERROR",
	"warn":      "WARN",
	"warning":   "WARN",
	"debug":     "DEBUG",
	"trace":     "DEBUG",
	"info":      "INFO",
	"started":   "INFO",
	"running":   "INFO",
}

func DetectLogLevel(msg string) string {
	msg = strings.TrimSpace(msg)

	if level := detectFromJSON(msg); level != "" {
		return level
	}
	if level := detectFromKeyValue(msg); level != "" {
		return level
	}
	if level := detectFromPrefix(msg); level != "" {
		return level
	}
	if level := detectFromHTTPCode(msg); level != "" {
		return level
	}
	if level := detectFromKeyword(msg); level != "" {
		return level
	}
	return "INFO"
}

func detectFromJSON(msg string) string {
	if !strings.HasPrefix(msg, "{") {
		return ""
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(msg), &obj); err != nil {
		return ""
	}
	for _, key := range []string{"level", "severity", "lvl", "loglevel"} {
		if val, ok := obj[key]; ok {
			if s, ok := val.(string); ok {
				return normalizeLevel(s)
			}
		}
	}
	return ""
}

func detectFromKeyValue(msg string) string {
	if match := keyValueRegex.FindStringSubmatch(msg); len(match) == 3 {
		return normalizeLevel(match[2])
	}
	return ""
}

func detectFromPrefix(msg string) string {
	if match := prefixRegex.FindStringSubmatch(msg); len(match) == 2 {
		return normalizeLevel(match[1])
	}
	return ""
}

func detectFromHTTPCode(msg string) string {
	if match := httpCodeRegex.FindStringSubmatch(msg); len(match) == 2 {
		if code, err := strconv.Atoi(match[1]); err == nil {
			if code >= 500 {
				return "ERROR"
			} else if code >= 400 {
				return "WARN"
			}
		}
	}
	return ""
}

func detectFromKeyword(msg string) string {
	if match := keywordRegex.FindStringSubmatch(msg); len(match) >= 2 {
		if isRouteKeyword(match[1], msg) {
			return "INFO"
		}
		return normalizeLevel(match[1])
	}
	return ""
}

func normalizeLevel(level string) string {
	if norm, ok := levelMap[strings.ToLower(level)]; ok {
		return norm
	}
	return "INFO"
}

func isRouteKeyword(keyword string, msg string) bool {
	lower := strings.ToLower(keyword)
	return (lower == "debug" || lower == "trace") && routeDebugRegex.MatchString(msg)
}
