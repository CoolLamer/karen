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

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lukasbauer/karen/internal/llm"
	"github.com/lukasbauer/karen/internal/store"
)

func TestIsValidE164(t *testing.T) {
	tests := []struct {
		phone string
		valid bool
	}{
		{"+420777123456", true},
		{"+1234567890", true},
		{"+44207123456", true},
		{"+86123456789012", true},
		{"420777123456", false},     // Missing +
		{"+0777123456", false},      // Starts with 0
		{"+123456", false},          // Too short (only 6 digits after +)
		{"", false},                 // Empty
		{"+", false},                // Just +
		{"+1", false},               // Too short
		{"phone", false},            // Not a number
		{"+420 777 123 456", false}, // Spaces
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			if got := isValidE164(tt.phone); got != tt.valid {
				t.Errorf("isValidE164(%q) = %v, want %v", tt.phone, got, tt.valid)
			}
		})
	}
}

func TestHashToken(t *testing.T) {
	token := "test-token-123"

	hash1 := hashToken(token)
	hash2 := hashToken(token)

	// Same token should produce same hash
	if hash1 != hash2 {
		t.Error("same token should produce same hash")
	}

	// Hash should be hex-encoded SHA256 (64 characters)
	if len(hash1) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash1))
	}

	// Different tokens should produce different hashes
	hash3 := hashToken("different-token")
	if hash1 == hash3 {
		t.Error("different tokens should produce different hashes")
	}
}

func TestGenerateDefaultSystemPrompt(t *testing.T) {
	prompt := llm.GenerateDefaultSystemPrompt("Jan")

	if !strings.Contains(prompt, "Jan") {
		t.Error("prompt should contain the user's name")
	}

	if !strings.Contains(prompt, "Karen") {
		t.Error("prompt should mention Karen")
	}

	if !strings.Contains(prompt, "TVŮJ ÚKOL") {
		t.Error("prompt should contain task description")
	}
}

