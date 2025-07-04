package journald

import (
	"log"
	"time"

	"log-agent/internal/processor"

	"github.com/coreos/go-systemd/v22/sdjournal"
)

func StartJournalStream(logger *processor.LogProcessor, stopChan chan struct{}) {
	j, err := sdjournal.NewJournal()
	if err != nil {
		log.Fatalf("[ERROR] Failed to open journald: %v", err)
	}
	defer j.Close()

	j.FlushMatches()

	if err := j.SeekTail(); err != nil {
		log.Fatalf("[ERROR] Failed to execute SeekTail: %v", err)
	}
	if _, err := j.Previous(); err != nil {
		log.Fatalf("[ERROR] Failed to execute Previous after SeekTail: %v", err)
	}

	for {
		select {
		case <-stopChan:
			log.Println("[INFO] Journald stream stopped.")
			return
		default:
		}

		switch ev := j.Wait(time.Second); ev {
		case sdjournal.SD_JOURNAL_APPEND, sdjournal.SD_JOURNAL_INVALIDATE:
		default:
			continue
		}

		n, err := j.Next()
		if err != nil {
			log.Printf("[ERROR] Failed to call Next(): %v", err)
			continue
		}
		if n == 0 {
			continue
		}

		entry, err := j.GetEntry()
		if err != nil {
			log.Printf("[ERROR] Failed to retrieve journal entry: %v", err)
			continue
		}

		msg := entry.Fields["MESSAGE"]
		if msg == "" {
			continue
		}

		source := getOrDefault(entry.Fields, "SYSLOG_IDENTIFIER", "unknown")
		prio := getOrDefault(entry.Fields, "PRIORITY", "unknown")
		unit := getOrDefault(entry.Fields, "_SYSTEMD_UNIT", "")
		pid := getOrDefault(entry.Fields, "_PID", "")
		uid := getOrDefault(entry.Fields, "_UID", "")

		metadata := map[string]string{
			"priority": prio,
			"unit":     unit,
			"pid":      pid,
			"uid":      uid,
			"journal":  "true",
		}

		logger.ProcessLog(source, msg, metadata)
	}
}

func getOrDefault(fields map[string]string, key, fallback string) string {
	if v, ok := fields[key]; ok {
		return v
	}
	return fallback
}
