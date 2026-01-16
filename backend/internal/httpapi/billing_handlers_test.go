package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lukasbauer/karen/internal/store"
)

// strPtr is defined in calls_handlers_test.go, reuse it here via test package scope

func TestGetPriceID(t *testing.T) {
	// Save original environment variables
	origBasicMonthly := stripePriceBasicMonthly
	origBasicAnnual := stripePriceBasicAnnual
	origProMonthly := stripePriceProMonthly
	origProAnnual := stripePriceProAnnual

	// Set test values
	stripePriceBasicMonthly = "price_basic_monthly"
	stripePriceBasicAnnual = "price_basic_annual"
	stripePriceProMonthly = "price_pro_monthly"
	stripePriceProAnnual = "price_pro_annual"

	defer func() {
		// Restore original values
		stripePriceBasicMonthly = origBasicMonthly
		stripePriceBasicAnnual = origBasicAnnual
		stripePriceProMonthly = origProMonthly
		stripePriceProAnnual = origProAnnual
	}()

	tests := []struct {
		name     string
		plan     string
		interval string
		expected string
	}{
		{"basic monthly", "basic", "monthly", "price_basic_monthly"},
		{"basic annual", "basic", "annual", "price_basic_annual"},
		{"pro monthly", "pro", "monthly", "price_pro_monthly"},
		{"pro annual", "pro", "annual", "price_pro_annual"},
		{"invalid plan", "invalid", "monthly", ""},
		{"empty plan", "", "monthly", ""},
		{"basic with invalid interval", "basic", "weekly", "price_basic_monthly"}, // defaults to monthly
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPriceID(tt.plan, tt.interval)
			if got != tt.expected {
				t.Errorf("getPriceID(%q, %q) = %q, want %q", tt.plan, tt.interval, got, tt.expected)
			}
		})
	}
}

func TestGetPlanFromPriceID(t *testing.T) {
	// Save original environment variables
	origBasicMonthly := stripePriceBasicMonthly
	origBasicAnnual := stripePriceBasicAnnual
	origProMonthly := stripePriceProMonthly
	origProAnnual := stripePriceProAnnual

	// Set test values
	stripePriceBasicMonthly = "price_basic_monthly"
	stripePriceBasicAnnual = "price_basic_annual"
	stripePriceProMonthly = "price_pro_monthly"
	stripePriceProAnnual = "price_pro_annual"

	defer func() {
		// Restore original values
		stripePriceBasicMonthly = origBasicMonthly
		stripePriceBasicAnnual = origBasicAnnual
		stripePriceProMonthly = origProMonthly
		stripePriceProAnnual = origProAnnual
	}()

	tests := []struct {
		name     string
		priceID  string
		expected string
	}{
		{"basic monthly price", "price_basic_monthly", "basic"},
		{"basic annual price", "price_basic_annual", "basic"},
		{"pro monthly price", "price_pro_monthly", "pro"},
		{"pro annual price", "price_pro_annual", "pro"},
		{"unknown price", "price_unknown", "basic"}, // defaults to basic
		{"empty price", "", "basic"},                // defaults to basic
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPlanFromPriceID(tt.priceID)
			if got != tt.expected {
				t.Errorf("getPlanFromPriceID(%q) = %q, want %q", tt.priceID, got, tt.expected)
			}
		})
	}
}

func TestHandleCreateCheckoutValidation(t *testing.T) {
	r := &Router{
		cfg: RouterConfig{
			JWTSecret: "test-secret-key",
			JWTExpiry: 1 * time.Hour,
		},
		logger: log.New(io.Discard, "", 0),
	}

	tests := []struct {
		name           string
		body           string
		authUser       *AuthUser
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "no auth user",
			body:           `{"plan": "basic", "interval": "monthly"}`,
			authUser:       nil,
			expectedStatus: http.StatusForbidden,
			expectedError:  "no tenant assigned",
		},
		{
			name:           "no tenant assigned",
			body:           `{"plan": "basic", "interval": "monthly"}`,
			authUser:       &AuthUser{ID: "user-123", Phone: "+420123456789", TenantID: nil},
			expectedStatus: http.StatusForbidden,
			expectedError:  "no tenant assigned",
		},
		{
			name:           "invalid request body",
			body:           `{invalid json}`,
			authUser:       &AuthUser{ID: "user-123", Phone: "+420123456789", TenantID: strPtr("tenant-123")},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name:           "invalid plan",
			body:           `{"plan": "invalid", "interval": "monthly"}`,
			authUser:       &AuthUser{ID: "user-123", Phone: "+420123456789", TenantID: strPtr("tenant-123")},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid plan or interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/billing/checkout", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			if tt.authUser != nil {
				ctx := context.WithValue(req.Context(), userContextKey, tt.authUser)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			r.handleCreateCheckout(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.expectedStatus)
			}

			var resp map[string]string
			_ = json.NewDecoder(rec.Body).Decode(&resp)
			if !strings.Contains(resp["error"], tt.expectedError) {
				t.Errorf("error = %q, should contain %q", resp["error"], tt.expectedError)
			}
		})
	}
}

