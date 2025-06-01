package processor

import (
	"regexp"
	"strings"
)

type ExceptionFSMState int

const (
	StateIdle ExceptionFSMState = iota
	StateCollecting
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

// -----------------------------
// Core processing
// -----------------------------
func (g *ExceptionGrouper) ProcessLine(line string) (*GroupedLog, bool) {
	normalized := normalizeLine(line)
	trimmed := strings.TrimSpace(line)

	if trimmed == "" {
		return nil, false
	}

	switch g.state {
	case StateIdle:
		if isExceptionStart(normalized) {
			g.currentGroup = []string{trimmed}
			g.state = StateCollecting
			return nil, false
		}
		return &GroupedLog{Message: trimmed, LogLevel: ""}, true

	case StateCollecting:
		if isStackLine(normalized) || isContinuation(normalized) {
			g.currentGroup = append(g.currentGroup, trimmed)
			return nil, false
		}

		// flush current exception
		msg := strings.Join(g.currentGroup, "\n")
		g.reset()

		// trata a linha atual como próxima entrada (pode ser nova exceção ou log normal)
		if isExceptionStart(normalized) {
			g.currentGroup = []string{trimmed}
			g.state = StateCollecting
			return &GroupedLog{Message: msg, LogLevel: "ERROR"}, true
		}

		return &GroupedLog{Message: msg + "\n" + trimmed, LogLevel: "ERROR"}, true
	}

	return nil, false
}

func (g *ExceptionGrouper) Flush() (*GroupedLog, bool) {
	if len(g.currentGroup) > 0 {
		msg := strings.Join(g.currentGroup, "\n")
		g.reset()
		return &GroupedLog{Message: msg, LogLevel: "ERROR"}, true
	}
	return nil, false
}

func (g *ExceptionGrouper) reset() {
	g.state = StateIdle
	g.currentGroup = nil
}

// -----------------------------
// Exception pattern helpers
// -----------------------------

// remove prefixos comuns: timestamps, níveis de log, etc.
var prefixCleaner = regexp.MustCompile(`^\[?\w+\]?\s*\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z| UTC)?\s*[-]*\s*`)

func normalizeLine(line string) string {
	return prefixCleaner.ReplaceAllString(line, "")
}

var exceptionStartPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(Exception|Error|Traceback|Caused by)`),
	regexp.MustCompile(`(?i)^panic:`),
	regexp.MustCompile(`(?i)^UnhandledPromiseRejectionWarning:`),
	regexp.MustCompile(`(?i)^Uncaught (Exception|Error)`),
	regexp.MustCompile(`(?i)^fatal error:`),
	regexp.MustCompile(`(?i)^\w*Error:`),          // TypeError:, ReferenceError:
	regexp.MustCompile(`(?i)^\w*Exception:`),      // NullPointerException:
	regexp.MustCompile(`(?i)^org\..*Exception\b`), // org.foo.Exception
	regexp.MustCompile(`(?i)^Traceback \(most recent call last\):`),
}

var stackLinePatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\s*at\s+.+?\(.*\)`),               // Java, JS
	regexp.MustCompile(`^\s*File\s+".+", line \d+, in .+`), // Python
	regexp.MustCompile(`^\s+from\s.+`),                     // Python chain
	regexp.MustCompile(`^\s+.+\.go:\d+.*$`),                // Go
	regexp.MustCompile(`^\s+goroutine\s+\d+`),              // Go
	regexp.MustCompile(`^\s+\.\.\.`),                       // Ruby
}

// Algumas linguagens quebram a stack em mais de uma linha (ex: Go ou Ruby com "...")
func isContinuation(line string) bool {
	return strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "    ")
}

func isExceptionStart(line string) bool {
	for _, r := range exceptionStartPatterns {
		if r.MatchString(line) {
			return true
		}
	}
	return false
}

func isStackLine(line string) bool {
	for _, r := range stackLinePatterns {
		if r.MatchString(line) {
			return true
		}
	}
	return false
}
