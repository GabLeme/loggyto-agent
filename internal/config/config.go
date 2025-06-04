package config

type Config struct {
	Endpoint          string
	APIKey            string
	APISecret         string
	IgnoredNamespaces []string
	IgnoredContainers []string
}
