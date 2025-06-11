package processor

import (
	"log-agent/internal/pipeline"
)

type LogProcessor struct {
	pipeline *pipeline.Pipeline
}

func NewLogProcessor(p *pipeline.Pipeline) *LogProcessor {
	return &LogProcessor{pipeline: p}
}

func (lp *LogProcessor) ProcessLog(source, logData string, metadata map[string]string) {
	lp.pipeline.Process(logData, metadata)
}

func (lp *LogProcessor) Flush(containerID string) (any, bool) {
	return nil, false
}
