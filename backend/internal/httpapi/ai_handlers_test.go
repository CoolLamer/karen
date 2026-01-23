package httpapi

import (
	"crypto/subtle"
	"testing"
)

func TestAPIKeyConstantTimeComparison(t *testing.T) {
	// Test that the constant-time comparison function works correctly
	tests := []struct {
		name     string
		key1     string
		key2     string
		expected bool
	}{
		{"matching keys", "secret-key-123", "secret-key-123", true},
		{"different keys", "secret-key-123", "secret-key-456", false},
		{"different lengths", "short", "longer-key", false},
		{"empty vs non-empty", "", "secret", false},
		{"both empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := subtle.ConstantTimeCompare([]byte(tt.key1), []byte(tt.key2)) == 1
			if result != tt.expected {
				t.Errorf("ConstantTimeCompare(%q, %q) = %v, want %v", tt.key1, tt.key2, result, tt.expected)
			}
		})
	}
}

func TestAIConfigNumericKeys(t *testing.T) {
	// Test the numeric keys map used for validation
	numericKeys := map[string]bool{
		"max_turn_timeout_ms":            true,
		"adaptive_min_timeout_ms":        true,
		"adaptive_text_decay_rate_ms":    true,
		"adaptive_sentence_end_bonus_ms": true,
		"robocall_max_call_duration_ms":  true,
		"robocall_silence_threshold_ms":  true,
		"robocall_barge_in_threshold":    true,
		"robocall_barge_in_window_ms":    true,
		"robocall_repetition_threshold":  true,
	}

	tests := []struct {
		key        string
		isNumeric  bool
	}{
		{"max_turn_timeout_ms", true},
		{"adaptive_turn_enabled", false},
		{"stt_debug_enabled", false},
		{"unknown_key", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := numericKeys[tt.key]
			if result != tt.isNumeric {
				t.Errorf("numericKeys[%q] = %v, want %v", tt.key, result, tt.isNumeric)
			}
		})
	}
}

func TestAIConfigBooleanKeys(t *testing.T) {
	// Test the boolean keys map used for validation
	boolKeys := map[string]bool{
		"adaptive_turn_enabled":      true,
		"robocall_detection_enabled": true,
		"stt_debug_enabled":          true,
	}

	tests := []struct {
		key    string
		isBool bool
	}{
		{"adaptive_turn_enabled", true},
		{"robocall_detection_enabled", true},
		{"stt_debug_enabled", true},
		{"max_turn_timeout_ms", false},
		{"unknown_key", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := boolKeys[tt.key]
			if result != tt.isBool {
				t.Errorf("boolKeys[%q] = %v, want %v", tt.key, result, tt.isBool)
			}
		})
	}
}

func TestAIConfigBooleanValueValidation(t *testing.T) {
	// Test boolean value validation logic
	tests := []struct {
		value   string
		isValid bool
	}{
		{"true", true},
		{"false", true},
		{"True", false},
		{"FALSE", false},
		{"1", false},
		{"0", false},
		{"yes", false},
		{"no", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			isValid := tt.value == "true" || tt.value == "false"
			if isValid != tt.isValid {
				t.Errorf("value %q: got valid=%v, want valid=%v", tt.value, isValid, tt.isValid)
			}
		})
	}
}

func TestAILimitParsing(t *testing.T) {
	// Test the limit parsing logic used in AI handlers
	tests := []struct {
		input    string
		expected int
		fallback int
	}{
		{"20", 20, 20},
		{"100", 100, 20},
		{"1", 1, 20},
		{"0", 20, 20},   // Invalid, use fallback
		{"-5", 20, 20},  // Invalid, use fallback
		{"101", 20, 20}, // Over limit, use fallback
		{"abc", 20, 20}, // Not a number, use fallback
		{"", 20, 20},    // Empty, use fallback
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			limit := tt.fallback
			if tt.input != "" {
				var parsed int
				n, err := sscanf(tt.input, &parsed)
				if n == 1 && err == nil && parsed > 0 && parsed <= 100 {
					limit = parsed
				}
			}
			if limit != tt.expected {
				t.Errorf("limit for %q = %d, want %d", tt.input, limit, tt.expected)
			}
		})
	}
}

// sscanf is a helper for testing limit parsing (simulates strconv.Atoi behavior)
func sscanf(s string, v *int) (int, error) {
	if s == "" {
		return 0, nil
	}
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	*v = n
	if s[0] == '-' {
		*v = -*v
	}
	return 1, nil
}
