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

	// Limpa quaisquer filtros antigos — sem AddMatch aqui, para ler tudo
	j.FlushMatches()
	log.Println("[DEBUG] Nenhum filtro de transporte aplicado (captura tudo)")

	// Posiciona no fim do journal
	if err := j.SeekTail(); err != nil {
		log.Fatalf("[ERROR] Falha ao executar SeekTail: %v", err)
	}
	log.Println("[DEBUG] Executado SeekTail com sucesso")

	// Posiciona no último registro (tail aponta após o último)
	if _, err := j.Previous(); err != nil {
		log.Fatalf("[ERROR] Falha ao executar Previous após SeekTail: %v", err)
	}
	log.Println("[DEBUG] Executado Previous para posicionar no último registro")

	log.Println("[INFO] Journald streaming started...")

	for {
		select {
		case <-stopChan:
			log.Println("[INFO] Journald stream stopped.")
			return
		default:
		}

		log.Println("[DEBUG] Esperando nova entrada...")
		switch ev := j.Wait(time.Second); ev {
		case sdjournal.SD_JOURNAL_APPEND, sdjournal.SD_JOURNAL_INVALIDATE:
			// Há novos registros no journal
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

func getOrDefault(fields map[string]string, key, fallback string) string {
	if v, ok := fields[key]; ok {
		return v
	}
	return fallback
}
