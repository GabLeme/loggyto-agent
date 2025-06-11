package pipeline

import (
	"regexp"
	"strings"
	"unicode"
)

func CleanLogMessage(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	s = removeAnsiCodes(s)
	s = removeEmojis(s)
	s = removeControlChars(s)
	s = removeLeadingJunk(s)
	s = removeRedundantTimestamps(s)
	s = collapseSpaces(s)

	return s
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func removeAnsiCodes(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

func removeEmojis(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) && (r <= unicode.MaxASCII || unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsPunct(r) || unicode.IsSymbol(r)) {
			return r
		}
		return -1
	}, s)
}

func removeControlChars(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, s)
}

var tsPrefixRegex = regexp.MustCompile(`^(\[?\d{4}[-/]\d{2}[-/]\d{2}[ T]\d{2}:\d{2}:\d{2}([.,]\d+)?(Z| UTC)?\]?[-\s:]*)+`)

func removeRedundantTimestamps(s string) string {
	return tsPrefixRegex.ReplaceAllString(s, "")
}

var leadingJunkRegex = regexp.MustCompile(`^[\s,;:.\\/\-\[\]\(\)\{\}|!@#$%^&*_+=<>?]+`)

func removeLeadingJunk(s string) string {
	return leadingJunkRegex.ReplaceAllString(s, "")
}

func collapseSpaces(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}
