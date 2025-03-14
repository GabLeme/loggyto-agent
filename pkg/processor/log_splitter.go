package processor

import (
	"regexp"
	"strings"
)

// splitLogEntries divide um log grande em múltiplas entradas menores corretamente
func splitLogEntries(logData string) []string {
	var entries []string

	// Expressões regulares para detectar padrões comuns de separação de logs
	timestampPattern := regexp.MustCompile(`\[\d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2} [+\-]\d{4}\]`)
	ipPattern := regexp.MustCompile(`^\d{1,3}(\.\d{1,3}){3} - - `) // Detecta IPs no início da linha (Nginx, Apache)
	jsonPattern := regexp.MustCompile(`^\{.*\}$`)                  // Detecta JSON bem formatado

	var currentEntry strings.Builder

	lines := strings.Split(logData, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		// Se detectarmos um novo timestamp, IP ou JSON no início da linha, iniciamos uma nova entrada
		if timestampPattern.MatchString(trimmedLine) || ipPattern.MatchString(trimmedLine) || jsonPattern.MatchString(trimmedLine) {
			if currentEntry.Len() > 0 {
				entries = append(entries, currentEntry.String())
				currentEntry.Reset()
			}
		}

		// Adicionamos a linha atual ao buffer
		currentEntry.WriteString(trimmedLine + "\n")
	}

	// Adiciona a última entrada, se houver
	if currentEntry.Len() > 0 {
		entries = append(entries, currentEntry.String())
	}

	return entries
}
