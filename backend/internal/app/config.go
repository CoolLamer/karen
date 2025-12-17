package app

import (
	"os"
)

type Config struct {
	HTTPAddr      string
	PublicBaseURL string
	DatabaseURL   string
	TwilioAuthTok string
	LogLevel      string
}

func LoadConfigFromEnv() Config {
	return Config{
		HTTPAddr:      getenv("HTTP_ADDR", ":8080"),
		PublicBaseURL: getenv("PUBLIC_BASE_URL", "http://localhost:8080"),
		DatabaseURL:   getenv("DATABASE_URL", ""),
		TwilioAuthTok: getenv("TWILIO_AUTH_TOKEN", ""),
		LogLevel:      getenv("LOG_LEVEL", "info"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}


