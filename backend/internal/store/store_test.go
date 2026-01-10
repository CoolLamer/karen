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

func TestGetCallDetailWithTenantCheck(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create tenant
	tenant, err := s.CreateTenant(ctx, "Detail Test Tenant", "Test prompt")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Create call with tenant
	callSid := "CADETAIL" + time.Now().Format("20060102150405")
	call := Call{
		TenantID:       &tenant.ID,
		Provider:       "twilio",
		ProviderCallID: callSid,
		FromNumber:     "+420777123456",
		ToNumber:       "+420228883001",
		Status:         "completed",
		StartedAt:      time.Now(),
	}

	err = s.UpsertCallWithTenant(ctx, call)
	if err != nil {
		t.Fatalf("UpsertCallWithTenant failed: %v", err)
	}

	// Test GetCallDetailWithTenantCheck
	detail, tenantID, err := s.GetCallDetailWithTenantCheck(ctx, callSid)
	if err != nil {
		t.Fatalf("GetCallDetailWithTenantCheck failed: %v", err)
	}

	if detail.ProviderCallID != callSid {
		t.Errorf("detail.ProviderCallID = %q, want %q", detail.ProviderCallID, callSid)
	}

	if tenantID == nil {
		t.Fatal("tenantID should not be nil")
	}
	if *tenantID != tenant.ID {
		t.Errorf("tenantID = %q, want %q", *tenantID, tenant.ID)
	}

	// Verify that detail.TenantID is also set
	if detail.TenantID == nil || *detail.TenantID != tenant.ID {
		t.Errorf("detail.TenantID = %v, want %q", detail.TenantID, tenant.ID)
	}

	// Test GetCallDetail (wrapper function)
	detail2, err := s.GetCallDetail(ctx, callSid)
	if err != nil {
		t.Fatalf("GetCallDetail failed: %v", err)
	}
	if detail2.ProviderCallID != callSid {
		t.Errorf("detail2.ProviderCallID = %q, want %q", detail2.ProviderCallID, callSid)
	}

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM calls WHERE provider_call_id = $1", callSid)
	_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
}

func TestGetTenantOwnerPhone(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create tenant
	tenant, err := s.CreateTenant(ctx, "Owner Phone Test", "Test prompt")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Create user and assign to tenant
	testPhone := "+420555" + time.Now().Format("150405")
	user, _, err := s.FindOrCreateUser(ctx, testPhone)
	if err != nil {
		t.Fatalf("FindOrCreateUser failed: %v", err)
	}

	err = s.AssignUserToTenant(ctx, user.ID, tenant.ID)
	if err != nil {
		t.Fatalf("AssignUserToTenant failed: %v", err)
	}

	// Test GetTenantOwnerPhone
	ownerPhone, err := s.GetTenantOwnerPhone(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("GetTenantOwnerPhone failed: %v", err)
	}

	if ownerPhone != testPhone {
		t.Errorf("ownerPhone = %q, want %q", ownerPhone, testPhone)
	}

	// Test with non-existent tenant
	_, err = s.GetTenantOwnerPhone(ctx, "non-existent-tenant-id")
	if err == nil {
		t.Error("GetTenantOwnerPhone should fail for non-existent tenant")
	}

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
}

func TestUpdateCallEndedBy(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create a call
	callSid := "CA" + time.Now().Format("20060102150405")
	call := Call{
		Provider:       "twilio",
		ProviderCallID: callSid,
		FromNumber:     "+420777123456",
		ToNumber:       "+420228883001",
		Status:         "in_progress",
		StartedAt:      time.Now(),
	}

	err := s.UpsertCall(ctx, call)
	if err != nil {
		t.Fatalf("UpsertCall failed: %v", err)
	}

	// Test updating ended_by to "agent"
	err = s.UpdateCallEndedBy(ctx, callSid, "agent")
	if err != nil {
		t.Fatalf("UpdateCallEndedBy (agent) failed: %v", err)
	}

	// Verify the update
	detail, err := s.GetCallDetail(ctx, callSid)
	if err != nil {
		t.Fatalf("GetCallDetail failed: %v", err)
	}
	if detail.EndedBy == nil {
		t.Error("ended_by should not be nil")
	} else if *detail.EndedBy != "agent" {
		t.Errorf("ended_by = %q, want %q", *detail.EndedBy, "agent")
	}

	// Test updating ended_by to "caller"
	err = s.UpdateCallEndedBy(ctx, callSid, "caller")
	if err != nil {
		t.Fatalf("UpdateCallEndedBy (caller) failed: %v", err)
	}

	// Verify the update
	detail2, err := s.GetCallDetail(ctx, callSid)
	if err != nil {
		t.Fatalf("GetCallDetail failed: %v", err)
	}
	if detail2.EndedBy == nil {
		t.Error("ended_by should not be nil")
	} else if *detail2.EndedBy != "caller" {
		t.Errorf("ended_by = %q, want %q", *detail2.EndedBy, "caller")
	}

	// Test updating non-existent call (should not error)
	err = s.UpdateCallEndedBy(ctx, "CA_NONEXISTENT", "agent")
	if err != nil {
		t.Errorf("UpdateCallEndedBy on non-existent call should not error: %v", err)
	}

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM calls WHERE provider_call_id = $1", callSid)
}

