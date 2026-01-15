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

func TestHandlePushRegister(t *testing.T) {
	r := &Router{
		cfg:    RouterConfig{},
		logger: log.New(io.Discard, "", 0),
	}

	t.Run("unauthorized without auth", func(t *testing.T) {
		body := `{"token": "device-token", "platform": "ios"}`
		req := httptest.NewRequest(http.MethodPost, "/api/push/register", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handlePushRegister(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		authUser := &AuthUser{ID: "user-123", Phone: "+420777123456"}
		ctx := context.WithValue(context.Background(), userContextKey, authUser)

		req := httptest.NewRequest(http.MethodPost, "/api/push/register", strings.NewReader("invalid json"))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handlePushRegister(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}

		var resp map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if !strings.Contains(resp["error"], "invalid request body") {
			t.Errorf("error = %q, should mention invalid request body", resp["error"])
		}
	})

	t.Run("missing token", func(t *testing.T) {
		authUser := &AuthUser{ID: "user-123", Phone: "+420777123456"}
		ctx := context.WithValue(context.Background(), userContextKey, authUser)

		body := `{"token": "", "platform": "ios"}`
		req := httptest.NewRequest(http.MethodPost, "/api/push/register", strings.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handlePushRegister(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}

		var resp map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if !strings.Contains(resp["error"], "token is required") {
			t.Errorf("error = %q, should mention token is required", resp["error"])
		}
	})

	t.Run("invalid platform", func(t *testing.T) {
		authUser := &AuthUser{ID: "user-123", Phone: "+420777123456"}
		ctx := context.WithValue(context.Background(), userContextKey, authUser)

		body := `{"token": "device-token", "platform": "windows"}`
		req := httptest.NewRequest(http.MethodPost, "/api/push/register", strings.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handlePushRegister(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}

		var resp map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if !strings.Contains(resp["error"], "platform must be") {
			t.Errorf("error = %q, should mention platform must be ios or android", resp["error"])
		}
	})
}

func TestHandlePushUnregister(t *testing.T) {
	r := &Router{
		cfg:    RouterConfig{},
		logger: log.New(io.Discard, "", 0),
	}

	t.Run("unauthorized without auth", func(t *testing.T) {
		body := `{"token": "device-token"}`
		req := httptest.NewRequest(http.MethodPost, "/api/push/unregister", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handlePushUnregister(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		authUser := &AuthUser{ID: "user-123", Phone: "+420777123456"}
		ctx := context.WithValue(context.Background(), userContextKey, authUser)

		req := httptest.NewRequest(http.MethodPost, "/api/push/unregister", strings.NewReader("invalid json"))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handlePushUnregister(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing token", func(t *testing.T) {
		authUser := &AuthUser{ID: "user-123", Phone: "+420777123456"}
		ctx := context.WithValue(context.Background(), userContextKey, authUser)

		body := `{"token": ""}`
		req := httptest.NewRequest(http.MethodPost, "/api/push/unregister", strings.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handlePushUnregister(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}

		var resp map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if !strings.Contains(resp["error"], "token is required") {
			t.Errorf("error = %q, should mention token is required", resp["error"])
		}
	})
}

// Integration tests (require database)
func TestPushTokensIntegration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	s := store.New(db)

	r := &Router{
		cfg: RouterConfig{
			JWTSecret: "test-secret-key",
			JWTExpiry: 1 * time.Hour,
		},
		logger: log.New(io.Discard, "", 0),
		store:  s,
	}

	// Create a test user
	testPhone := "+420444" + time.Now().Format("150405")
	user, _, err := s.FindOrCreateUser(ctx, testPhone)
	if err != nil {
		t.Fatalf("FindOrCreateUser failed: %v", err)
	}
	defer func() {
		_, _ = db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	}()

	t.Run("register and unregister token", func(t *testing.T) {
		authUser := &AuthUser{ID: user.ID, Phone: testPhone}
		reqCtx := context.WithValue(ctx, userContextKey, authUser)

		// Register a token
		body := `{"token": "test-device-token-123", "platform": "ios"}`
		req := httptest.NewRequest(http.MethodPost, "/api/push/register", strings.NewReader(body))
		req = req.WithContext(reqCtx)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handlePushRegister(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("register status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		// Verify token was registered
		tokens, err := s.GetUserPushTokens(ctx, user.ID)
		if err != nil {
			t.Fatalf("GetUserPushTokens failed: %v", err)
		}
		if len(tokens) != 1 {
			t.Fatalf("expected 1 token, got %d", len(tokens))
		}
		if tokens[0].Token != "test-device-token-123" {
			t.Errorf("token = %q, want %q", tokens[0].Token, "test-device-token-123")
		}
		if tokens[0].Platform != "ios" {
			t.Errorf("platform = %q, want %q", tokens[0].Platform, "ios")
		}

		// Unregister the token
		body = `{"token": "test-device-token-123"}`
		req = httptest.NewRequest(http.MethodPost, "/api/push/unregister", strings.NewReader(body))
		req = req.WithContext(reqCtx)
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()

		r.handlePushUnregister(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("unregister status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		// Verify token was removed
		tokens, err = s.GetUserPushTokens(ctx, user.ID)
		if err != nil {
			t.Fatalf("GetUserPushTokens failed: %v", err)
		}
		if len(tokens) != 0 {
			t.Errorf("expected 0 tokens after unregister, got %d", len(tokens))
		}
	})

	t.Run("register android token", func(t *testing.T) {
		authUser := &AuthUser{ID: user.ID, Phone: testPhone}
		reqCtx := context.WithValue(ctx, userContextKey, authUser)

		body := `{"token": "android-token-456", "platform": "android"}`
		req := httptest.NewRequest(http.MethodPost, "/api/push/register", strings.NewReader(body))
		req = req.WithContext(reqCtx)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handlePushRegister(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("register status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		tokens, err := s.GetUserPushTokens(ctx, user.ID)
		if err != nil {
			t.Fatalf("GetUserPushTokens failed: %v", err)
		}
		if len(tokens) != 1 {
			t.Fatalf("expected 1 token, got %d", len(tokens))
		}
		if tokens[0].Platform != "android" {
			t.Errorf("platform = %q, want %q", tokens[0].Platform, "android")
		}

		// Cleanup
		_, _ = db.Exec(ctx, "DELETE FROM device_push_tokens WHERE user_id = $1", user.ID)
	})
}
