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
	multiSpaceRegex     = regexp.MustCompile(`\s{2,}`)
	controlCharRegex    = regexp.MustCompile(`[\x00-\x1F\x7F]+`)
	escapedUnicodeRegex = regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)
	escapedNewlineRegex = regexp.MustCompile(`(\\n|\\r|\r|\n|\t)+`)
)

func cleanLogMessage(logData string) string {
	logData = ansiRegex.ReplaceAllString(logData, "")
	logData = controlCharRegex.ReplaceAllString(logData, " ")
	logData = escapedNewlineRegex.ReplaceAllString(logData, " ")
	logData = escapedUnicodeRegex.ReplaceAllStringFunc(logData, func(s string) string {
		r, err := decodeUnicodeEscape(s)
		if err != nil {
			return ""
		}
		return r
	})
	logData = removeNonPrintable(logData)
	logData = multiSpaceRegex.ReplaceAllString(logData, " ")
	return strings.TrimSpace(logData)
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
	for _, r := range s {
		if unicode.IsPrint(r) && !isEmoji(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isEmoji(r rune) bool {
	return (r >= 0x1F600 && r <= 0x1F64F) || // emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) || // misc symbols
		(r >= 0x1F680 && r <= 0x1F6FF) || // transport & map symbols
		(r >= 0x2600 && r <= 0x26FF) || // miscellaneous symbols
		(r >= 0x2700 && r <= 0x27BF) // dingbats
}