func TestGreetingUtteranceStorage(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create a call
	callSid := "CA" + time.Now().Format("20060102150405")
	call := Call{
		Provider:       "twilio",
		ProviderCallID: callSid,
		FromNumber:     "+420777123456",
		ToNumber:       "+420228883001",
		Status:         "in_progress",
		StartedAt:      time.Now(),
	}

	err := s.UpsertCall(ctx, call)
	if err != nil {
		t.Fatalf("UpsertCall failed: %v", err)
	}

	// Get the call ID
	detail, err := s.GetCallDetail(ctx, callSid)
	if err != nil {
		t.Fatalf("GetCallDetail failed: %v", err)
	}
	callID := detail.ID

	// Simulate storing greeting as first utterance
	startTime := time.Now().UTC()
	greetingText := "Dobrý den, tady Asistentka Karen."
	err = s.InsertUtterance(ctx, callID, Utterance{
		Speaker:     "agent",
		Text:        greetingText,
		Sequence:    1,
		StartedAt:   &startTime,
		Interrupted: false,
	})
	if err != nil {
		t.Fatalf("InsertUtterance (greeting) failed: %v", err)
	}

	// Store a caller utterance (sequence 2)
	callerStartTime := time.Now().UTC()
	err = s.InsertUtterance(ctx, callID, Utterance{
		Speaker:     "caller",
		Text:        "Ahoj, potřebuji mluvit s Lukášem.",
		Sequence:    2,
		StartedAt:   &callerStartTime,
		Interrupted: false,
	})
	if err != nil {
		t.Fatalf("InsertUtterance (caller) failed: %v", err)
	}

	// Retrieve call detail with utterances
	result, err := s.GetCallDetail(ctx, callSid)
	if err != nil {
		t.Fatalf("GetCallDetail with utterances failed: %v", err)
	}

	// Verify we have 2 utterances
	if len(result.Utterances) != 2 {
		t.Fatalf("got %d utterances, want 2", len(result.Utterances))
	}

	// Verify first utterance is the greeting
	firstUtterance := result.Utterances[0]
	if firstUtterance.Sequence != 1 {
		t.Errorf("first utterance sequence = %d, want 1", firstUtterance.Sequence)
	}
	if firstUtterance.Speaker != "agent" {
		t.Errorf("first utterance speaker = %q, want %q", firstUtterance.Speaker, "agent")
	}
	if firstUtterance.Text != greetingText {
		t.Errorf("first utterance text = %q, want %q", firstUtterance.Text, greetingText)
	}

	// Verify second utterance is the caller
	secondUtterance := result.Utterances[1]
	if secondUtterance.Sequence != 2 {
		t.Errorf("second utterance sequence = %d, want 2", secondUtterance.Sequence)
	}
	if secondUtterance.Speaker != "caller" {
		t.Errorf("second utterance speaker = %q, want %q", secondUtterance.Speaker, "caller")
	}

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM call_utterances WHERE call_id = $1", callID)
	_, _ = db.Exec(ctx, "DELETE FROM calls WHERE provider_call_id = $1", callSid)
}

func TestCallListIncludesEndedBy(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	s := New(db)
	ctx := context.Background()

	// Create tenant
	tenant, err := s.CreateTenant(ctx, "EndedBy Test Tenant", "Test prompt")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Create call with ended_by set
	callSid := "CA" + time.Now().Format("20060102150405")
	call := Call{
		TenantID:       &tenant.ID,
		Provider:       "twilio",
		ProviderCallID: callSid,
		FromNumber:     "+420777123456",
		ToNumber:       "+420228883001",
		Status:         "completed",
		StartedAt:      time.Now(),
	}

	err = s.UpsertCallWithTenant(ctx, call)
	if err != nil {
		t.Fatalf("UpsertCallWithTenant failed: %v", err)
	}

	// Set ended_by
	err = s.UpdateCallEndedBy(ctx, callSid, "agent")
	if err != nil {
		t.Fatalf("UpdateCallEndedBy failed: %v", err)
	}

	// Test ListCalls includes ended_by
	calls, err := s.ListCalls(ctx, 100)
	if err != nil {
		t.Fatalf("ListCalls failed: %v", err)
	}

	found := false
	for _, c := range calls {
		if c.ProviderCallID == callSid {
			found = true
			if c.EndedBy == nil {
				t.Error("ended_by should not be nil in ListCalls")
			} else if *c.EndedBy != "agent" {
				t.Errorf("ended_by = %q, want %q", *c.EndedBy, "agent")
			}
			break
		}
	}
	if !found {
		t.Error("call not found in ListCalls")
	}

	// Test ListCallsByTenant includes ended_by
	tenantCalls, err := s.ListCallsByTenant(ctx, tenant.ID, 100)
	if err != nil {
		t.Fatalf("ListCallsByTenant failed: %v", err)
	}

	if len(tenantCalls) != 1 {
		t.Fatalf("got %d calls, want 1", len(tenantCalls))
	}
	if tenantCalls[0].EndedBy == nil {
		t.Error("ended_by should not be nil in ListCallsByTenant")
	} else if *tenantCalls[0].EndedBy != "agent" {
		t.Errorf("ended_by = %q, want %q", *tenantCalls[0].EndedBy, "agent")
	}

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM calls WHERE provider_call_id = $1", callSid)
	_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
}
