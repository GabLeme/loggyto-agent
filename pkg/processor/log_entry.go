package processor

type LogEntry struct {
	Timestamp         string            `json:"timestamp"`
	Message           string            `json:"message"`
	Labels            map[string]string `json:"labels"`
	Level             string            `json:"level"`
	MessageId         string            `json:"message_id"`
	TimestampInferred bool              `json:"timestamp_inferred"`
}
