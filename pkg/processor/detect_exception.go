package processor

import (
	"regexp"
	"strings"
)

type ExceptionFSMState int

const (
	StateIdle ExceptionFSMState = iota
	StateStarted
	StateStackTrace
)

var (
	startPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(Exception|Error|Traceback|Caused by)`),
		regexp.MustCompile(`(?i)^panic:`),
		regexp.MustCompile(`(?i)^UnhandledPromiseRejectionWarning:`),
		regexp.MustCompile(`(?i)^Uncaught (Exception|Error)`),
		regexp.MustCompile(`(?i)^fatal error:`),
	}

	stackPatterns = []*regexp.Regexp{
		regexp.MustCompile(`^\s*at\s+.+?\(.*\)`),               // Java, Node.js
		regexp.MustCompile(`^\s*File\s+".+", line \d+, in .+`), // Python
		regexp.MustCompile(`^\s+from\s.+`),                     // Python import chain
		regexp.MustCompile(`^\s+.+\.go:\d+.*$`),                // Go
		regexp.MustCompile(`^\s+goroutine\s+\d+`),              // Go goroutine start
	}
)

type GroupedLog struct {
	Message  string
	LogLevel string
}

type ExceptionGrouper struct {
	state        ExceptionFSMState
	currentGroup []string
}

func NewExceptionGrouper() *ExceptionGrouper {
	return &ExceptionGrouper{
		state: StateIdle,
	}
}

func (g *ExceptionGrouper) ProcessLine(line string) (*GroupedLog, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil, false
	}

	switch g.state {
	case StateIdle:
		if isStartPattern(trimmed) {
			g.currentGroup = []string{trimmed}
			g.state = StateStarted
			return nil, false
		}
		return &GroupedLog{Message: trimmed, LogLevel: ""}, true

	case StateStarted:
		if isStackPattern(trimmed) {
			g.currentGroup = append(g.currentGroup, trimmed)
			g.state = StateStackTrace
			return nil, false
		}
		complete := strings.Join(g.currentGroup, "\n")
		g.reset()
		return &GroupedLog{Message: complete + "\n" + trimmed, LogLevel: "ERROR"}, true

	case StateStackTrace:
		if isStackPattern(trimmed) {
			g.currentGroup = append(g.currentGroup, trimmed)
			return nil, false
		}
		complete := strings.Join(g.currentGroup, "\n")
		g.reset()
		return &GroupedLog{Message: complete + "\n" + trimmed, LogLevel: "ERROR"}, true
	}

	return nil, false
}

func (g *ExceptionGrouper) Flush() (*GroupedLog, bool) {
	if len(g.currentGroup) > 0 {
		complete := strings.Join(g.currentGroup, "\n")
		g.reset()
		return &GroupedLog{Message: complete, LogLevel: "ERROR"}, true
	}
	return nil, false
}

func (g *ExceptionGrouper) reset() {
	g.state = StateIdle
	g.currentGroup = nil
}

func isStartPattern(line string) bool {
	for _, regex := range startPatterns {
		if regex.MatchString(line) {
			return true
		}
	}
	return false
}

func isStackPattern(line string) bool {
	for _, regex := range stackPatterns {
		if regex.MatchString(line) {
			return true
		}
	}
	return false
}
