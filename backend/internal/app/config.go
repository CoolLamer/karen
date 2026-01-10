package app

import (
	"os"
	"strings"
	"time"
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

	// Voice settings (defaults, overridden by tenant config)
	GreetingText string
	TTSVoiceID   string // ElevenLabs voice ID

	// Twilio Verify (SMS OTP)
	TwilioAccountSID      string
	TwilioVerifyServiceID string

	// JWT Authentication
	JWTSecret string
	JWTExpiry time.Duration

	// Admin access
	AdminPhones []string
}

func LoadConfigFromEnv() Config {
	jwtExpiry, err := time.ParseDuration(getenv("JWT_EXPIRY", "24h"))
	if err != nil {
		jwtExpiry = 24 * time.Hour
	}

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

		// Voice settings (defaults, overridden by tenant config)
		GreetingText: getenv("GREETING_TEXT", "Dobrý den, tady Asistentka Karen. Lukáš nemá čas, ale můžu vám pro něj zanechat vzkaz - co od něj potřebujete?"),
		TTSVoiceID:   getenv("TTS_VOICE_ID", ""), // ElevenLabs voice ID

		// Twilio Verify (SMS OTP)
		TwilioAccountSID:      getenv("TWILIO_ACCOUNT_SID", ""),
		TwilioVerifyServiceID: getenv("TWILIO_VERIFY_SERVICE_SID", ""),

		// JWT Authentication
		JWTSecret: os.Getenv("JWT_SECRET"), // Required - no fallback for security
		JWTExpiry: jwtExpiry,

		// Admin access
		AdminPhones: parseAdminPhones(os.Getenv("ADMIN_PHONES")),
	}
}

func parseAdminPhones(s string) []string {
	if s == "" {
		return nil
	}
	var phones []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			phones = append(phones, p)
		}
	}
	return phones
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}


