package store

import (
	"context"
	"testing"
	"time"
)

// getTestDB is defined in store_test.go

func TestGetPlanCallLimit(t *testing.T) {
	tests := []struct {
		name     string
		plan     string
		expected int
	}{
		{"trial plan", "trial", 20},
		{"basic plan", "basic", 50},
		{"pro plan", "pro", -1},        // unlimited
		{"unknown plan", "unknown", 0}, // defaults to 0
		{"empty plan", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPlanCallLimit(tt.plan)
			if got != tt.expected {
				t.Errorf("GetPlanCallLimit(%q) = %d, want %d", tt.plan, got, tt.expected)
			}
		})
	}
}

func TestCanTenantReceiveCalls(t *testing.T) {
	now := time.Now()
	futureDate := now.Add(7 * 24 * time.Hour)
	pastDate := now.Add(-7 * 24 * time.Hour)

	tests := []struct {
		name               string
		tenant             *Tenant
		expectedCanReceive bool
		expectedReason     string
	}{
		{
			name: "trial tenant with calls remaining and time remaining",
			tenant: &Tenant{
				Plan:               "trial",
				Status:             "active",
				CurrentPeriodCalls: 10,
				TrialEndsAt:        &futureDate,
			},
			expectedCanReceive: true,
			expectedReason:     "ok",
		},
		{
			name: "trial tenant at call limit",
			tenant: &Tenant{
				Plan:               "trial",
				Status:             "active",
				CurrentPeriodCalls: 20,
				TrialEndsAt:        &futureDate,
			},
			expectedCanReceive: false,
			expectedReason:     "limit_exceeded",
		},
		{
			name: "trial tenant with expired trial",
			tenant: &Tenant{
				Plan:               "trial",
				Status:             "active",
				CurrentPeriodCalls: 5,
				TrialEndsAt:        &pastDate,
			},
			expectedCanReceive: false,
			expectedReason:     "trial_expired",
		},
		{
			name: "basic plan within limit",
			tenant: &Tenant{
				Plan:               "basic",
				Status:             "active",
				CurrentPeriodCalls: 30,
			},
			expectedCanReceive: true,
			expectedReason:     "ok",
		},
		{
			name: "basic plan at limit",
			tenant: &Tenant{
				Plan:               "basic",
				Status:             "active",
				CurrentPeriodCalls: 50,
			},
			expectedCanReceive: false,
			expectedReason:     "limit_exceeded",
		},
		{
			name: "pro plan unlimited",
			tenant: &Tenant{
				Plan:               "pro",
				Status:             "active",
				CurrentPeriodCalls: 1000,
			},
			expectedCanReceive: true,
			expectedReason:     "ok",
		},
		{
			name: "cancelled subscription",
			tenant: &Tenant{
				Plan:               "basic",
				Status:             "cancelled",
				CurrentPeriodCalls: 10,
			},
			expectedCanReceive: false,
			expectedReason:     "subscription_cancelled",
		},
		{
			name: "suspended subscription",
			tenant: &Tenant{
				Plan:               "basic",
				Status:             "suspended",
				CurrentPeriodCalls: 10,
			},
			expectedCanReceive: false,
			expectedReason:     "subscription_suspended",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := CanTenantReceiveCalls(tt.tenant)
			if status.CanReceive != tt.expectedCanReceive {
				t.Errorf("CanReceive = %v, want %v", status.CanReceive, tt.expectedCanReceive)
			}
			if status.Reason != tt.expectedReason {
				t.Errorf("Reason = %q, want %q", status.Reason, tt.expectedReason)
			}
		})
	}
}

func TestCanTenantReceiveCalls_TrialCalculations(t *testing.T) {
	now := time.Now()
	// Use 5 days + 12 hours to ensure we get 5 full days after truncation
	fiveDaysFromNow := now.Add(5*24*time.Hour + 12*time.Hour)

	tenant := &Tenant{
		Plan:               "trial",
		Status:             "active",
		CurrentPeriodCalls: 15, // 5 calls remaining
		TrialEndsAt:        &fiveDaysFromNow,
	}

	status := CanTenantReceiveCalls(tenant)

	// Allow 4-5 days due to rounding (calculation uses integer division of hours/24)
	if status.TrialDaysLeft < 4 || status.TrialDaysLeft > 5 {
		t.Errorf("TrialDaysLeft = %d, want 4-5", status.TrialDaysLeft)
	}
	if status.TrialCallsLeft != 5 {
		t.Errorf("TrialCallsLeft = %d, want 5", status.TrialCallsLeft)
	}
	if status.CallsUsed != 15 {
		t.Errorf("CallsUsed = %d, want 15", status.CallsUsed)
	}
	if status.CallsLimit != 20 {
		t.Errorf("CallsLimit = %d, want 20", status.CallsLimit)
	}
}

// Integration tests (require database)

