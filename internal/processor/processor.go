package processor

import (
	"log"
	"regexp"
	"strings"
	"time"

	"log-agent/internal/logentry"
	"log-agent/internal/outputs"
	"log-agent/internal/sender"
	"log-agent/internal/utils"

	"github.com/google/uuid"
)

type LogProcessor struct {
	output       outputs.Output
	groupers     map[string]*ExceptionGrouper
	sender       *sender.Sender
	messageCache *utils.MessageCache
}

func NewLogProcessor(s *sender.Sender, output outputs.Output) *LogProcessor {
	return &LogProcessor{
		output:       output,
		groupers:     make(map[string]*ExceptionGrouper),
		sender:       s,
		messageCache: utils.NewMessageCache(15 * time.Second), // TTL configurável
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

	if isInvalidLog(cleanedMessage) {
		return
	}

	// 🔽 Rate-limit: ignora logs duplicados em um intervalo curto
	if !p.messageCache.ShouldProcess(cleanedMessage) {
		return
	}

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

func isInvalidLog(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true
	}
	match, _ := regexp.MatchString(`^[-=*_.~\\\/\s]+$`, trimmed)
	return match
}
