package processor

import "regexp"

var logLevelPatterns = []struct {
	level   string
	pattern *regexp.Regexp
}{
	{"ERROR", regexp.MustCompile(`(?i)\b(error|fatal|critical)\b`)},
	{"WARN", regexp.MustCompile(`(?i)\b(warn|warning)\b`)},
	{"DEBUG", regexp.MustCompile(`(?i)\b(debug|trace)\b`)},
	{"INFO", regexp.MustCompile(`(?i)\b(info|notice)\b`)},
}

func detectLogLevel(logData string) string {
	for _, lp := range logLevelPatterns {
		if lp.pattern.MatchString(logData) {
			return lp.level
		}
	}
	return "INFO"
}
