package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/lukasbauer/karen/internal/eventlog"
	"github.com/lukasbauer/karen/internal/notifications"
	"github.com/lukasbauer/karen/internal/store"
)

type RouterConfig struct {
	PublicBaseURL string

	// Twilio credentials
	TwilioAuthToken       string
	TwilioAccountSID      string
	TwilioVerifyServiceID string

	// Voice AI providers
	DeepgramAPIKey   string
	OpenAIAPIKey     string
	ElevenLabsAPIKey string

	// STT settings
	STTEndpointingMs  int // Deepgram endpointing in ms (silence threshold)
	STTUtteranceEndMs int // Hard timeout after last speech, regardless of noise

	// Voice settings (defaults, can be overridden by tenant)
	GreetingText  string
	TTSVoiceID    string
	TTSStability  float64 // ElevenLabs voice stability (0.0-1.0)
	TTSSimilarity float64 // ElevenLabs voice similarity boost (0.0-1.0)

	// Shared HTTP client for TTS (connection pooling)
	TTSHTTPClient *http.Client

	// JWT Authentication
	JWTSecret string
	JWTExpiry time.Duration

	// Admin access (phone numbers that have admin privileges)
	AdminPhones []string

	// Notifications
	DiscordWebhookURL string

	// APNs Push Notifications
	APNsKeyPath    string // Path to .p8 key file
	APNsKeyID      string // Key ID from Apple Developer Portal
	APNsTeamID     string // Team ID from Apple Developer Portal
	APNsBundleID   string // App bundle ID (e.g., cz.zvednu.app)
	APNsProduction bool   // Use production environment

	// AI Debug API (for Claude CLI remote debugging)
	AIDebugAPIKey string // API key for AI debug endpoints
}

type Router struct {
	cfg      RouterConfig
	logger   *log.Logger
	store    *store.Store
	eventLog *eventlog.Logger
	discord  *notifications.Discord
	apns     *notifications.APNsClient
	calls    *CallRegistry
	mux      *http.ServeMux
}

func NewRouter(cfg RouterConfig, logger *log.Logger, s *store.Store, eventLog *eventlog.Logger, calls *CallRegistry) http.Handler {
	// Initialize APNs client (may be nil if not configured)
	apnsClient, err := notifications.NewAPNsClient(notifications.APNsConfig{
		KeyPath:    cfg.APNsKeyPath,
		KeyID:      cfg.APNsKeyID,
		TeamID:     cfg.APNsTeamID,
		BundleID:   cfg.APNsBundleID,
		Production: cfg.APNsProduction,
	}, logger)
	if err != nil {
		logger.Printf("Warning: APNs client initialization failed: %v", err)
	}

	r := &Router{
		cfg:      cfg,
		logger:   logger,
		store:    s,
		eventLog: eventLog,
		discord:  notifications.NewDiscord(cfg.DiscordWebhookURL, logger),
		apns:     apnsClient,
		calls:    calls,
		mux:      http.NewServeMux(),
	}

	r.routes()
	return withSentryRecovery(withCORS(r.mux))
}

