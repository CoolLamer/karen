package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// getTestDB returns a database pool for testing.
// Skips the test if DATABASE_URL is not set.
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

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	return db
}

func TestTenantOperations(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create a tenant
	tenant, err := s.CreateTenant(ctx, "Test Tenant", "Test system prompt for {{name}}")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	if tenant.ID == "" {
		t.Error("tenant ID should not be empty")
	}
	if tenant.Name != "Test Tenant" {
		t.Errorf("tenant name = %q, want %q", tenant.Name, "Test Tenant")
	}
	if tenant.SystemPrompt != "Test system prompt for {{name}}" {
		t.Errorf("tenant system_prompt = %q, want %q", tenant.SystemPrompt, "Test system prompt for {{name}}")
	}
	if tenant.Plan != "trial" {
		t.Errorf("tenant plan = %q, want %q", tenant.Plan, "trial")
	}
	if tenant.Status != "active" {
		t.Errorf("tenant status = %q, want %q", tenant.Status, "active")
	}

	// Get tenant by ID
	retrieved, err := s.GetTenantByID(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("GetTenantByID failed: %v", err)
	}
	if retrieved.ID != tenant.ID {
		t.Errorf("retrieved tenant ID = %q, want %q", retrieved.ID, tenant.ID)
	}

	// Update tenant
	err = s.UpdateTenant(ctx, tenant.ID, map[string]any{
		"name":            "Updated Tenant",
		"marketing_email": "test@example.com",
	})
	if err != nil {
		t.Fatalf("UpdateTenant failed: %v", err)
	}

	updated, err := s.GetTenantByID(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("GetTenantByID after update failed: %v", err)
	}
	if updated.Name != "Updated Tenant" {
		t.Errorf("updated tenant name = %q, want %q", updated.Name, "Updated Tenant")
	}

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
}

func TestUserOperations(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	testPhone := "+420777" + time.Now().Format("150405") // Unique phone number

	// FindOrCreateUser - creates new user
	user, isNew, err := s.FindOrCreateUser(ctx, testPhone)
	if err != nil {
		t.Fatalf("FindOrCreateUser failed: %v", err)
	}
	if !isNew {
		t.Error("expected isNew = true for new user")
	}
	if user.Phone != testPhone {
		t.Errorf("user phone = %q, want %q", user.Phone, testPhone)
	}
	if !user.PhoneVerified {
		t.Error("expected phone_verified = true after verification")
	}
	if user.Role != "owner" {
		t.Errorf("user role = %q, want %q", user.Role, "owner")
	}

	// FindOrCreateUser - finds existing user
	user2, isNew2, err := s.FindOrCreateUser(ctx, testPhone)
	if err != nil {
		t.Fatalf("FindOrCreateUser (existing) failed: %v", err)
	}
	if isNew2 {
		t.Error("expected isNew = false for existing user")
	}
	if user2.ID != user.ID {
		t.Errorf("user2 ID = %q, want %q", user2.ID, user.ID)
	}

	// GetUserByPhone
	user3, err := s.GetUserByPhone(ctx, testPhone)
	if err != nil {
		t.Fatalf("GetUserByPhone failed: %v", err)
	}
	if user3.ID != user.ID {
		t.Errorf("user3 ID = %q, want %q", user3.ID, user.ID)
	}

	// GetUserByID
	user4, err := s.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if user4.Phone != testPhone {
		t.Errorf("user4 phone = %q, want %q", user4.Phone, testPhone)
	}

	// UpdateUserName
	err = s.UpdateUserName(ctx, user.ID, "Test User")
	if err != nil {
		t.Fatalf("UpdateUserName failed: %v", err)
	}

	user5, _ := s.GetUserByID(ctx, user.ID)
	if user5.Name == nil || *user5.Name != "Test User" {
		t.Errorf("user name = %v, want %q", user5.Name, "Test User")
	}

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
}