func TestIncrementTenantUsage(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create test tenant
	tenant, err := s.CreateTenant(ctx, "Usage Test Tenant", "Test prompt", "cs")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Cleanup at the end
	defer func() {
		_, _ = db.Exec(ctx, "DELETE FROM tenant_usage WHERE tenant_id = $1", tenant.ID)
		_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
	}()

	t.Run("increment regular call", func(t *testing.T) {
		// Get initial state
		initialUsage, _ := s.GetTenantCurrentUsage(ctx, tenant.ID)
		initialCalls := 0
		initialTimeSaved := 0
		if initialUsage != nil {
			initialCalls = initialUsage.CallsCount
			initialTimeSaved = initialUsage.TimeSavedSeconds
		}

		// Increment usage: 60 seconds, not spam
		err := s.IncrementTenantUsage(ctx, tenant.ID, 60, false)
		if err != nil {
			t.Fatalf("IncrementTenantUsage failed: %v", err)
		}

		// Verify
		usage, err := s.GetTenantCurrentUsage(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenantCurrentUsage failed: %v", err)
		}

		if usage.CallsCount != initialCalls+1 {
			t.Errorf("CallsCount = %d, want %d", usage.CallsCount, initialCalls+1)
		}

		// Time saved should be call duration (60s) + overhead (120s for non-spam) = 180s
		expectedTimeSaved := initialTimeSaved + 60 + 120
		if usage.TimeSavedSeconds != expectedTimeSaved {
			t.Errorf("TimeSavedSeconds = %d, want %d", usage.TimeSavedSeconds, expectedTimeSaved)
		}
	})

	t.Run("increment spam call", func(t *testing.T) {
		initialUsage, _ := s.GetTenantCurrentUsage(ctx, tenant.ID)
		initialTimeSaved := 0
		if initialUsage != nil {
			initialTimeSaved = initialUsage.TimeSavedSeconds
		}

		// Increment usage: 30 seconds, spam
		err := s.IncrementTenantUsage(ctx, tenant.ID, 30, true)
		if err != nil {
			t.Fatalf("IncrementTenantUsage failed: %v", err)
		}

		// Verify
		usage, _ := s.GetTenantCurrentUsage(ctx, tenant.ID)

		// Time saved should be call duration (30s) + overhead (300s for spam) = 330s
		expectedTimeSaved := initialTimeSaved + 30 + 300
		if usage.TimeSavedSeconds != expectedTimeSaved {
			t.Errorf("TimeSavedSeconds = %d, want %d", usage.TimeSavedSeconds, expectedTimeSaved)
		}

		// Spam call should increment spam counter
		if usage.SpamCallsBlocked < 1 {
			t.Errorf("SpamCallsBlocked = %d, want >= 1", usage.SpamCallsBlocked)
		}
	})
}

func TestGetTenantBillingInfo(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create test tenant
	tenant, err := s.CreateTenant(ctx, "Billing Info Test", "Test prompt", "cs")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Cleanup at the end
	defer func() {
		_, _ = db.Exec(ctx, "DELETE FROM tenant_usage WHERE tenant_id = $1", tenant.ID)
		_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
	}()

	t.Run("new tenant billing info", func(t *testing.T) {
		info, err := s.GetTenantBillingInfo(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenantBillingInfo failed: %v", err)
		}

		if info.Plan != "trial" {
			t.Errorf("Plan = %q, want %q", info.Plan, "trial")
		}
		if info.Status != "active" {
			t.Errorf("Status = %q, want %q", info.Status, "active")
		}
		if info.TrialEndsAt == nil {
			t.Error("TrialEndsAt should not be nil for new tenant")
		}
		if info.StripeCustomerID != nil {
			t.Errorf("StripeCustomerID should be nil for new tenant, got %v", info.StripeCustomerID)
		}
	})
}

func TestUpdateTenantBilling(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create test tenant
	tenant, err := s.CreateTenant(ctx, "Update Billing Test", "Test prompt", "cs")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Cleanup at the end
	defer func() {
		_, _ = db.Exec(ctx, "DELETE FROM tenant_usage WHERE tenant_id = $1", tenant.ID)
		_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
	}()

	t.Run("update plan and status", func(t *testing.T) {
		err := s.UpdateTenantBilling(ctx, tenant.ID, map[string]any{
			"plan":   "basic",
			"status": "active",
		})
		if err != nil {
			t.Fatalf("UpdateTenantBilling failed: %v", err)
		}

		info, err := s.GetTenantBillingInfo(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenantBillingInfo failed: %v", err)
		}

		if info.Plan != "basic" {
			t.Errorf("Plan = %q, want %q", info.Plan, "basic")
		}
	})

	t.Run("update stripe customer id", func(t *testing.T) {
		err := s.UpdateTenantBilling(ctx, tenant.ID, map[string]any{
			"stripe_customer_id": "cus_test123",
		})
		if err != nil {
			t.Fatalf("UpdateTenantBilling failed: %v", err)
		}

		info, err := s.GetTenantBillingInfo(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenantBillingInfo failed: %v", err)
		}

		if info.StripeCustomerID == nil || *info.StripeCustomerID != "cus_test123" {
			t.Errorf("StripeCustomerID = %v, want cus_test123", info.StripeCustomerID)
		}
	})

	t.Run("update subscription id", func(t *testing.T) {
		err := s.UpdateTenantBilling(ctx, tenant.ID, map[string]any{
			"stripe_subscription_id": "sub_test456",
		})
		if err != nil {
			t.Fatalf("UpdateTenantBilling failed: %v", err)
		}

		info, err := s.GetTenantBillingInfo(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenantBillingInfo failed: %v", err)
		}

		if info.StripeSubscriptionID == nil || *info.StripeSubscriptionID != "sub_test456" {
			t.Errorf("StripeSubscriptionID = %v, want sub_test456", info.StripeSubscriptionID)
		}
	})
}

