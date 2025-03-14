package processor

import (
	"encoding/json"
	"log-agent/pkg/outputs"
)

type LogProcessor struct {
	output outputs.Output
}

func NewLogProcessor(output outputs.Output) *LogProcessor {
	return &LogProcessor{output: output}
}

func (p *LogProcessor) ProcessLog(source, logData string, metadata map[string]string) {
	entries := splitLogEntries(logData)

	for _, entryData := range entries {
		cleanedMessage := cleanLogMessage(entryData)
		parsedMessage, _ := parseJSONMessage(cleanedMessage)
		level := detectLogLevel(cleanedMessage)
		tags := extractTags(cleanedMessage)
		traceID, spanID := extractTracingInfo(cleanedMessage)
		timestamp := normalizeTimestamp(cleanedMessage)
		eventType := "APPLICATION_LOG"

		var labels map[string]string
		if rawLabels, ok := metadata["labels"]; ok {
			json.Unmarshal([]byte(rawLabels), &labels)
		}

		entry := LogEntry{
			Timestamp:     timestamp,
			Source:        source,
			Level:         level,
			Message:       parsedMessage,
			Tags:          tags,
			TraceID:       traceID,
			SpanID:        spanID,
			EventType:     eventType,
			HostName:      metadata["host_name"],
			MachineIP:     metadata["machine_ip"],
			OS:            metadata["os"],
			Architecture:  metadata["architecture"],
			ContainerID:   metadata["container_id"],
			ContainerName: metadata["container_name"],
			Image:         metadata["image"],
			Labels:        labels,
			PodName:       metadata["pod_name"],
			Namespace:     metadata["namespace"],
		}

		logJSON, err := json.Marshal(entry)
		if err != nil {
			continue
		}

		p.output.Write(string(logJSON))
	}
}