func TestJWTGeneration(t *testing.T) {
	// Create a minimal router for testing
	r := &Router{
		cfg: RouterConfig{
			JWTSecret: "test-secret-key",
			JWTExpiry: 1 * time.Hour,
		},
	}

	user := &store.User{
		ID:    "user-123",
		Phone: "+420777123456",
	}
	tenantID := "tenant-456"
	user.TenantID = &tenantID

	token, expiresAt, err := r.generateJWT(user)
	if err != nil {
		t.Fatalf("generateJWT failed: %v", err)
	}

	if token == "" {
		t.Error("token should not be empty")
	}

	if time.Until(expiresAt) < 50*time.Minute {
		t.Error("token should expire in about 1 hour")
	}

	// Parse and validate the token
	parsed, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	claims, ok := parsed.Claims.(*JWTClaims)
	if !ok {
		t.Fatal("failed to cast claims")
	}

	if claims.UserID != "user-123" {
		t.Errorf("claims.UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.Phone != "+420777123456" {
		t.Errorf("claims.Phone = %q, want %q", claims.Phone, "+420777123456")
	}
	if claims.TenantID == nil || *claims.TenantID != "tenant-456" {
		t.Errorf("claims.TenantID = %v, want %q", claims.TenantID, "tenant-456")
	}
}

func TestWithAuthMiddleware(t *testing.T) {
	// Create router with test config
	r := &Router{
		cfg: RouterConfig{
			JWTSecret: "test-secret-key",
			JWTExpiry: 1 * time.Hour,
		},
		logger: log.New(io.Discard, "", 0),
	}

	// Create a test handler that checks for auth user
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		user := getAuthUser(req.Context())
		if user == nil {
			t.Error("auth user should be in context")
			http.Error(w, "no user", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(user.ID))
	})

	// Wrap with auth middleware
	protected := r.withAuth(testHandler)

	t.Run("missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		protected(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid authorization format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		rec := httptest.NewRecorder()

		protected(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()

		protected(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})
}

func TestSendCodeValidation(t *testing.T) {
	r := &Router{
		cfg:    RouterConfig{},
		logger: log.New(io.Discard, "", 0),
	}

	t.Run("invalid phone format", func(t *testing.T) {
		body := `{"phone": "invalid"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/send-code", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handleSendCode(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}

		var resp map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if !strings.Contains(resp["error"], "invalid phone format") {
			t.Errorf("error = %q, should mention invalid phone format", resp["error"])
		}
	})

	t.Run("twilio not configured", func(t *testing.T) {
		body := `{"phone": "+420777123456"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/send-code", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handleSendCode(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
		}
	})
}

func TestVerifyCodeValidation(t *testing.T) {
	r := &Router{
		cfg:    RouterConfig{},
		logger: log.New(io.Discard, "", 0),
	}

	t.Run("invalid phone format", func(t *testing.T) {
		body := `{"phone": "invalid", "code": "123456"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/verify-code", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handleVerifyCode(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid code length", func(t *testing.T) {
		body := `{"phone": "+420777123456", "code": "123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/verify-code", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handleVerifyCode(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}

		var resp map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if !strings.Contains(resp["error"], "6 digits") {
			t.Errorf("error = %q, should mention 6 digits", resp["error"])
		}
	})

	t.Run("twilio not configured", func(t *testing.T) {
		body := `{"phone": "+420777123456", "code": "123456"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/verify-code", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handleVerifyCode(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
		}
	})
}

func TestGetAuthUser(t *testing.T) {
	t.Run("no user in context", func(t *testing.T) {
		ctx := context.Background()
		user := getAuthUser(ctx)
		if user != nil {
			t.Error("expected nil user for empty context")
		}
	})

	t.Run("user in context", func(t *testing.T) {
		authUser := &AuthUser{
			ID:    "user-123",
			Phone: "+420777123456",
		}
		ctx := context.WithValue(context.Background(), userContextKey, authUser)

		user := getAuthUser(ctx)
		if user == nil {
			t.Fatal("expected user in context")
		}
		if user.ID != "user-123" {
			t.Errorf("user ID = %q, want %q", user.ID, "user-123")
		}
	})
}

// Integration tests (require database)
func getTestRouterWithDB(t *testing.T) (*Router, *pgxpool.Pool, func()) {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	s := store.New(db)

	r := &Router{
		cfg: RouterConfig{
			JWTSecret: "test-secret-key-for-integration",
			JWTExpiry: 1 * time.Hour,
		},
		logger: log.New(io.Discard, "", 0),
		store:  s,
	}

	cleanup := func() {
		db.Close()
	}

	return r, db, cleanup
}

func TestGenerateSystemPromptWithVIPs(t *testing.T) {
	t.Run("basic prompt with name only", func(t *testing.T) {
		prompt := llm.GenerateSystemPromptWithVIPs("Jan", nil, nil)

		if !strings.Contains(prompt, "Jan") {
			t.Error("prompt should contain the user's name")
		}
		if !strings.Contains(prompt, "Karen") {
			t.Error("prompt should mention Karen")
		}
		if strings.Contains(prompt, "[PŘEPOJIT]") {
			t.Error("prompt without VIPs should not contain forward marker")
		}
		if strings.Contains(prompt, "KRIZOVÉ SITUACE") {
			t.Error("prompt without VIPs should not contain VIP section")
		}
	})

	t.Run("prompt with VIP names", func(t *testing.T) {
		vipNames := []string{"Máma", "Táta", "Honza"}
		prompt := llm.GenerateSystemPromptWithVIPs("Petr", vipNames, nil)

		if !strings.Contains(prompt, "Petr") {
			t.Error("prompt should contain the user's name")
		}
		if !strings.Contains(prompt, "KRIZOVÉ SITUACE") {
			t.Error("prompt with VIPs should contain VIP section")
		}
		for _, vip := range vipNames {
			if !strings.Contains(prompt, vip) {
				t.Errorf("prompt should contain VIP name %q", vip)
			}
		}
		if !strings.Contains(prompt, "[PŘEPOJIT]") {
			t.Error("prompt with VIPs should contain forward marker instruction")
		}
	})

	t.Run("prompt with empty VIP names", func(t *testing.T) {
		emptyVips := []string{}
		prompt := llm.GenerateSystemPromptWithVIPs("Eva", emptyVips, nil)

		if strings.Contains(prompt, "KRIZOVÉ SITUACE") {
			t.Error("prompt with empty VIPs should not contain VIP section")
		}
	})

	t.Run("prompt with marketing email", func(t *testing.T) {
		email := "nabidky@example.com"
		prompt := llm.GenerateSystemPromptWithVIPs("Karel", nil, &email)

		if !strings.Contains(prompt, "Karel") {
			t.Error("prompt should contain the user's name")
		}
		if !strings.Contains(prompt, email) {
			t.Error("prompt should contain the marketing email")
		}
		if !strings.Contains(prompt, "MARKETING") {
			t.Error("prompt with email should contain marketing section")
		}
	})

	t.Run("prompt with both VIPs and marketing email", func(t *testing.T) {
		vipNames := []string{"Rodina"}
		email := "info@firma.cz"
		prompt := llm.GenerateSystemPromptWithVIPs("Lukáš", vipNames, &email)

		if !strings.Contains(prompt, "Lukáš") {
			t.Error("prompt should contain the user's name")
		}
		if !strings.Contains(prompt, "KRIZOVÉ SITUACE") {
			t.Error("prompt should contain VIP section")
		}
		if !strings.Contains(prompt, "Rodina") {
			t.Error("prompt should contain VIP name")
		}
		if !strings.Contains(prompt, email) {
			t.Error("prompt should contain marketing email")
		}
	})

	t.Run("prompt with nil marketing email", func(t *testing.T) {
		prompt := llm.GenerateSystemPromptWithVIPs("Anna", nil, nil)

		// When no email is set, marketing section should still exist but without an email address
		if !strings.Contains(prompt, "MARKETING") {
			t.Error("prompt should contain marketing section")
		}
		if !strings.Contains(prompt, "nemá zájem") {
			t.Error("prompt without email should tell callers owner has no interest")
		}
		// Should not contain specific email address pattern
		if strings.Contains(prompt, "@") {
			t.Error("prompt without email should not contain email address")
		}
	})

	t.Run("prompt with empty marketing email", func(t *testing.T) {
		emptyEmail := ""
		prompt := llm.GenerateSystemPromptWithVIPs("Martin", nil, &emptyEmail)

		// When email is empty, marketing section should still exist but without an email address
		if !strings.Contains(prompt, "MARKETING") {
			t.Error("prompt should contain marketing section")
		}
		// Should not contain specific email address pattern
		if strings.Contains(prompt, "mohou nabídku poslat na email") {
			t.Error("prompt with empty email should not offer to send email")
		}
	})
}

func TestCompleteOnboardingIntegration(t *testing.T) {
	r, db, cleanup := getTestRouterWithDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test user first
	testPhone := "+420111" + time.Now().Format("150405")
	user, _, err := r.store.FindOrCreateUser(ctx, testPhone)
	if err != nil {
		t.Fatalf("FindOrCreateUser failed: %v", err)
	}

	// Create auth context
	authUser := &AuthUser{
		ID:       user.ID,
		TenantID: nil, // No tenant yet
		Phone:    testPhone,
	}
	reqCtx := context.WithValue(ctx, userContextKey, authUser)

	// Complete onboarding
	body := `{"name": "Test User"}`
	req := httptest.NewRequest(http.MethodPost, "/api/onboarding/complete", strings.NewReader(body))
	req = req.WithContext(reqCtx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.handleCompleteOnboarding(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Check response contains tenant
	if resp["tenant"] == nil {
		t.Error("response should contain tenant")
	}
	tenant := resp["tenant"].(map[string]any)
	if tenant["name"] != "Test User" {
		t.Errorf("tenant name = %q, want %q", tenant["name"], "Test User")
	}

	// Check response contains token
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("response should contain token")
	}

	// Cleanup using db directly
	tenantID := tenant["id"].(string)
	_, _ = db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenantID)
}

func TestCompleteOnboardingWithOrphanedTenant(t *testing.T) {
	r, db, cleanup := getTestRouterWithDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test user
	testPhone := "+420222" + time.Now().Format("150405")
	user, _, err := r.store.FindOrCreateUser(ctx, testPhone)
	if err != nil {
		t.Fatalf("FindOrCreateUser failed: %v", err)
	}

	// Assign user to a fake tenant ID (simulates orphaned reference)
	orphanedTenantID := "orphaned-tenant-" + time.Now().Format("150405")
	_, err = db.Exec(ctx, "UPDATE users SET tenant_id = $1 WHERE id = $2", orphanedTenantID, user.ID)
	if err != nil {
		t.Fatalf("failed to set orphaned tenant_id: %v", err)
	}

	// Verify user has orphaned tenant_id
	updatedUser, _ := r.store.GetUserByID(ctx, user.ID)
	if updatedUser.TenantID == nil || *updatedUser.TenantID != orphanedTenantID {
		t.Fatalf("user should have orphaned tenant_id, got %v", updatedUser.TenantID)
	}

	// Create auth context with orphaned tenant_id
	authUser := &AuthUser{
		ID:       user.ID,
		TenantID: &orphanedTenantID,
		Phone:    testPhone,
	}
	reqCtx := context.WithValue(ctx, userContextKey, authUser)

	// Complete onboarding - should succeed since orphaned tenant is cleared
	body := `{"name": "Orphan Test User"}`
	req := httptest.NewRequest(http.MethodPost, "/api/onboarding/complete", strings.NewReader(body))
	req = req.WithContext(reqCtx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.handleCompleteOnboarding(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Check response contains new tenant
	if resp["tenant"] == nil {
		t.Error("response should contain tenant")
	}
	tenant := resp["tenant"].(map[string]any)
	if tenant["name"] != "Orphan Test User" {
		t.Errorf("tenant name = %q, want %q", tenant["name"], "Orphan Test User")
	}

	// Verify user's tenant_id was updated to new tenant
	finalUser, _ := r.store.GetUserByID(ctx, user.ID)
	newTenantID := tenant["id"].(string)
	if finalUser.TenantID == nil || *finalUser.TenantID != newTenantID {
		t.Errorf("user tenant_id = %v, want %q", finalUser.TenantID, newTenantID)
	}

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", newTenantID)
}

func TestCompleteOnboardingAlreadyOnboarded(t *testing.T) {
	r, db, cleanup := getTestRouterWithDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test user and tenant
	testPhone := "+420333" + time.Now().Format("150405")
	user, _, err := r.store.FindOrCreateUser(ctx, testPhone)
	if err != nil {
		t.Fatalf("FindOrCreateUser failed: %v", err)
	}

	tenant, err := r.store.CreateTenant(ctx, "Existing Tenant", "Test prompt")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	err = r.store.AssignUserToTenant(ctx, user.ID, tenant.ID)
	if err != nil {
		t.Fatalf("AssignUserToTenant failed: %v", err)
	}

	// Create auth context with existing tenant
	authUser := &AuthUser{
		ID:       user.ID,
		TenantID: &tenant.ID,
		Phone:    testPhone,
	}
	reqCtx := context.WithValue(ctx, userContextKey, authUser)

	// Try to complete onboarding - should fail since already onboarded
	body := `{"name": "Should Not Work"}`
	req := httptest.NewRequest(http.MethodPost, "/api/onboarding/complete", strings.NewReader(body))
	req = req.WithContext(reqCtx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.handleCompleteOnboarding(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["error"] != "already onboarded" {
		t.Errorf("error = %q, want %q", resp["error"], "already onboarded")
	}

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
}
