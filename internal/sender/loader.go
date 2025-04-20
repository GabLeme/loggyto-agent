package sender

import "os"

func LoadConfigFromEnv() Config {
	return Config{
		Endpoint:  os.Getenv("LOGGYTO_ENDPOINT"),
		APIKey:    os.Getenv("LOGGYTO_API_KEY"),
		APISecret: os.Getenv("LOGGYTO_API_SECRET"),
	}
}
