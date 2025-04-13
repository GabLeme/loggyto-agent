package processor

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ParsedLog struct {
	Timestamp string
	Level     string
	Message   string
	Metadata  map[string]string
}

var timestampKeys = []string{"timestamp", "time", "t", "@timestamp", "ts", "log_time"}
var messageKeys = []string{"msg", "message", "log", "log_message", "m"}
var levelKeys = []string{"level", "severity", "s", "lvl"}

func TryParseJSONLog(line string) (*ParsedLog, bool) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return nil, false
	}

	result := &ParsedLog{
		Metadata: make(map[string]string),
	}

	// Try timestamp (including nested $date)
	for _, key := range timestampKeys {
		if val, ok := raw[key]; ok {
			if ts := extractTimestamp(val); ts != "" {
				result.Timestamp = ts
				break
			}
		}
	}

	// Try level
	for _, key := range levelKeys {
		if val, ok := raw[key]; ok {
			result.Level = normalizeLevel(fmt.Sprint(val))
			break
		}
	}

	// Try message
	for _, key := range messageKeys {
		if val, ok := raw[key]; ok {
			result.Message = fmt.Sprint(val)
			break
		}
	}

	// Store everything else as metadata
	for k, v := range raw {
		// Already captured
		if contains(k, append(append(timestampKeys, messageKeys...), levelKeys...)) {
			continue
		}
		if submap, ok := v.(map[string]interface{}); ok {
			for sk, sv := range submap {
				result.Metadata[fmt.Sprintf("%s_%s", k, sk)] = fmt.Sprint(sv)
			}
		} else {
			result.Metadata[k] = fmt.Sprint(v)
		}
	}

	return result, true
}

func extractTimestamp(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case float64:
		// Assume it's unix millis
		t := time.Unix(0, int64(v)*int64(time.Millisecond))
		return t.Format(time.RFC3339)
	case map[string]interface{}:
		// MongoDB {"$date": "..."}
		if dateStr, ok := v["$date"].(string); ok {
			return dateStr
		}
	}
	return ""
}

func normalizeLevel(lvl string) string {
	lvl = strings.ToUpper(strings.TrimSpace(lvl))
	switch lvl {
	case "I", "INFO", "30":
		return "INFO"
	case "W", "WARN", "40":
		return "WARN"
	case "E", "ERR", "ERROR", "50":
		return "ERROR"
	case "D", "DEBUG", "20":
		return "DEBUG"
	default:
		return "INFO"
	}
}

func contains(key string, list []string) bool {
	for _, item := range list {
		if item == key {
			return true
		}
	}
	return false
}
