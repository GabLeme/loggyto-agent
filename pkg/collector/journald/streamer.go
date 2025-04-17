package journald

import (
	"log"
	"time"

	"log-agent/pkg/processor"

	"github.com/coreos/go-systemd/v22/sdjournal"
)

func StartJournalStream(logger *processor.LogProcessor, stopChan chan struct{}) {
	j, err := sdjournal.NewJournal()
	if err != nil {
		log.Fatalf("[ERROR] Failed to open journald: %v", err)
	}
	defer j.Close()

	// Exemplo: coletar apenas logs de unidades systemd específicas
	// j.AddMatch("_SYSTEMD_UNIT=sshd.service")

	j.SeekTail()
	j.Next()

	for {
		select {
		case <-stopChan:
			log.Println("[INFO] Journald stream stopped.")
			return
		default:
			n, err := j.Next()
			if err != nil || n == 0 {
				time.Sleep(300 * time.Millisecond)
				continue
			}

			entry, err := j.GetEntry()
			if err != nil {
				log.Printf("[ERROR] Failed to get journald entry: %v", err)
				continue
			}

			logMessage := entry.Fields["MESSAGE"]
			if logMessage == "" {
				continue
			}

			//timestamp := time.Unix(0, int64(entry.RealtimeTimestamp)*int64(time.Microsecond))
			prio := entry.Fields["PRIORITY"]
			source := entry.Fields["SYSLOG_IDENTIFIER"] // ou "_SYSTEMD_UNIT" se preferir
			if source == "" {
				source = "unknown"
			}

			metadata := map[string]string{
				"priority": prio,
				"unit":     entry.Fields["_SYSTEMD_UNIT"],
				"pid":      entry.Fields["_PID"],
				"uid":      entry.Fields["_UID"],
				"journal":  "true",
			}

			// Envia para o LogProcessor
			logger.ProcessLog(source, logMessage, metadata)
		}
	}
}
