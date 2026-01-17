package costs

import (
	"testing"
)

func TestCalculateCallCosts(t *testing.T) {
	tests := []struct {
		name    string
		metrics CallMetrics
		want    CallCosts
	}{
		{
			name: "typical 2 minute call",
			metrics: CallMetrics{
				CallDurationSeconds: 120, // 2 minutes
				STTDurationSeconds:  120, // Same as call
				LLMInputTokens:      500, // Typical conversation
				LLMOutputTokens:     200, // AI responses
				TTSCharacters:       400, // Spoken response chars
			},
			// Twilio: 2 * 0.85 = 1.7 -> 2 cents
			// STT: 2 * 0.77 = 1.54 -> 2 cents
			// LLM: (500/1000)*0.015 + (200/1000)*0.06 = 0.0075 + 0.012 = 0.0195 -> 0 cents
			// TTS: (400/1000)*18 = 7.2 -> 7 cents
			// Total: 2 + 2 + 0 + 7 = 11 cents
			want: CallCosts{
				TwilioCostCents: 2,
				STTCostCents:    2,
				LLMCostCents:    0,
				TTSCostCents:    7,
				TotalCostCents:  11,
			},
		},
		{
			name: "short 30 second call",
			metrics: CallMetrics{
				CallDurationSeconds: 30,
				STTDurationSeconds:  30,
				LLMInputTokens:      100,
				LLMOutputTokens:     50,
				TTSCharacters:       100,
			},
			// Twilio: 0.5 * 0.85 = 0.425 -> 0 cents
			// STT: 0.5 * 0.77 = 0.385 -> 0 cents
			// LLM: very small -> 0 cents
			// TTS: (100/1000)*18 = 1.8 -> 2 cents
			want: CallCosts{
				TwilioCostCents: 0,
				STTCostCents:    0,
				LLMCostCents:    0,
				TTSCostCents:    2,
				TotalCostCents:  2,
			},
		},
		{
			name: "long 10 minute call with lots of conversation",
			metrics: CallMetrics{
				CallDurationSeconds: 600,  // 10 minutes
				STTDurationSeconds:  600,  // Same as call
				LLMInputTokens:      5000, // Long conversation
				LLMOutputTokens:     2000, // Detailed responses
				TTSCharacters:       4000, // Lots of spoken text
			},
			// Twilio: 10 * 0.85 = 8.5 -> 9 cents
			// STT: 10 * 0.77 = 7.7 -> 8 cents
			// LLM: (5000/1000)*0.015 + (2000/1000)*0.06 = 0.075 + 0.12 = 0.195 -> 0 cents
			// TTS: (4000/1000)*18 = 72 -> 72 cents
			// Total: 9 + 8 + 0 + 72 = 89 cents
			want: CallCosts{
				TwilioCostCents: 9,
				STTCostCents:    8,
				LLMCostCents:    0,
				TTSCostCents:    72,
				TotalCostCents:  89,
			},
		},
		{
			name: "zero duration call (edge case)",
			metrics: CallMetrics{
				CallDurationSeconds: 0,
				STTDurationSeconds:  0,
				LLMInputTokens:      0,
				LLMOutputTokens:     0,
				TTSCharacters:       0,
			},
			want: CallCosts{
				TwilioCostCents: 0,
				STTCostCents:    0,
				LLMCostCents:    0,
				TTSCostCents:    0,
				TotalCostCents:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateCallCosts(tt.metrics)
			if got.TwilioCostCents != tt.want.TwilioCostCents {
				t.Errorf("TwilioCostCents = %d, want %d", got.TwilioCostCents, tt.want.TwilioCostCents)
			}
			if got.STTCostCents != tt.want.STTCostCents {
				t.Errorf("STTCostCents = %d, want %d", got.STTCostCents, tt.want.STTCostCents)
			}
			if got.LLMCostCents != tt.want.LLMCostCents {
				t.Errorf("LLMCostCents = %d, want %d", got.LLMCostCents, tt.want.LLMCostCents)
			}
			if got.TTSCostCents != tt.want.TTSCostCents {
				t.Errorf("TTSCostCents = %d, want %d", got.TTSCostCents, tt.want.TTSCostCents)
			}
			if got.TotalCostCents != tt.want.TotalCostCents {
				t.Errorf("TotalCostCents = %d, want %d", got.TotalCostCents, tt.want.TotalCostCents)
			}
		})
	}
}

func TestCalculateMonthlyPhoneRentalCost(t *testing.T) {
	tests := []struct {
		name       string
		phoneCount int
		want       int
	}{
		{"no phones", 0, 0},
		{"one phone", 1, 150},   // $1.50
		{"two phones", 2, 300},  // $3.00
		{"five phones", 5, 750}, // $7.50
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateMonthlyPhoneRentalCost(tt.phoneCount)
			if got != tt.want {
				t.Errorf("CalculateMonthlyPhoneRentalCost(%d) = %d, want %d", tt.phoneCount, got, tt.want)
			}
		})
	}
}

func TestCalculatePhoneRentalCost_Prorated(t *testing.T) {
	tests := []struct {
		name         string
		phoneCount   int
		daysInPeriod int
		want         int
	}{
		{"full month", 1, 30, 150}, // $1.50
		{"half month", 1, 15, 75},  // $0.75
		{"10 days", 2, 10, 100},    // 2 phones * 150 * (10/30) = 100
		{"zero days defaults to 30", 1, 0, 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculatePhoneRentalCost(tt.phoneCount, tt.daysInPeriod)
			if got != tt.want {
				t.Errorf("CalculatePhoneRentalCost(%d, %d) = %d, want %d",
					tt.phoneCount, tt.daysInPeriod, got, tt.want)
			}
		})
	}
}