func TestHandleCreatePortalValidation(t *testing.T) {
	r := &Router{
		cfg: RouterConfig{
			JWTSecret: "test-secret-key",
			JWTExpiry: 1 * time.Hour,
		},
		logger: log.New(io.Discard, "", 0),
	}

	tests := []struct {
		name           string
		authUser       *AuthUser
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "no auth user",
			authUser:       nil,
			expectedStatus: http.StatusForbidden,
			expectedError:  "no tenant assigned",
		},
		{
			name:           "no tenant assigned",
			authUser:       &AuthUser{ID: "user-123", Phone: "+420123456789", TenantID: nil},
			expectedStatus: http.StatusForbidden,
			expectedError:  "no tenant assigned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/billing/portal", nil)

			if tt.authUser != nil {
				ctx := context.WithValue(req.Context(), userContextKey, tt.authUser)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			r.handleCreatePortal(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.expectedStatus)
			}

			var resp map[string]string
			_ = json.NewDecoder(rec.Body).Decode(&resp)
			if !strings.Contains(resp["error"], tt.expectedError) {
				t.Errorf("error = %q, should contain %q", resp["error"], tt.expectedError)
			}
		})
	}
}

func TestHandleStripeWebhookValidation(t *testing.T) {
	r := &Router{
		cfg:    RouterConfig{},
		logger: log.New(io.Discard, "", 0),
	}

	tests := []struct {
		name           string
		body           string
		signature      string
		expectedStatus int
	}{
		{
			name:           "missing signature",
			body:           `{"type": "checkout.session.completed"}`,
			signature:      "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid signature",
			body:           `{"type": "checkout.session.completed"}`,
			signature:      "invalid-signature",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", strings.NewReader(tt.body))
			if tt.signature != "" {
				req.Header.Set("Stripe-Signature", tt.signature)
			}

			rec := httptest.NewRecorder()
			r.handleStripeWebhook(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.expectedStatus)
			}
		})
	}
}

func TestHandleGetBillingValidation(t *testing.T) {
	r := &Router{
		cfg: RouterConfig{
			JWTSecret: "test-secret-key",
		},
		logger: log.New(io.Discard, "", 0),
	}

	tests := []struct {
		name           string
		authUser       *AuthUser
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "no auth user",
			authUser:       nil,
			expectedStatus: http.StatusNotFound,
			expectedError:  "no tenant",
		},
		{
			name:           "no tenant assigned",
			authUser:       &AuthUser{ID: "user-123", Phone: "+420123456789", TenantID: nil},
			expectedStatus: http.StatusNotFound,
			expectedError:  "no tenant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/billing", nil)

			if tt.authUser != nil {
				ctx := context.WithValue(req.Context(), userContextKey, tt.authUser)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			r.handleGetBilling(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.expectedStatus)
			}

			var resp map[string]string
			_ = json.NewDecoder(rec.Body).Decode(&resp)
			if !strings.Contains(resp["error"], tt.expectedError) {
				t.Errorf("error = %q, should contain %q", resp["error"], tt.expectedError)
			}
		})
	}
}

// Integration tests (require database)

func getTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	return db
}

