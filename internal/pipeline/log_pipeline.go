package pipeline

import (
	"log"
	"time"

	"github.com/google/uuid"
)

type LogEntry struct {
	Message           string            `json:"message"`
	Timestamp         time.Time         `json:"timestamp"`
	Level             string            `json:"level"`
	MessageId         string            `json:"message_id"`
	Labels            map[string]string `json:"labels"`
	TimestampInferred bool              `json:"timestamp_inferred"`
}

type Pipeline struct {
	Splitter      func(string) []string
	Formatter     func(string) string
	Deduplicator  func(string) bool
	LevelDetector func(string) string
	TimestampFunc func(string) (time.Time, bool)
	Sender        func(*LogEntry) error
}

func NewPipeline(
	splitter func(string) []string,
	formatter func(string) string,
	deduplicator func(string) bool,
	levelDetector func(string) string,
	timestampFunc func(string) (time.Time, bool),
	sender func(*LogEntry) error,
) *Pipeline {
	return &Pipeline{
		Splitter:      splitter,
		Formatter:     formatter,
		Deduplicator:  deduplicator,
		LevelDetector: levelDetector,
		TimestampFunc: timestampFunc,
		Sender:        sender,
	}
}

func (p *Pipeline) Process(raw string, metadata map[string]string) {
	lines := p.Splitter(raw)

	for _, line := range lines {
		formatted := p.Formatter(line)
		if formatted == "" {
			continue
		}

		if !p.Deduplicator(formatted) {
			continue
		}

		level := p.LevelDetector(formatted)
		ts, inferred := p.TimestampFunc(formatted)

		entry := &LogEntry{
			Message:           formatted,
			Timestamp:         ts,
			Level:             level,
			MessageId:         uuid.New().String(),
			Labels:            metadata,
			TimestampInferred: inferred,
		}

		if err := p.Sender(entry); err != nil {
			log.Printf("[ERROR] Failed to send log entry: %v | entry=%+v", err, entry)
		}
	}
}
