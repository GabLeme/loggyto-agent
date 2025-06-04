package config

import (
	"os"
	"strings"
)

func LoadConfigFromEnv() Config {
	return Config{
		Endpoint:          os.Getenv("LOGGYTO_ENDPOINT"),
		APIKey:            os.Getenv("LOGGYTO_API_KEY"),
		APISecret:         os.Getenv("LOGGYTO_API_SECRET"),
		IgnoredNamespaces: parseCommaList(os.Getenv("LOGGYTO_IGNORED_NAMESPACES")),
		IgnoredContainers: parseCommaList(os.Getenv("LOGGYTO_IGNORED_CONTAINERS")),
	}
}

func parseCommaList(val string) []string {
	if val == "" {
		return []string{}
	}
	parts := strings.Split(val, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}
