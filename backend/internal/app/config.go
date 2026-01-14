package app

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr      string
	PublicBaseURL string
	DatabaseURL   string
	TwilioAuthTok string
	LogLevel      string

	// Error monitoring
	SentryDSN string

	// Voice AI providers
	DeepgramAPIKey   string
	OpenAIAPIKey     string
	ElevenLabsAPIKey string

	// STT settings
	STTEndpointingMs  int // Deepgram endpointing in ms (silence threshold)
	STTUtteranceEndMs int // Hard timeout after last speech, regardless of noise

	// Voice settings (defaults, overridden by tenant config)
	GreetingText  string
	TTSVoiceID    string  // ElevenLabs voice ID
	TTSStability  float64 // ElevenLabs voice stability (0.0-1.0, default 0.5)
	TTSSimilarity float64 // ElevenLabs voice similarity boost (0.0-1.0, default 0.75)

	// Twilio Verify (SMS OTP)
	TwilioAccountSID      string
	TwilioVerifyServiceID string

	// JWT Authentication
	JWTSecret string
	JWTExpiry time.Duration

	// Admin access
	AdminPhones []string

	// Notifications
	DiscordWebhookURL string
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

		// Error monitoring
		SentryDSN: os.Getenv("SENTRY_DSN"),

		// Voice AI providers
		DeepgramAPIKey:   getenv("DEEPGRAM_API_KEY", ""),
		OpenAIAPIKey:     getenv("OPENAI_API_KEY", ""),
		ElevenLabsAPIKey: getenv("ELEVENLABS_API_KEY", ""),

		// STT settings
		// Deepgram endpointing controls how quickly we decide the caller finished speaking.
		// Too low -> fragmented utterances and interruptive back-and-forth; too high -> sluggish turns.
		STTEndpointingMs: getenvIntClamped("STT_ENDPOINTING_MS", 800, 200, 4000),
		// Utterance end is a hard timeout after last speech detected, regardless of background noise.
		// This prevents noisy environments from keeping turns open indefinitely.
		// Default 1000ms for faster response; client-side 4s max timeout provides safety net.
		STTUtteranceEndMs: getenvIntClamped("STT_UTTERANCE_END_MS", 1000, 500, 5000),

		// Voice settings (defaults, overridden by tenant config)
		GreetingText:  getenv("GREETING_TEXT", "Dobrý den, tady asistentka Karen. Majitel telefonu teď nemůže přijmout hovor, ale můžu vám pro něj zanechat vzkaz - co od něj potřebujete?"),
		TTSVoiceID:    getenv("TTS_VOICE_ID", ""),                           // ElevenLabs voice ID
		TTSStability:  getenvFloatClamped("TTS_STABILITY", 0.5, 0.0, 1.0),   // Voice stability (0.0-1.0)
		TTSSimilarity: getenvFloatClamped("TTS_SIMILARITY", 0.75, 0.0, 1.0), // Voice similarity boost (0.0-1.0)

		// Twilio Verify (SMS OTP)
		TwilioAccountSID:      getenv("TWILIO_ACCOUNT_SID", ""),
		TwilioVerifyServiceID: getenv("TWILIO_VERIFY_SERVICE_SID", ""),

		// JWT Authentication
		JWTSecret: os.Getenv("JWT_SECRET"), // Required - no fallback for security
		JWTExpiry: jwtExpiry,

		// Admin access
		AdminPhones: parseAdminPhones(os.Getenv("ADMIN_PHONES")),

		// Notifications
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
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

// getenvFloatClamped reads a float env var and clamps it to [min, max] range.
// Returns def if the env var is not set or cannot be parsed.
func getenvFloatClamped(k string, def, min, max float64) float64 {
	if v := os.Getenv(k); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			if f < min {
				return min
			}
			if f > max {
				return max
			}
			return f
		}
	}
	return def
}

// getenvIntClamped reads an int env var and clamps it to [min, max] range.
// Returns def if the env var is not set or cannot be parsed.
func getenvIntClamped(k string, def, min, max int) int {
	if v := os.Getenv(k); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			if i < min {
				return min
			}
			if i > max {
				return max
			}
			return i
		}
	}
	return def
}
