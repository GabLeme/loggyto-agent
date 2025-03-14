package processor

type LogEntry struct {
	Timestamp     string            `json:"timestamp"`
	Source        string            `json:"source"`
	ContainerID   string            `json:"container_id,omitempty"`
	ContainerName string            `json:"container_name,omitempty"`
	Image         string            `json:"image,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	PodName       string            `json:"pod_name,omitempty"`
	Namespace     string            `json:"namespace,omitempty"`
	HostName      string            `json:"host_name"`
	MachineIP     string            `json:"machine_ip"`
	OS            string            `json:"os"`
	Architecture  string            `json:"architecture"`
	Level         string            `json:"level"`
	Message       interface{}       `json:"message"`
	EventType     string            `json:"event_type,omitempty"`
	TraceID       string            `json:"trace_id,omitempty"`
	SpanID        string            `json:"span_id,omitempty"`
	Tags          map[string]string `json:"tags,omitempty"`
}