func (r *Router) routes() {
	// Health check
	r.mux.HandleFunc("GET /healthz", r.handleHealthz)
	r.mux.HandleFunc("GET /readyz", r.handleReadyz)

	// Twilio webhooks (no auth - signature verified)
	r.mux.HandleFunc("POST /telephony/inbound", r.handleTwilioInbound)
	r.mux.HandleFunc("POST /telephony/status", r.handleTwilioStatus)
	r.mux.HandleFunc("GET /media", r.handleMediaWS)

	// Auth endpoints (public)
	r.mux.HandleFunc("POST /auth/send-code", r.handleSendCode)
	r.mux.HandleFunc("POST /auth/verify-code", r.handleVerifyCode)
	r.mux.HandleFunc("POST /auth/refresh", r.handleRefreshToken)
	r.mux.HandleFunc("POST /auth/logout", r.withAuth(r.handleLogout))

	// Protected API endpoints
	r.mux.HandleFunc("GET /api/me", r.withAuth(r.handleGetMe))
	r.mux.HandleFunc("GET /api/calls", r.withAuth(r.handleListCalls))
	r.mux.HandleFunc("GET /api/calls/unresolved-count", r.withAuth(r.handleGetUnresolvedCount))
	r.mux.HandleFunc("GET /api/calls/", r.withAuth(r.handleGetCall))
	r.mux.HandleFunc("PATCH /api/calls/", r.withAuth(r.handleCallPatch))
	r.mux.HandleFunc("DELETE /api/calls/", r.withAuth(r.handleCallDelete))
	r.mux.HandleFunc("GET /api/tenant", r.withAuth(r.handleGetTenant))
	r.mux.HandleFunc("PATCH /api/tenant", r.withAuth(r.handleUpdateTenant))
	r.mux.HandleFunc("GET /api/billing", r.withAuth(r.handleGetBilling))

	// Onboarding (protected)
	r.mux.HandleFunc("POST /api/onboarding/complete", r.withAuth(r.handleCompleteOnboarding))

	// Push notifications (protected)
	r.mux.HandleFunc("POST /api/push/register", r.withAuth(r.handlePushRegister))
	r.mux.HandleFunc("POST /api/push/unregister", r.withAuth(r.handlePushUnregister))

	// Billing endpoints (protected)
	r.mux.HandleFunc("POST /api/billing/checkout", r.withAuth(r.handleCreateCheckout))
	r.mux.HandleFunc("POST /api/billing/portal", r.withAuth(r.handleCreatePortal))

	// Voice selection endpoints (protected)
	r.mux.HandleFunc("GET /api/voices", r.withAuth(r.handleListVoices))
	r.mux.HandleFunc("POST /api/voices/preview", r.withAuth(r.handlePreviewVoice))

	// Stripe webhook (no auth - signature verified)
	r.mux.HandleFunc("POST /webhooks/stripe", r.handleStripeWebhook)

	// Admin endpoints (requires admin phone)
	r.mux.HandleFunc("GET /admin/phone-numbers", r.withAdmin(r.handleAdminListPhoneNumbers))
	r.mux.HandleFunc("POST /admin/phone-numbers", r.withAdmin(r.handleAdminAddPhoneNumber))
	r.mux.HandleFunc("DELETE /admin/phone-numbers/{id}", r.withAdmin(r.handleAdminDeletePhoneNumber))
	r.mux.HandleFunc("PATCH /admin/phone-numbers/{id}", r.withAdmin(r.handleAdminUpdatePhoneNumber))
	r.mux.HandleFunc("GET /admin/tenants", r.withAdmin(r.handleAdminListTenants))

	// Admin call logs (for debugging)
	r.mux.HandleFunc("GET /admin/calls", r.withAdmin(r.handleAdminListCalls))
	r.mux.HandleFunc("GET /admin/calls/{providerCallId}", r.withAdmin(r.handleAdminGetCallDetail))
	r.mux.HandleFunc("GET /admin/calls/{providerCallId}/events", r.withAdmin(r.handleAdminGetCallEvents))

	// Admin users dashboard
	r.mux.HandleFunc("GET /admin/tenants/details", r.withAdmin(r.handleAdminListTenantsWithDetails))
	r.mux.HandleFunc("GET /admin/tenants/{tenantId}/users", r.withAdmin(r.handleAdminGetTenantUsers))
	r.mux.HandleFunc("GET /admin/tenants/{tenantId}/calls", r.withAdmin(r.handleAdminGetTenantCalls))
	r.mux.HandleFunc("GET /admin/tenants/{tenantId}/costs", r.withAdmin(r.handleAdminGetTenantCosts))
	r.mux.HandleFunc("PATCH /admin/tenants/{tenantId}", r.withAdmin(r.handleAdminUpdateTenant))
	r.mux.HandleFunc("DELETE /admin/tenants/{tenantId}", r.withAdmin(r.handleAdminDeleteTenant))
	r.mux.HandleFunc("PATCH /admin/users/{userId}/reset-onboarding", r.withAdmin(r.handleAdminResetUserOnboarding))

	// Global config (admin only)
	r.mux.HandleFunc("GET /admin/config", r.withAdmin(r.handleAdminListGlobalConfig))
	r.mux.HandleFunc("PATCH /admin/config/{key}", r.withAdmin(r.handleAdminUpdateGlobalConfig))

	// Notification audit logs (admin only)
	r.mux.HandleFunc("GET /admin/notification-logs", r.withAdmin(r.handleAdminListNotificationLogs))

	// AI Debug API (for Claude CLI remote debugging)
	r.mux.HandleFunc("GET /ai/health", r.handleAIHealth) // Unauthenticated health check
	r.mux.HandleFunc("GET /ai/calls", r.withAIKey(r.handleAIListCalls))
	r.mux.HandleFunc("GET /ai/calls/{callSid}/events", r.withAIKey(r.handleAIGetCallEvents))
	r.mux.HandleFunc("GET /ai/tenants/{tenantId}/calls", r.withAIKey(r.handleAIGetTenantCalls))
	r.mux.HandleFunc("GET /ai/stats", r.withAIKey(r.handleAIGetStats))
	r.mux.HandleFunc("GET /ai/config", r.withAIKey(r.handleAIListConfig))
	r.mux.HandleFunc("PATCH /ai/config/{key}", r.withAIKey(r.handleAIUpdateConfig))
}

func (r *Router) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (r *Router) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	if r.calls.IsDraining() {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("draining"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func withSentryRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				hub := sentry.CurrentHub().Clone()
				hub.Scope().SetRequest(req)
				hub.RecoverWithContext(req.Context(), err)
				hub.Flush(2 * time.Second)
				http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, req)
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,X-API-Key")
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, req)
	})
}

// withAIKey is middleware that requires a valid AI Debug API key.
// It checks for the API key in the X-API-Key header or Authorization: Bearer header.
func (r *Router) withAIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Check if AI Debug API is configured
		if r.cfg.AIDebugAPIKey == "" {
			http.Error(w, `{"error": "AI Debug API not configured"}`, http.StatusServiceUnavailable)
			return
		}

		// Get API key from X-API-Key header or Authorization: Bearer header
		apiKey := req.Header.Get("X-API-Key")
		if apiKey == "" {
			authHeader := req.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					apiKey = parts[1]
				}
			}
		}

		if apiKey == "" {
			http.Error(w, `{"error": "missing API key - provide via X-API-Key header or Authorization: Bearer <key>"}`, http.StatusUnauthorized)
			return
		}

		// Validate API key using constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(r.cfg.AIDebugAPIKey)) != 1 {
			http.Error(w, `{"error": "invalid API key"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, req)
	}
}

func nowUTC() time.Time { return time.Now().UTC() }

// captureError sends an error to Sentry with request context
func captureError(req *http.Request, err error, msg string) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetRequest(req)
		scope.SetExtra("message", msg)
		sentry.CaptureException(err)
	})
}

func wsURLFromPublicBase(publicBase string) string {
	// http://x -> ws://x
	// https://x -> wss://x
	if strings.HasPrefix(publicBase, "https://") {
		return "wss://" + strings.TrimPrefix(publicBase, "https://")
	}
	if strings.HasPrefix(publicBase, "http://") {
		return "ws://" + strings.TrimPrefix(publicBase, "http://")
	}
	// assume already host[:port]
	return "wss://" + publicBase
}