func TestBillingIntegration(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := store.New(db)
	ctx := context.Background()

	// Create test tenant
	tenant, err := s.CreateTenant(ctx, "Test Billing Tenant", "Test prompt", "cs")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Cleanup at the end
	defer func() {
		_, _ = db.Exec(ctx, "DELETE FROM tenant_usage WHERE tenant_id = $1", tenant.ID)
		_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
	}()

	// Test that new tenant starts with trial plan
	t.Run("new tenant has trial plan", func(t *testing.T) {
		if tenant.Plan != "trial" {
			t.Errorf("new tenant plan = %q, want %q", tenant.Plan, "trial")
		}
		if tenant.Status != "active" {
			t.Errorf("new tenant status = %q, want %q", tenant.Status, "active")
		}
	})

	// Test trial_ends_at is set
	t.Run("new tenant has trial_ends_at set", func(t *testing.T) {
		if tenant.TrialEndsAt == nil {
			t.Error("trial_ends_at should not be nil for new tenant")
		} else {
			// Should be approximately 14 days from now
			expectedEnd := time.Now().Add(14 * 24 * time.Hour)
			diff := tenant.TrialEndsAt.Sub(expectedEnd)
			if diff < -time.Hour || diff > time.Hour {
				t.Errorf("trial_ends_at = %v, expected around %v", tenant.TrialEndsAt, expectedEnd)
			}
		}
	})

	// Test CanTenantReceiveCalls for new trial
	t.Run("new trial tenant can receive calls", func(t *testing.T) {
		// Need to get fresh tenant to pass to CanTenantReceiveCalls
		status := store.CanTenantReceiveCalls(tenant)
		if !status.CanReceive {
			t.Errorf("new trial tenant should be able to receive calls, got reason: %s", status.Reason)
		}
	})

	// Test GetTenantBillingInfo
	t.Run("GetTenantBillingInfo returns correct info", func(t *testing.T) {
		billingInfo, err := s.GetTenantBillingInfo(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenantBillingInfo failed: %v", err)
		}
		if billingInfo.Plan != "trial" {
			t.Errorf("billing plan = %q, want %q", billingInfo.Plan, "trial")
		}
		if billingInfo.Status != "active" {
			t.Errorf("billing status = %q, want %q", billingInfo.Status, "active")
		}
	})

	// Test UpdateTenantBilling
	t.Run("UpdateTenantBilling updates plan and status", func(t *testing.T) {
		err := s.UpdateTenantBilling(ctx, tenant.ID, map[string]any{
			"plan":               "basic",
			"status":             "active",
			"stripe_customer_id": "cus_test123",
		})
		if err != nil {
			t.Fatalf("UpdateTenantBilling failed: %v", err)
		}

		// Verify updates
		billingInfo, err := s.GetTenantBillingInfo(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenantBillingInfo failed: %v", err)
		}
		if billingInfo.Plan != "basic" {
			t.Errorf("updated plan = %q, want %q", billingInfo.Plan, "basic")
		}
		if billingInfo.StripeCustomerID == nil || *billingInfo.StripeCustomerID != "cus_test123" {
			t.Errorf("stripe_customer_id not updated correctly")
		}
	})

	// Test GetTenantIDByStripeCustomer
	t.Run("GetTenantIDByStripeCustomer finds tenant", func(t *testing.T) {
		foundID, err := s.GetTenantIDByStripeCustomer(ctx, "cus_test123")
		if err != nil {
			t.Fatalf("GetTenantIDByStripeCustomer failed: %v", err)
		}
		if foundID != tenant.ID {
			t.Errorf("found tenant ID = %q, want %q", foundID, tenant.ID)
		}
	})

	// Test GetTenantIDByStripeCustomer with non-existent customer
	t.Run("GetTenantIDByStripeCustomer returns error for unknown customer", func(t *testing.T) {
		_, err := s.GetTenantIDByStripeCustomer(ctx, "cus_nonexistent")
		if err == nil {
			t.Error("expected error for non-existent customer")
		}
	})

	// Test IncrementTenantUsage
	t.Run("IncrementTenantUsage increments counters", func(t *testing.T) {
		// Get initial usage
		initialUsage, _ := s.GetTenantCurrentUsage(ctx, tenant.ID)
		initialCalls := 0
		if initialUsage != nil {
			initialCalls = initialUsage.CallsCount
		}

		// Increment usage (60 seconds, not spam)
		err := s.IncrementTenantUsage(ctx, tenant.ID, 60, false)
		if err != nil {
			t.Fatalf("IncrementTenantUsage failed: %v", err)
		}

		// Verify increment
		usage, err := s.GetTenantCurrentUsage(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenantCurrentUsage failed: %v", err)
		}
		if usage.CallsCount != initialCalls+1 {
			t.Errorf("calls_count = %d, want %d", usage.CallsCount, initialCalls+1)
		}
	})

	// Test ResetTenantPeriodCalls
	t.Run("ResetTenantPeriodCalls resets counter", func(t *testing.T) {
		err := s.ResetTenantPeriodCalls(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("ResetTenantPeriodCalls failed: %v", err)
		}

		// Verify reset by checking tenant directly
		var currentCalls int
		err = db.QueryRow(ctx, "SELECT COALESCE(current_period_calls, 0) FROM tenants WHERE id = $1", tenant.ID).Scan(&currentCalls)
		if err != nil {
			t.Fatalf("failed to query current_period_calls: %v", err)
		}
		if currentCalls != 0 {
			t.Errorf("current_period_calls = %d, want 0", currentCalls)
		}
	})
}

