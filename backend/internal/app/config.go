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

	// Voice AI providers
	DeepgramAPIKey   string
	OpenAIAPIKey     string
	ElevenLabsAPIKey string

	// Voice settings
	GreetingText string
	TTSVoiceID   string // ElevenLabs voice ID
}

func LoadConfigFromEnv() Config {
	return Config{
		HTTPAddr:      getenv("HTTP_ADDR", ":8080"),
		PublicBaseURL: getenv("PUBLIC_BASE_URL", "http://localhost:8080"),
		DatabaseURL:   getenv("DATABASE_URL", ""),
		TwilioAuthTok: getenv("TWILIO_AUTH_TOKEN", ""),
		LogLevel:      getenv("LOG_LEVEL", "info"),

		// Voice AI providers
		DeepgramAPIKey:   getenv("DEEPGRAM_API_KEY", ""),
		OpenAIAPIKey:     getenv("OPENAI_API_KEY", ""),
		ElevenLabsAPIKey: getenv("ELEVENLABS_API_KEY", ""),

		// Voice settings
		GreetingText: getenv("GREETING_TEXT", "Dobrý den, tady Asistentka Karen. Lukáš nemá čas, ale můžu vám pro něj zanechat vzkaz - co od něj potřebujete?"),
		TTSVoiceID:   getenv("TTS_VOICE_ID", ""), // ElevenLabs voice ID
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}


