package journald

import (
	"log"
	"time"

	"log-agent/pkg/processor"

	"github.com/coreos/go-systemd/v22/sdjournal"
)

func StartJournalStream(logger *processor.LogProcessor, stopChan chan struct{}) {
	log.Println("[DEBUG] Iniciando abertura do journal...")

	j, err := sdjournal.NewJournal()
	if err != nil {
		log.Fatalf("[ERROR] Falha ao abrir o journald: %v", err)
	}
	defer j.Close()

	log.Println("[DEBUG] Journal aberto com sucesso")

	if err := j.SeekTail(); err != nil {
		log.Fatalf("[ERROR] Falha ao executar SeekTail: %v", err)
	}

	log.Println("[DEBUG] Executado SeekTail com sucesso")

	j.Next() // move para além da última entrada
	log.Println("[INFO] Journald streaming started...")

	for {
		select {
		case <-stopChan:
			log.Println("[INFO] Journald stream stopped.")
			return
		default:
			r := j.Wait(time.Second)
			if r == sdjournal.SD_JOURNAL_NOP {
				continue
			}

			n, err := j.Next()
			if err != nil {
				log.Printf("[ERROR] Falha ao chamar Next(): %v", err)
				continue
			}
			if n == 0 {
				continue
			}

			entry, err := j.GetEntry()
			if err != nil {
				log.Printf("[ERROR] Falha ao obter entrada do journal: %v", err)
				continue
			}

			logMessage := entry.Fields["MESSAGE"]
			if logMessage == "" {
				continue
			}

			// Campos opcionais defensivos
			prio := getOrDefault(entry.Fields, "PRIORITY", "unknown")
			source := getOrDefault(entry.Fields, "SYSLOG_IDENTIFIER", "unknown")
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

			log.Printf("[DEBUG] New journal entry from '%s': %s", source, logMessage)

			logger.ProcessLog(source, logMessage, metadata)
		}
	}
}

// getOrDefault retorna o valor de um campo ou um padrão caso não exista
func getOrDefault(fields map[string]string, key string, fallback string) string {
	if val, ok := fields[key]; ok {
		return val
	}
	return fallback
}