func TestTenantUserAssignment(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create tenant
	tenant, err := s.CreateTenant(ctx, "Tenant for User", "Prompt")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Create user
	testPhone := "+420888" + time.Now().Format("150405")
	user, _, err := s.FindOrCreateUser(ctx, testPhone)
	if err != nil {
		t.Fatalf("FindOrCreateUser failed: %v", err)
	}

	// User should not have tenant initially
	if user.TenantID != nil {
		t.Error("new user should not have tenant_id")
	}

	// Assign user to tenant
	err = s.AssignUserToTenant(ctx, user.ID, tenant.ID)
	if err != nil {
		t.Fatalf("AssignUserToTenant failed: %v", err)
	}

	// Verify assignment
	user2, _ := s.GetUserByID(ctx, user.ID)
	if user2.TenantID == nil || *user2.TenantID != tenant.ID {
		t.Errorf("user tenant_id = %v, want %q", user2.TenantID, tenant.ID)
	}

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
}

func TestPhoneNumberRouting(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create tenant
	tenant, err := s.CreateTenant(ctx, "Routing Test Tenant", "Test prompt")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Assign phone number
	twilioNumber := "+1555" + time.Now().Format("0102150405")
	err = s.AssignPhoneNumberToTenant(ctx, tenant.ID, twilioNumber, "PNTEST123")
	if err != nil {
		t.Fatalf("AssignPhoneNumberToTenant failed: %v", err)
	}

	// GetTenantByTwilioNumber
	found, err := s.GetTenantByTwilioNumber(ctx, twilioNumber)
	if err != nil {
		t.Fatalf("GetTenantByTwilioNumber failed: %v", err)
	}
	if found.ID != tenant.ID {
		t.Errorf("found tenant ID = %q, want %q", found.ID, tenant.ID)
	}

	// GetTenantPhoneNumbers
	numbers, err := s.GetTenantPhoneNumbers(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("GetTenantPhoneNumbers failed: %v", err)
	}
	if len(numbers) != 1 {
		t.Fatalf("got %d phone numbers, want 1", len(numbers))
	}
	if numbers[0].TwilioNumber != twilioNumber {
		t.Errorf("phone number = %q, want %q", numbers[0].TwilioNumber, twilioNumber)
	}

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM tenant_phone_numbers WHERE tenant_id = $1", tenant.ID)
	_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
}

func TestSessionOperations(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create user first
	testPhone := "+420999" + time.Now().Format("150405")
	user, _, err := s.FindOrCreateUser(ctx, testPhone)
	if err != nil {
		t.Fatalf("FindOrCreateUser failed: %v", err)
	}

	// Create session
	tokenHash := "test-token-hash-" + time.Now().Format("150405")
	expiresAt := time.Now().Add(24 * time.Hour)
	err = s.CreateSession(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Check session is valid
	valid, err := s.IsSessionValid(ctx, tokenHash)
	if err != nil {
		t.Fatalf("IsSessionValid failed: %v", err)
	}
	if !valid {
		t.Error("session should be valid")
	}

	// Revoke session
	err = s.RevokeSession(ctx, tokenHash)
	if err != nil {
		t.Fatalf("RevokeSession failed: %v", err)
	}

	// Check session is no longer valid
	valid2, err := s.IsSessionValid(ctx, tokenHash)
	if err != nil {
		t.Fatalf("IsSessionValid after revoke failed: %v", err)
	}
	if valid2 {
		t.Error("session should be invalid after revocation")
	}

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM user_sessions WHERE user_id = $1", user.ID)
	_, _ = db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
}

func TestCallsWithTenant(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create tenant
	tenant, err := s.CreateTenant(ctx, "Calls Test Tenant", "Test prompt")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Create call with tenant
	callSid := "CA" + time.Now().Format("20060102150405")
	call := Call{
		TenantID:       &tenant.ID,
		Provider:       "twilio",
		ProviderCallID: callSid,
		FromNumber:     "+420777123456",
		ToNumber:       "+420228883001",
		Status:         "in_progress",
		StartedAt:      time.Now(),
	}

	err = s.UpsertCallWithTenant(ctx, call)
	if err != nil {
		t.Fatalf("UpsertCallWithTenant failed: %v", err)
	}

	// List calls by tenant
	calls, err := s.ListCallsByTenant(ctx, tenant.ID, 100)
	if err != nil {
		t.Fatalf("ListCallsByTenant failed: %v", err)
	}
	if len(calls) != 1 {
		t.Fatalf("got %d calls, want 1", len(calls))
	}
	if calls[0].ProviderCallID != callSid {
		t.Errorf("call provider_call_id = %q, want %q", calls[0].ProviderCallID, callSid)
	}

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM calls WHERE provider_call_id = $1", callSid)
	_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
}
