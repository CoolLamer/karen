package httpapi

import (
	"context"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lukasbauer/karen/internal/store"
)

func TestTwiMLStructures(t *testing.T) {
	t.Run("basic TwiML response", func(t *testing.T) {
		resp := twimlResponse{
			Connect: &twimlConnect{
				Stream: twimlStream{
					URL: "wss://example.com/media",
					Parameters: []twimlParameter{
						{Name: "callSid", Value: "CA123"},
					},
				},
			},
		}

		out, err := xml.MarshalIndent(resp, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal TwiML: %v", err)
		}

		xmlStr := string(out)

		if !strings.Contains(xmlStr, "<Response>") {
			t.Error("TwiML should contain <Response>")
		}
		if !strings.Contains(xmlStr, "<Connect>") {
			t.Error("TwiML should contain <Connect>")
		}
		if !strings.Contains(xmlStr, `url="wss://example.com/media"`) {
			t.Error("TwiML should contain stream URL")
		}
		if !strings.Contains(xmlStr, `name="callSid"`) {
			t.Error("TwiML should contain callSid parameter")
		}
		if !strings.Contains(xmlStr, `value="CA123"`) {
			t.Error("TwiML should contain callSid value")
		}
	})

	t.Run("TwiML with multiple parameters", func(t *testing.T) {
		resp := twimlResponse{
			Connect: &twimlConnect{
				Stream: twimlStream{
					URL: "wss://example.com/media",
					Parameters: []twimlParameter{
						{Name: "callSid", Value: "CA123"},
						{Name: "tenantId", Value: "tenant-456"},
						{Name: "tenantConfig", Value: `{"system_prompt":"test"}`},
					},
				},
			},
		}

		out, _ := xml.MarshalIndent(resp, "", "  ")
		xmlStr := string(out)

		if strings.Count(xmlStr, "<Parameter") != 3 {
			t.Errorf("TwiML should contain 3 parameters, got: %s", xmlStr)
		}
	})
}

func TestWsURLFromPublicBase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"http://localhost:8080", "ws://localhost:8080"},
		{"https://example.com", "wss://example.com"},
		{"https://api.example.com/v1", "wss://api.example.com/v1"},
		{"example.com", "wss://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := wsURLFromPublicBase(tt.input)
			if got != tt.expected {
				t.Errorf("wsURLFromPublicBase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestHandleTwilioInboundBasic(t *testing.T) {
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
			PublicBaseURL: "https://example.com",
		},
		logger: log.New(io.Discard, "", 0),
		store:  s,
	}

	t.Run("missing CallSid", func(t *testing.T) {
		form := url.Values{}
		form.Set("From", "+420777123456")
		form.Set("To", "+420228883001")

		req := httptest.NewRequest(http.MethodPost, "/telephony/inbound", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		r.handleTwilioInbound(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid inbound call", func(t *testing.T) {
		form := url.Values{}
		form.Set("CallSid", "CA123456")
		form.Set("From", "+420777123456")
		form.Set("To", "+420228883001")

		req := httptest.NewRequest(http.MethodPost, "/telephony/inbound", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		r.handleTwilioInbound(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		body := rec.Body.String()

		// Check Content-Type
		if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/xml") {
			t.Errorf("Content-Type = %q, want text/xml", ct)
		}

		// Check TwiML structure
		if !strings.Contains(body, "<?xml") {
			t.Error("response should be XML")
		}
		if !strings.Contains(body, "<Response>") {
			t.Error("response should contain <Response>")
		}
		if !strings.Contains(body, "<Connect>") {
			t.Error("response should contain <Connect>")
		}
		if !strings.Contains(body, "wss://example.com/media") {
			t.Error("response should contain media URL")
		}
		if !strings.Contains(body, "CA123456") {
			t.Error("response should contain callSid parameter")
		}
	})

	// Cleanup any test calls
	_, _ = db.Exec(ctx, "DELETE FROM calls WHERE provider_call_id = 'CA123456'")
}

func TestHandleTwilioStatus(t *testing.T) {
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
		logger: log.New(io.Discard, "", 0),
		store:  s,
	}

	form := url.Values{}
	form.Set("CallSid", "CA123456_status")
	form.Set("CallStatus", "completed")

	req := httptest.NewRequest(http.MethodPost, "/telephony/status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	r.handleTwilioStatus(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

// Integration test with real database
func TestTenantRoutingIntegration(t *testing.T) {
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
			PublicBaseURL: "https://example.com",
		},
		logger: log.New(io.Discard, "", 0),
		store:  s,
	}

	// Create tenant
	tenant, err := s.CreateTenant(ctx, "Routing Test", "Test prompt for routing")
	if err != nil {
		t.Fatalf("CreateTenant failed: %v", err)
	}

	// Assign phone number
	twilioNumber := "+1888" + time.Now().Format("0102150405")
	err = s.AssignPhoneNumberToTenant(ctx, tenant.ID, twilioNumber, "PNTEST")
	if err != nil {
		t.Fatalf("AssignPhoneNumberToTenant failed: %v", err)
	}

	t.Run("call to tenant number includes tenant config", func(t *testing.T) {
		form := url.Values{}
		form.Set("CallSid", "CA"+time.Now().Format("150405"))
		form.Set("From", "+420777123456")
		form.Set("To", twilioNumber)

		req := httptest.NewRequest(http.MethodPost, "/telephony/inbound", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		r.handleTwilioInbound(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		body := rec.Body.String()

		// Should contain tenant ID parameter
		if !strings.Contains(body, `name="tenantId"`) {
			t.Error("response should contain tenantId parameter")
		}

		// Should contain tenant config parameter
		if !strings.Contains(body, `name="tenantConfig"`) {
			t.Error("response should contain tenantConfig parameter")
		}

		// Should contain the system prompt in config
		if !strings.Contains(body, "Test prompt for routing") {
			t.Error("response should contain tenant's system prompt in config")
		}
	})

	t.Run("call to unknown number has no tenant config", func(t *testing.T) {
		form := url.Values{}
		form.Set("CallSid", "CA"+time.Now().Format("150406"))
		form.Set("From", "+420777123456")
		form.Set("To", "+1999999999") // Unknown number

		req := httptest.NewRequest(http.MethodPost, "/telephony/inbound", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		r.handleTwilioInbound(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		body := rec.Body.String()

		// Should still work but without tenant config
		if !strings.Contains(body, "<Response>") {
			t.Error("response should still contain valid TwiML")
		}

		// Should NOT contain tenantId parameter
		if strings.Contains(body, `name="tenantId"`) {
			t.Error("response should NOT contain tenantId for unknown number")
		}
	})

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM calls WHERE to_number = $1", twilioNumber)
	_, _ = db.Exec(ctx, "DELETE FROM tenant_phone_numbers WHERE tenant_id = $1", tenant.ID)
	_, _ = db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
}
