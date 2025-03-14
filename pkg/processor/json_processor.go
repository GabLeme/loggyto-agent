package processor

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"log-agent/pkg/outputs"
	"log-agent/pkg/utils"
)

type LogProcessor struct {
	output   outputs.Output
	hostInfo map[string]string
}

func NewLogProcessor(output outputs.Output) *LogProcessor {
	return &LogProcessor{
		output:   output,
		hostInfo: utils.GetHostMetadata(),
	}
}

type LogEntry struct {
	Timestamp    string            `json:"timestamp"`
	Source       string            `json:"source"`
	ContainerID  string            `json:"container_id,omitempty"`
	PodName      string            `json:"pod_name,omitempty"`
	Namespace    string            `json:"namespace,omitempty"`
	HostName     string            `json:"host_name"`
	MachineIP    string            `json:"machine_ip"`
	OS           string            `json:"os"`
	Architecture string            `json:"architecture"`
	Level        string            `json:"level"`
	Message      string            `json:"message"`
	Tags         map[string]string `json:"tags,omitempty"`
}

func (p *LogProcessor) ProcessLog(source, logData string, metadata map[string]string) {
	entry := LogEntry{
		Timestamp:    time.Now().Format(time.RFC3339),
		Source:       source,
		Level:        detectLogLevel(logData),
		Message:      logData,
		Tags:         extractTags(logData),
		HostName:     p.hostInfo["host_name"],
		MachineIP:    p.hostInfo["machine_ip"],
		OS:           p.hostInfo["os"],
		Architecture: p.hostInfo["architecture"],
	}

	if val, ok := metadata["container_id"]; ok {
		entry.ContainerID = val
	}
	if val, ok := metadata["pod_name"]; ok {
		entry.PodName = val
	}
	if val, ok := metadata["namespace"]; ok {
		entry.Namespace = val
	}

	logJSON, err := json.Marshal(entry)
	if err != nil {
		return
	}

	p.output.Write(string(logJSON))
}

func detectLogLevel(logData string) string {
	logData = strings.ToLower(logData)

	switch {
	case strings.Contains(logData, "error"):
		return "ERROR"
	case strings.Contains(logData, "warn"):
		return "WARN"
	case strings.Contains(logData, "debug"):
		return "DEBUG"
	default:
		return "INFO"
	}
}

func extractTags(logData string) map[string]string {
	tagPattern := regexp.MustCompile(`\[(\w+)=(.*?)\]`)
	matches := tagPattern.FindAllStringSubmatch(logData, -1)

	tags := make(map[string]string)
	for _, match := range matches {
		if len(match) == 3 {
			tags[match[1]] = match[2]
		}
	}

	return tags
}
