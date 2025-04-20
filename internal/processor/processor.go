package processor

import (
	"log"

	"log-agent/internal/logentry"
	"log-agent/internal/outputs"
	"log-agent/internal/sender"

	"github.com/google/uuid"
)

type LogProcessor struct {
	output   outputs.Output
	groupers map[string]*ExceptionGrouper
	sender   *sender.Sender
}

func NewLogProcessor(s *sender.Sender, output outputs.Output) *LogProcessor {
	return &LogProcessor{
		output:   output,
		groupers: make(map[string]*ExceptionGrouper),
		sender:   s,
	}
}

func (p *LogProcessor) ProcessLog(source, logData string, metadata map[string]string) {
	containerID := metadata["container_id"]

	grouper, ok := p.groupers[containerID]
	if !ok {
		grouper = NewExceptionGrouper()
		p.groupers[containerID] = grouper
	}

	grouped, ready := grouper.ProcessLine(logData)
	if !ready || grouped == nil {
		return
	}

	cleanedMessage := cleanLogMessage(grouped.Message)
	logLevel := grouped.LogLevel
	if logLevel == "" {
		logLevel = detectLogLevel(cleanedMessage)
	}
	timestamp, _, inferred := TryExtractTimestamp(cleanedMessage)

	entry := logentry.LogEntry{
		Message:           cleanedMessage,
		Timestamp:         timestamp,
		Level:             logLevel,
		MessageId:         uuid.New().String(),
		Labels:            metadata,
		TimestampInferred: inferred,
	}

	if err := p.sender.Send(entry); err != nil {
		log.Printf("[ERROR] Failed to send log entry: %v | entry=%+v", err, entry)
	}
}

func (p *LogProcessor) Flush(containerID string) (*GroupedLog, bool) {
	if grouper, ok := p.groupers[containerID]; ok {
		return grouper.Flush()
	}
	return nil, false
}
