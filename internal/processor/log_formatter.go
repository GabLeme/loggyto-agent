package processor

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ansiRegex           = regexp.MustCompile(`\x1b\[[0-9;]*m|\[\d+m`)
	controlCharRegex    = regexp.MustCompile(`[\x00-\x1F\x7F]+`)
	escapedUnicodeRegex = regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)
	newlineTabRegex     = regexp.MustCompile(`[\r\n\t\\]+`)
	multiSpaceRegex     = regexp.MustCompile(`\s{2,}`)
)

func cleanLogMessage(logData string) string {
	logData = ansiRegex.ReplaceAllString(logData, "")
	logData = controlCharRegex.ReplaceAllString(logData, " ")
	logData = newlineTabRegex.ReplaceAllString(logData, " ")
	logData = escapedUnicodeRegex.ReplaceAllStringFunc(logData, decodeUnicodeEscapeOrDrop)
	logData = removeNonPrintable(logData)
	logData = multiSpaceRegex.ReplaceAllString(logData, " ")
	return strings.TrimSpace(logData)
}

func decodeUnicodeEscapeOrDrop(s string) string {
	r, err := decodeUnicodeEscape(s)
	if err != nil {
		return ""
	}
	return r
}

func decodeUnicodeEscape(s string) (string, error) {
	var r rune
	_, err := fmt.Sscanf(s, `\u%04x`, &r)
	if err != nil || !utf8.ValidRune(r) {
		return "", err
	}
	return string(r), nil
}

func removeNonPrintable(s string) string {
	var b strings.Builder
	b.Grow(len(s)) // performance: evita múltiplas realocações
	for _, r := range s {
		if unicode.IsPrint(r) && !isEmoji(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isEmoji(r rune) bool {
	switch {
	case r >= 0x1F600 && r <= 0x1F64F: // emoticons
	case r >= 0x1F300 && r <= 0x1F5FF: // symbols & pictographs
	case r >= 0x1F680 && r <= 0x1F6FF: // transport & map
	case r >= 0x2600 && r <= 0x26FF: // misc symbols
	case r >= 0x2700 && r <= 0x27BF: // dingbats
	default:
		return false
	}
	return true
}
