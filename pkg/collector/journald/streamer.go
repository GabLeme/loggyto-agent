package journald

import (
	"log"
	"time"

	"log-agent/pkg/processor"

	"github.com/coreos/go-systemd/v22/sdjournal"
)

func StartJournalStream(logger *processor.LogProcessor, stopChan chan struct{}) {
	// Abre o journal
	j, err := sdjournal.NewJournal()
	if err != nil {
		log.Fatalf("[ERROR] Failed to open journald: %v", err)
	}
	defer j.Close()

	// Vai para o final para coletar apenas logs novos
	if err := j.SeekTail(); err != nil {
		log.Fatalf("[ERROR] Failed to seek to tail: %v", err)
	}

	// Move o cursor uma entrada para frente para não pegar a última repetida
	j.Next()

	log.Println("[INFO] Journald streaming started...")

	for {
		select {
		case <-stopChan:
			log.Println("[INFO] Journald stream stopped.")
			return
		default:
			// Aguarda nova entrada com timeout de 1s
			r := j.Wait(time.Second)
			if r == sdjournal.SD_JOURNAL_NOP {
				continue // nenhum novo log, volta para esperar
			}

			n, err := j.Next()
			if err != nil {
				log.Printf("[ERROR] Failed to read next journal entry: %v", err)
				continue
			}
			if n == 0 {
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

			prio := entry.Fields["PRIORITY"]
			source := entry.Fields["SYSLOG_IDENTIFIER"]
			if source == "" {
				source = "unknown"
			}

			unit := entry.Fields["_SYSTEMD_UNIT"]
			pid := entry.Fields["_PID"]
			uid := entry.Fields["_UID"]

			metadata := map[string]string{
				"priority": prio,
				"unit":     unit,
				"pid":      pid,
				"uid":      uid,
				"journal":  "true",
			}

			// DEBUG ATIVADO
			log.Printf("[DEBUG] New journal entry from '%s': %s", source, logMessage)

			// Envia para o LogProcessor
			logger.ProcessLog(source, logMessage, metadata)
		}
	}
}