func TestGetTenantIDByStripeCustomer(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create test tenant with stripe customer ID
	tenant, err := s.CreateTenant(ctx, "Stripe Lookup Test", "Test prompt", "cs")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Set stripe customer ID
	err = s.UpdateTenantBilling(ctx, tenant.ID, map[string]any{
		"stripe_customer_id": "cus_lookup_test",
	})
	if err != nil {
		t.Fatalf("UpdateTenantBilling failed: %v", err)
	}

	// Cleanup at the end
	defer func() {
		_, _ = db.Exec(ctx, "DELETE FROM tenant_usage WHERE tenant_id = $1", tenant.ID)
		_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
	}()

	t.Run("find existing customer", func(t *testing.T) {
		foundID, err := s.GetTenantIDByStripeCustomer(ctx, "cus_lookup_test")
		if err != nil {
			t.Fatalf("GetTenantIDByStripeCustomer failed: %v", err)
		}
		if foundID != tenant.ID {
			t.Errorf("found ID = %q, want %q", foundID, tenant.ID)
		}
	})

	t.Run("not found for non-existent customer", func(t *testing.T) {
		_, err := s.GetTenantIDByStripeCustomer(ctx, "cus_nonexistent")
		if err == nil {
			t.Error("expected error for non-existent customer")
		}
	})
}

func TestResetTenantPeriodCalls(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create test tenant
	tenant, err := s.CreateTenant(ctx, "Reset Calls Test", "Test prompt", "cs")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Cleanup at the end
	defer func() {
		_, _ = db.Exec(ctx, "DELETE FROM tenant_usage WHERE tenant_id = $1", tenant.ID)
		_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
	}()

	// Set some calls
	_, err = db.Exec(ctx, "UPDATE tenants SET current_period_calls = 15 WHERE id = $1", tenant.ID)
	if err != nil {
		t.Fatalf("failed to set calls: %v", err)
	}

	// Reset
	err = s.ResetTenantPeriodCalls(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("ResetTenantPeriodCalls failed: %v", err)
	}

	// Verify
	var calls int
	err = db.QueryRow(ctx, "SELECT COALESCE(current_period_calls, 0) FROM tenants WHERE id = $1", tenant.ID).Scan(&calls)
	if err != nil {
		t.Fatalf("failed to query calls: %v", err)
	}
	if calls != 0 {
		t.Errorf("current_period_calls = %d, want 0", calls)
	}
}

func TestGetTenantTotalTimeSaved(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create test tenant
	tenant, err := s.CreateTenant(ctx, "Time Saved Test", "Test prompt", "cs")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Cleanup at the end
	defer func() {
		_, _ = db.Exec(ctx, "DELETE FROM tenant_usage WHERE tenant_id = $1", tenant.ID)
		_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
	}()

	t.Run("new tenant has zero time saved", func(t *testing.T) {
		timeSaved, err := s.GetTenantTotalTimeSaved(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenantTotalTimeSaved failed: %v", err)
		}
		if timeSaved != 0 {
			t.Errorf("timeSaved = %d, want 0", timeSaved)
		}
	})

	t.Run("time saved accumulates across calls", func(t *testing.T) {
		// Add a few calls
		_ = s.IncrementTenantUsage(ctx, tenant.ID, 60, false) // 60+120 = 180s
		_ = s.IncrementTenantUsage(ctx, tenant.ID, 30, true)  // 30+300 = 330s
		_ = s.IncrementTenantUsage(ctx, tenant.ID, 45, false) // 45+120 = 165s

		timeSaved, err := s.GetTenantTotalTimeSaved(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenantTotalTimeSaved failed: %v", err)
		}

		expectedTotal := 180 + 330 + 165
		if timeSaved != expectedTotal {
			t.Errorf("timeSaved = %d, want %d", timeSaved, expectedTotal)
		}
	})
}
