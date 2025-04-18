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

	// Limpa quaisquer filtros implícitos
	j.FlushMatches()

	// Captura todas as entradas que chegam via journald (inclui syslog)
	if err := j.AddMatch("_TRANSPORT=journal"); err != nil {
		log.Fatalf("[ERROR] Falha ao adicionar filtro journal: %v", err)
	}
	log.Println("[DEBUG] Filtro de transporte journal aplicado com sucesso")

	// Move o cursor para o final para só pegar novas entradas
	if err := j.SeekTail(); err != nil {
		log.Fatalf("[ERROR] Falha ao executar SeekTail: %v", err)
	}
	log.Println("[DEBUG] Executado SeekTail com sucesso")

	// Avança cursor para a próxima entrada
	if _, err := j.Next(); err != nil {
		log.Fatalf("[ERROR] Falha ao mover cursor após SeekTail: %v", err)
	}
	log.Println("[DEBUG] Cursor avançado após SeekTail")

	log.Println("[INFO] Journald streaming started...")

	for {
		select {
		case <-stopChan:
			log.Println("[INFO] Journald stream stopped.")
			return
		default:
			log.Println("[DEBUG] Esperando nova entrada...")
			switch r := j.Wait(time.Second); r {
			case sdjournal.SD_JOURNAL_NOP:
				continue
			case sdjournal.SD_JOURNAL_APPEND, sdjournal.SD_JOURNAL_INVALIDATE:
				// Há algo novo para ler
			default:
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

			log.Printf("[DEBUG] Entrada bruta: %+v", entry)

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

			log.Printf("[DEBUG] New journal entry from '%s': %s", source, msg)
			logger.ProcessLog(source, msg, metadata)
		}
	}
}

func getOrDefault(fields map[string]string, key string, fallback string) string {
	if val, ok := fields[key]; ok {
		return val
	}
	return fallback
}
