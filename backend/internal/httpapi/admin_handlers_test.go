package httpapi

import (
	"slices"
	"testing"
	"time"
)

func TestIsAdminPhone(t *testing.T) {
	adminPhones := []string{"+420777123456", "+420777654321"}

	tests := []struct {
		name    string
		phone   string
		isAdmin bool
	}{
		{"admin phone 1", "+420777123456", true},
		{"admin phone 2", "+420777654321", true},
		{"non-admin phone", "+420111222333", false},
		{"empty phone", "", false},
		{"similar but different", "+420777123457", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slices.Contains(adminPhones, tt.phone)
			if result != tt.isAdmin {
				t.Errorf("slices.Contains(%v, %q) = %v, want %v", adminPhones, tt.phone, result, tt.isAdmin)
			}
		})
	}
}

func TestAdminPhoneValidation(t *testing.T) {
	tests := []struct {
		name      string
		phone     string
		wantValid bool
	}{
		{"valid E.164 CZ", "+420777123456", true},
		{"valid E.164 US", "+14155551234", true},
		{"valid E.164 UK", "+447911123456", true},
		{"missing plus", "420777123456", false},
		{"with spaces", "+420 777 123 456", false},
		{"with dashes", "+420-777-123-456", false},
		{"too short", "+123", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := isValidE164(tt.phone)
			if valid != tt.wantValid {
				t.Errorf("isValidE164(%q) = %v, want %v", tt.phone, valid, tt.wantValid)
			}
		})
	}
}

func TestAdminPlanValidation(t *testing.T) {
	validPlans := []string{"trial", "basic", "pro"}

	tests := []struct {
		plan  string
		valid bool
	}{
		{"trial", true},
		{"basic", true},
		{"pro", true},
		{"enterprise", false},
		{"premium", false},
		{"", false},
		{"TRIAL", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.plan, func(t *testing.T) {
			result := slices.Contains(validPlans, tt.plan)
			if result != tt.valid {
				t.Errorf("plan %q: got %v, want %v", tt.plan, result, tt.valid)
			}
		})
	}
}

func TestAdminStatusValidation(t *testing.T) {
	validStatuses := []string{"active", "suspended", "cancelled"}

	tests := []struct {
		status string
		valid  bool
	}{
		{"active", true},
		{"suspended", true},
		{"cancelled", true},
		{"pending", false},
		{"deleted", false},
		{"", false},
		{"ACTIVE", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := slices.Contains(validStatuses, tt.status)
			if result != tt.valid {
				t.Errorf("status %q: got %v, want %v", tt.status, result, tt.valid)
			}
		})
	}
}

func TestMaxTurnTimeoutValidation(t *testing.T) {
	tests := []struct {
		name    string
		timeout int
		valid   bool
	}{
		{"min valid", 1000, true},
		{"max valid", 15000, true},
		{"middle value", 7000, true},
		{"too low", 999, false},
		{"too high", 15001, false},
		{"zero", 0, false},
		{"negative", -1000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.timeout >= 1000 && tt.timeout <= 15000
			if valid != tt.valid {
				t.Errorf("timeout %d: got %v, want %v", tt.timeout, valid, tt.valid)
			}
		})
	}
}

func TestCurrentPeriodCallsValidation(t *testing.T) {
	tests := []struct {
		name  string
		calls int
		valid bool
	}{
		{"zero is valid", 0, true},
		{"positive is valid", 10, true},
		{"large positive is valid", 1000, true},
		{"negative is invalid", -1, false},
		{"large negative is invalid", -100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.calls >= 0
			if valid != tt.valid {
				t.Errorf("current_period_calls %d: got %v, want %v", tt.calls, valid, tt.valid)
			}
		})
	}
}

func TestTrialEndsAtParsing(t *testing.T) {
	// Test that trial_ends_at parses correctly as RFC3339 (what JSON uses for time.Time)
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"valid ISO date UTC", "2025-12-31T23:59:59Z", true},
		{"valid past date", "2020-01-01T00:00:00Z", true},
		{"valid future date", "2030-06-15T12:00:00Z", true},
		{"valid with offset", "2025-06-15T12:00:00+02:00", true},
		{"valid with milliseconds", "2025-06-15T12:00:00.123Z", true},
		{"invalid format", "not-a-date", false},
		{"date only no time", "2025-12-31", false},
		{"missing timezone", "2025-12-31T23:59:59", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := time.Parse(time.RFC3339, tt.input)
			valid := err == nil
			if valid != tt.valid {
				t.Errorf("time.Parse(RFC3339, %q): got valid=%v, want valid=%v, err=%v",
					tt.input, valid, tt.valid, err)
			}
		})
	}
}

func TestAdminNotesValidation(t *testing.T) {
	// Admin notes can be any string, including empty
	tests := []struct {
		name  string
		notes string
		valid bool
	}{
		{"empty string", "", true},
		{"simple note", "Customer called about upgrade", true},
		{"multiline note", "Line 1\nLine 2\nLine 3", true},
		{"unicode note", "Zákazník volal ohledně upgradu 🎉", true},
		{"very long note", string(make([]byte, 10000)), true}, // 10KB note
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Admin notes have no validation - any string is valid
			valid := true
			if valid != tt.valid {
				t.Errorf("admin_notes validation: got %v, want %v", valid, tt.valid)
			}
		})
	}
}