func TestTrialLimitEnforcement(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := store.New(db)
	ctx := context.Background()

	// Create test tenant
	tenant, err := s.CreateTenant(ctx, "Trial Limit Test", "Test prompt", "cs")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Cleanup at the end
	defer func() {
		_, _ = db.Exec(ctx, "DELETE FROM tenant_usage WHERE tenant_id = $1", tenant.ID)
		_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
	}()

	// Helper to get fresh tenant
	getFreshTenant := func() *store.Tenant {
		freshTenant, err := s.GetTenantByID(ctx, tenant.ID)
		if err != nil {
			t.Fatalf("GetTenantByID failed: %v", err)
		}
		return freshTenant
	}

	// Simulate reaching 20 calls (trial limit)
	t.Run("trial limit blocks after 20 calls", func(t *testing.T) {
		// Set current_period_calls to 20
		_, err := db.Exec(ctx, "UPDATE tenants SET current_period_calls = 20 WHERE id = $1", tenant.ID)
		if err != nil {
			t.Fatalf("failed to update current_period_calls: %v", err)
		}

		status := store.CanTenantReceiveCalls(getFreshTenant())
		if status.CanReceive {
			t.Error("trial tenant at 20 calls should not be able to receive calls")
		}
		if status.Reason != "limit_exceeded" {
			t.Errorf("reason = %q, want %q", status.Reason, "limit_exceeded")
		}
	})

	// Test expired trial (trial_ends_at in the past)
	t.Run("expired trial blocks calls", func(t *testing.T) {
		// Reset calls and set expired trial
		expiredTime := time.Now().Add(-24 * time.Hour)
		_, err := db.Exec(ctx, "UPDATE tenants SET current_period_calls = 0, trial_ends_at = $1 WHERE id = $2", expiredTime, tenant.ID)
		if err != nil {
			t.Fatalf("failed to update trial_ends_at: %v", err)
		}

		status := store.CanTenantReceiveCalls(getFreshTenant())
		if status.CanReceive {
			t.Error("expired trial tenant should not be able to receive calls")
		}
		if status.Reason != "trial_expired" {
			t.Errorf("reason = %q, want %q", status.Reason, "trial_expired")
		}
	})

	// Test paid plan has no limits
	t.Run("paid plan has unlimited calls", func(t *testing.T) {
		// Upgrade to basic plan
		err := s.UpdateTenantBilling(ctx, tenant.ID, map[string]any{
			"plan":   "basic",
			"status": "active",
		})
		if err != nil {
			t.Fatalf("UpdateTenantBilling failed: %v", err)
		}

		// Set high call count
		_, err = db.Exec(ctx, "UPDATE tenants SET current_period_calls = 100 WHERE id = $1", tenant.ID)
		if err != nil {
			t.Fatalf("failed to update current_period_calls: %v", err)
		}

		// For basic plan with 50 call limit, should be blocked
		status := store.CanTenantReceiveCalls(getFreshTenant())
		if status.CanReceive {
			t.Error("basic plan at 100 calls should be blocked")
		}
		if status.Reason != "limit_exceeded" {
			t.Errorf("reason = %q, want %q", status.Reason, "limit_exceeded")
		}

		// Upgrade to pro plan (unlimited)
		err = s.UpdateTenantBilling(ctx, tenant.ID, map[string]any{
			"plan": "pro",
		})
		if err != nil {
			t.Fatalf("UpdateTenantBilling failed: %v", err)
		}

		status = store.CanTenantReceiveCalls(getFreshTenant())
		if !status.CanReceive {
			t.Errorf("pro plan should have unlimited calls, got reason: %s", status.Reason)
		}
	})
}
