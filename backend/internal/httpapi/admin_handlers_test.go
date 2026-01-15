package httpapi

import (
	"slices"
	"testing"
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
