package processor

import (
	"bytes"
	"encoding/json"
	"net/http"

	"log-agent/pkg/outputs"

	"github.com/google/uuid"
)

const logEndpoint = "http://10.12.0.10:9090/logs" // Substitua pela URL real

type LogProcessor struct {
	output   outputs.Output
	groupers map[string]*ExceptionGrouper
}

func NewLogProcessor(output outputs.Output) *LogProcessor {
	return &LogProcessor{
		output:   output,
		groupers: make(map[string]*ExceptionGrouper),
	}
}

func (p *LogProcessor) ProcessLog(source, logData string, metadata map[string]string) {
	containerID := metadata["container_id"]

	// if parsed, ok := TryParseJSONLog(logData); ok {
	// 	entry := LogEntry{
	// 		Message:   parsed.Message,
	// 		Timestamp: parsed.Timestamp,
	// 		Level:     parsed.Level,
	// 		MessageId: uuid.New().String(),
	// 		Labels:    mergeMaps(metadata, parsed.Metadata),
	// 	}
	// 	logJSON, _ := json.Marshal(entry)
	// 	return
	// }

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

	entry := LogEntry{
		Message:           cleanedMessage,
		Timestamp:         timestamp,
		Level:             logLevel,
		MessageId:         uuid.New().String(),
		Labels:            metadata,
		TimestampInferred: inferred,
	}

	logJSON, _ := json.Marshal(entry)
	go sendLogToEndpoint(logJSON)
	println(string(logJSON))
}

func (p *LogProcessor) Flush(containerID string) (*GroupedLog, bool) {
	if grouper, ok := p.groupers[containerID]; ok {
		return grouper.Flush()
	}
	return nil, false
}

func sendLogToEndpoint(logData []byte) {
	resp, err := http.Post(logEndpoint, "application/json", bytes.NewBuffer(logData))
	if err != nil {
		println("err")
		return
	}
	defer resp.Body.Close()
}
