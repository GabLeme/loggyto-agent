package journald

import (
	"log"
	"log-agent/internal/processor"
)

type JournaldCollector struct {
	stopChan chan struct{}
	Logger   *processor.LogProcessor
}

func NewJournaldCollector(logger *processor.LogProcessor) *JournaldCollector {
	return &JournaldCollector{
		stopChan: make(chan struct{}),
		Logger:   logger,
	}
}

func (jc *JournaldCollector) Start() {
	log.Println("[INFO] Journald Collector started...")
	go StartJournalStream(jc.Logger, jc.stopChan)
}

func (jc *JournaldCollector) Stop() {
	log.Println("[WARNING] Stopping Journald Collector...")
	close(jc.stopChan)
}
