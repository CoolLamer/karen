// Package costs provides cost calculation for API usage.
package costs

import (
	"os"
	"strconv"
)

// Pricing constants (in cents per unit for precision).
// These are based on 2026 market rates and can be overridden via environment variables.
var (
	// TwilioCentsPerMinute is the cost per minute for Twilio inbound voice calls.
	// Default: $0.0085/min = 0.85 cents/min
	TwilioCentsPerMinute = getEnvFloat("COST_TWILIO_CENTS_PER_MIN", 0.85)

	// DeepgramCentsPerMinute is the cost per minute for Deepgram Nova-3 streaming STT.
	// Default: $0.0077/min = 0.77 cents/min
	DeepgramCentsPerMinute = getEnvFloat("COST_DEEPGRAM_CENTS_PER_MIN", 0.77)

	// OpenAICentsPerThousandInputTokens is the cost per 1K input tokens for GPT-4o-mini.
	// Default: $0.15/1M = $0.00015/1K = 0.015 cents/1K tokens
	OpenAICentsPerThousandInputTokens = getEnvFloat("COST_OPENAI_INPUT_CENTS_PER_1K", 0.015)

	// OpenAICentsPerThousandOutputTokens is the cost per 1K output tokens for GPT-4o-mini.
	// Default: $0.60/1M = $0.0006/1K = 0.06 cents/1K tokens
	OpenAICentsPerThousandOutputTokens = getEnvFloat("COST_OPENAI_OUTPUT_CENTS_PER_1K", 0.06)

	// ElevenLabsCentsPerThousandChars is the cost per 1K characters for ElevenLabs TTS.
	// Default: $0.18/1K chars = 18 cents/1K chars
	ElevenLabsCentsPerThousandChars = getEnvFloat("COST_ELEVENLABS_CENTS_PER_1K_CHARS", 18.0)

	// PhoneRentalCentsPerMonth is the monthly cost per phone number.
	// Default: $1.50/month = 150 cents/month
	PhoneRentalCentsPerMonth = getEnvInt("COST_PHONE_RENTAL_CENTS_PER_MONTH", 150)
)

// CallMetrics contains the raw metrics from a call used for cost calculation.
type CallMetrics struct {
	CallDurationSeconds int // Total call duration (for Twilio billing)
	STTDurationSeconds  int // Audio processed by STT (may differ from call duration)
	LLMInputTokens      int // Tokens sent to LLM
	LLMOutputTokens     int // Tokens received from LLM
	TTSCharacters       int // Characters sent to TTS
}

// CallCosts contains the calculated costs for a call in cents.
type CallCosts struct {
	TwilioCostCents int
	STTCostCents    int
	LLMCostCents    int
	TTSCostCents    int
	TotalCostCents  int
}

// CalculateCallCosts computes the costs for a call based on usage metrics.
func CalculateCallCosts(m CallMetrics) CallCosts {
	// Convert seconds to minutes (round up for billing)
	callMinutes := float64(m.CallDurationSeconds) / 60.0
	sttMinutes := float64(m.STTDurationSeconds) / 60.0

	// Calculate individual costs
	twilioCents := callMinutes * TwilioCentsPerMinute
	sttCents := sttMinutes * DeepgramCentsPerMinute

	// LLM costs: per 1K tokens
	llmInputCents := (float64(m.LLMInputTokens) / 1000.0) * OpenAICentsPerThousandInputTokens
	llmOutputCents := (float64(m.LLMOutputTokens) / 1000.0) * OpenAICentsPerThousandOutputTokens
	llmCents := llmInputCents + llmOutputCents

	// TTS costs: per 1K characters
	ttsCents := (float64(m.TTSCharacters) / 1000.0) * ElevenLabsCentsPerThousandChars

	// Round to nearest cent (we store as integers)
	costs := CallCosts{
		TwilioCostCents: roundToInt(twilioCents),
		STTCostCents:    roundToInt(sttCents),
		LLMCostCents:    roundToInt(llmCents),
		TTSCostCents:    roundToInt(ttsCents),
	}
	costs.TotalCostCents = costs.TwilioCostCents + costs.STTCostCents + costs.LLMCostCents + costs.TTSCostCents

	return costs
}

// CalculatePhoneRentalCost calculates the prorated phone rental cost for a given number of days.
func CalculatePhoneRentalCost(phoneCount int, daysInPeriod int) int {
	if daysInPeriod <= 0 {
		daysInPeriod = 30 // Default to 30 days
	}

	// Full monthly cost per phone
	monthlyCost := phoneCount * PhoneRentalCentsPerMonth

	// Prorate based on days (assuming 30-day month for simplicity)
	prorated := float64(monthlyCost) * (float64(daysInPeriod) / 30.0)

	return roundToInt(prorated)
}

// CalculateMonthlyPhoneRentalCost returns the full monthly cost for the given number of phones.
func CalculateMonthlyPhoneRentalCost(phoneCount int) int {
	return phoneCount * PhoneRentalCentsPerMonth
}

// roundToInt rounds a float to the nearest integer.
func roundToInt(f float64) int {
	if f < 0 {
		return int(f - 0.5)
	}
	return int(f + 0.5)
}

// getEnvFloat returns an environment variable as float64, or the default if not set.
func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

// getEnvInt returns an environment variable as int, or the default if not set.
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
