package processor

import (
	"regexp"
)

func extractTags(logData string) map[string]string {
	tagPattern := regexp.MustCompile(`\[(\w+)=([^\]]+)\]`)
	matches := tagPattern.FindAllStringSubmatch(logData, -1)

	tags := make(map[string]string)
	for _, match := range matches {
		if len(match) == 3 {
			tags[match[1]] = match[2]
		}
	}

	return tags
}
