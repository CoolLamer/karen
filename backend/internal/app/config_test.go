package app

import (
	"os"
	"testing"
)

func TestGetenv(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		defValue string
		want     string
	}{
		{
			name:     "env set",
			envKey:   "TEST_ENV_VAR",
			envValue: "custom_value",
			defValue: "default",
			want:     "custom_value",
		},
		{
			name:     "env not set",
			envKey:   "TEST_ENV_VAR_NOTSET",
			envValue: "",
			defValue: "default",
			want:     "default",
		},
		{
			name:     "empty default",
			envKey:   "TEST_ENV_VAR_EMPTY",
			envValue: "",
			defValue: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			got := getenv(tt.envKey, tt.defValue)
			if got != tt.want {
				t.Errorf("getenv(%q, %q) = %q, want %q", tt.envKey, tt.defValue, got, tt.want)
			}
		})
	}
}

func TestGetenvIntClamped(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		def      int
		min      int
		max      int
		want     int
	}{
		{
			name:     "value within range",
			envKey:   "TEST_INT_NORMAL",
			envValue: "500",
			def:      100,
			min:      0,
			max:      1000,
			want:     500,
		},
		{
			name:     "value below min - clamp to min",
			envKey:   "TEST_INT_LOW",
			envValue: "-100",
			def:      100,
			min:      0,
			max:      1000,
			want:     0,
		},
		{
			name:     "value above max - clamp to max",
			envKey:   "TEST_INT_HIGH",
			envValue: "2000",
			def:      100,
			min:      0,
			max:      1000,
			want:     1000,
		},
		{
			name:     "env not set - use default",
			envKey:   "TEST_INT_NOTSET",
			envValue: "",
			def:      100,
			min:      0,
			max:      1000,
			want:     100,
		},
		{
			name:     "invalid value - use default",
			envKey:   "TEST_INT_INVALID",
			envValue: "not_a_number",
			def:      100,
			min:      0,
			max:      1000,
			want:     100,
		},
		{
			name:     "boundary: exactly min",
			envKey:   "TEST_INT_MIN",
			envValue: "200",
			def:      500,
			min:      200,
			max:      800,
			want:     200,
		},
		{
			name:     "boundary: exactly max",
			envKey:   "TEST_INT_MAX",
			envValue: "800",
			def:      500,
			min:      200,
			max:      800,
			want:     800,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			got := getenvIntClamped(tt.envKey, tt.def, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("getenvIntClamped(%q, %d, %d, %d) = %d, want %d",
					tt.envKey, tt.def, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

func TestGetenvFloatClamped(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		def      float64
		min      float64
		max      float64
		want     float64
	}{
		{
			name:     "value within range",
			envKey:   "TEST_FLOAT_NORMAL",
			envValue: "0.5",
			def:      0.3,
			min:      0.0,
			max:      1.0,
			want:     0.5,
		},
		{
			name:     "value below min - clamp to min",
			envKey:   "TEST_FLOAT_LOW",
			envValue: "-0.5",
			def:      0.3,
			min:      0.0,
			max:      1.0,
			want:     0.0,
		},
		{
			name:     "value above max - clamp to max",
			envKey:   "TEST_FLOAT_HIGH",
			envValue: "1.5",
			def:      0.3,
			min:      0.0,
			max:      1.0,
			want:     1.0,
		},
		{
			name:     "env not set - use default",
			envKey:   "TEST_FLOAT_NOTSET",
			envValue: "",
			def:      0.75,
			min:      0.0,
			max:      1.0,
			want:     0.75,
		},
		{
			name:     "invalid value - use default",
			envKey:   "TEST_FLOAT_INVALID",
			envValue: "not_a_float",
			def:      0.5,
			min:      0.0,
			max:      1.0,
			want:     0.5,
		},
		{
			name:     "boundary: exactly min",
			envKey:   "TEST_FLOAT_MIN",
			envValue: "0.0",
			def:      0.5,
			min:      0.0,
			max:      1.0,
			want:     0.0,
		},
		{
			name:     "boundary: exactly max",
			envKey:   "TEST_FLOAT_MAX",
			envValue: "1.0",
			def:      0.5,
			min:      0.0,
			max:      1.0,
			want:     1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			got := getenvFloatClamped(tt.envKey, tt.def, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("getenvFloatClamped(%q, %f, %f, %f) = %f, want %f",
					tt.envKey, tt.def, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

func TestParseAdminPhones(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single phone",
			input: "+420777123456",
			want:  []string{"+420777123456"},
		},
		{
			name:  "multiple phones",
			input: "+420777123456,+420777654321",
			want:  []string{"+420777123456", "+420777654321"},
		},
		{
			name:  "phones with spaces",
			input: "+420777123456, +420777654321, +420777111222",
			want:  []string{"+420777123456", "+420777654321", "+420777111222"},
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "phones with extra whitespace",
			input: "  +420777123456  ,  +420777654321  ",
			want:  []string{"+420777123456", "+420777654321"},
		},
		{
			name:  "trailing comma",
			input: "+420777123456,",
			want:  []string{"+420777123456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAdminPhones(tt.input)

			if len(got) != len(tt.want) {
				t.Errorf("parseAdminPhones(%q) returned %d phones, want %d", tt.input, len(got), len(tt.want))
				return
			}

			for i, phone := range got {
				if phone != tt.want[i] {
					t.Errorf("parseAdminPhones(%q)[%d] = %q, want %q", tt.input, i, phone, tt.want[i])
				}
			}
		})
	}
}

func TestLoadConfigFromEnvDefaults(t *testing.T) {
	// Clear any existing env vars that might interfere
	keysToClean := []string{
		"HTTP_ADDR", "PUBLIC_BASE_URL", "DATABASE_URL", "LOG_LEVEL",
		"STT_ENDPOINTING_MS", "STT_UTTERANCE_END_MS",
		"TTS_STABILITY", "TTS_SIMILARITY",
	}
	for _, key := range keysToClean {
		os.Unsetenv(key)
	}

	cfg := LoadConfigFromEnv()

	// Test default values
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":8080")
	}

	if cfg.PublicBaseURL != "http://localhost:8080" {
		t.Errorf("PublicBaseURL = %q, want %q", cfg.PublicBaseURL, "http://localhost:8080")
	}

	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}

	// STT defaults
	if cfg.STTEndpointingMs != 800 {
		t.Errorf("STTEndpointingMs = %d, want %d", cfg.STTEndpointingMs, 800)
	}

	if cfg.STTUtteranceEndMs != 1000 {
		t.Errorf("STTUtteranceEndMs = %d, want %d", cfg.STTUtteranceEndMs, 1000)
	}

	// TTS defaults
	if cfg.TTSStability != 0.5 {
		t.Errorf("TTSStability = %f, want %f", cfg.TTSStability, 0.5)
	}

	if cfg.TTSSimilarity != 0.75 {
		t.Errorf("TTSSimilarity = %f, want %f", cfg.TTSSimilarity, 0.75)
	}
}

func TestLoadConfigFromEnvCustomValues(t *testing.T) {
	// Set custom values
	os.Setenv("HTTP_ADDR", ":9090")
	os.Setenv("PUBLIC_BASE_URL", "https://api.example.com")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("STT_ENDPOINTING_MS", "1200")
	os.Setenv("STT_UTTERANCE_END_MS", "1500")
	os.Setenv("TTS_STABILITY", "0.7")
	os.Setenv("TTS_SIMILARITY", "0.85")
	os.Setenv("ADMIN_PHONES", "+420777123456,+420777654321")

	defer func() {
		os.Unsetenv("HTTP_ADDR")
		os.Unsetenv("PUBLIC_BASE_URL")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("STT_ENDPOINTING_MS")
		os.Unsetenv("STT_UTTERANCE_END_MS")
		os.Unsetenv("TTS_STABILITY")
		os.Unsetenv("TTS_SIMILARITY")
		os.Unsetenv("ADMIN_PHONES")
	}()

	cfg := LoadConfigFromEnv()

	if cfg.HTTPAddr != ":9090" {
		t.Errorf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":9090")
	}

	if cfg.PublicBaseURL != "https://api.example.com" {
		t.Errorf("PublicBaseURL = %q, want %q", cfg.PublicBaseURL, "https://api.example.com")
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}

	if cfg.STTEndpointingMs != 1200 {
		t.Errorf("STTEndpointingMs = %d, want %d", cfg.STTEndpointingMs, 1200)
	}

	if cfg.STTUtteranceEndMs != 1500 {
		t.Errorf("STTUtteranceEndMs = %d, want %d", cfg.STTUtteranceEndMs, 1500)
	}

	if cfg.TTSStability != 0.7 {
		t.Errorf("TTSStability = %f, want %f", cfg.TTSStability, 0.7)
	}

	if cfg.TTSSimilarity != 0.85 {
		t.Errorf("TTSSimilarity = %f, want %f", cfg.TTSSimilarity, 0.85)
	}

	if len(cfg.AdminPhones) != 2 {
		t.Errorf("AdminPhones length = %d, want 2", len(cfg.AdminPhones))
	}
}
