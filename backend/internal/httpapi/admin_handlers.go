package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5"
)

// withAdmin is middleware that requires admin authentication.
// It wraps withAuth and additionally checks if the user's phone is in the admin list.
func (r *Router) withAdmin(next http.HandlerFunc) http.HandlerFunc {
	return r.withAuth(func(w http.ResponseWriter, req *http.Request) {
		authUser := getAuthUser(req.Context())
		if authUser == nil {
			http.Error(w, `{"error": "not authenticated"}`, http.StatusUnauthorized)
			return
		}

		// Check if user's phone is in admin list
		if !slices.Contains(r.cfg.AdminPhones, authUser.Phone) {
			http.Error(w, `{"error": "admin access required"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, req)
	})
}

// handleAdminListPhoneNumbers returns all phone numbers with tenant info.
func (r *Router) handleAdminListPhoneNumbers(w http.ResponseWriter, req *http.Request) {
	numbers, err := r.store.ListAllPhoneNumbers(req.Context())
	if err != nil {
		r.logger.Printf("admin: failed to list phone numbers: %v", err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to list phone numbers"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"phone_numbers": numbers})
}

// handleAdminAddPhoneNumber adds a new phone number to the pool.
func (r *Router) handleAdminAddPhoneNumber(w http.ResponseWriter, req *http.Request) {
	var body struct {
		TwilioNumber string  `json:"twilio_number"`
		TwilioSID    *string `json:"twilio_sid"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate E.164 format
	if !isValidE164(body.TwilioNumber) {
		http.Error(w, `{"error": "invalid phone format, use E.164 (e.g. +420123456789)"}`, http.StatusBadRequest)
		return
	}

	pn, err := r.store.AddPhoneNumberToPool(req.Context(), body.TwilioNumber, body.TwilioSID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			http.Error(w, `{"error": "phone number already exists"}`, http.StatusConflict)
			return
		}
		r.logger.Printf("admin: failed to add phone number: %v", err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to add phone number"}`, http.StatusInternalServerError)
		return
	}

	r.logger.Printf("admin: added phone number %s to pool", body.TwilioNumber)
	writeJSON(w, http.StatusCreated, pn)
}

// handleAdminDeletePhoneNumber removes a phone number from the system.
func (r *Router) handleAdminDeletePhoneNumber(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "missing id"}`, http.StatusBadRequest)
		return
	}

	err := r.store.DeletePhoneNumber(req.Context(), id)
	if err != nil {
		r.logger.Printf("admin: failed to delete phone number %s: %v", id, err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to delete phone number"}`, http.StatusInternalServerError)
		return
	}

	r.logger.Printf("admin: deleted phone number %s", id)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// handleAdminListTenants returns all tenants for admin dropdowns.
func (r *Router) handleAdminListTenants(w http.ResponseWriter, req *http.Request) {
	tenants, err := r.store.ListAllTenants(req.Context())
	if err != nil {
		r.logger.Printf("admin: failed to list tenants: %v", err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to list tenants"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"tenants": tenants})
}

// handleAdminUpdatePhoneNumber updates a phone number's assignment.
func (r *Router) handleAdminUpdatePhoneNumber(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "missing id"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		TenantID *string `json:"tenant_id"` // nil or empty string = unassign
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Convert empty string to nil for unassignment
	var tenantID *string
	if body.TenantID != nil && *body.TenantID != "" {
		tenantID = body.TenantID
	}

	err := r.store.UpdatePhoneNumber(req.Context(), id, tenantID)
	if err != nil {
		r.logger.Printf("admin: failed to update phone number %s: %v", id, err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to update phone number"}`, http.StatusInternalServerError)
		return
	}

	r.logger.Printf("admin: updated phone number %s", id)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// handleAdminListCalls returns all calls with pagination (for admin debugging).
func (r *Router) handleAdminListCalls(w http.ResponseWriter, req *http.Request) {
	limit := 100
	if l := req.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	calls, err := r.store.ListCalls(req.Context(), limit)
	if err != nil {
		r.logger.Printf("admin: failed to list calls: %v", err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to list calls"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"calls": calls})
}

// handleAdminGetCallDetail returns call details with transcript (utterances) for a specific call.
func (r *Router) handleAdminGetCallDetail(w http.ResponseWriter, req *http.Request) {
	providerCallID := req.PathValue("providerCallId")
	if providerCallID == "" {
		http.Error(w, `{"error": "missing call ID"}`, http.StatusBadRequest)
		return
	}

	call, err := r.store.GetCallDetail(req.Context(), providerCallID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error": "call not found"}`, http.StatusNotFound)
			return
		}
		r.logger.Printf("admin: failed to get call detail %s: %v", providerCallID, err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to get call detail"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, call)
}

// handleAdminGetCallEvents returns events for a specific call.
func (r *Router) handleAdminGetCallEvents(w http.ResponseWriter, req *http.Request) {
	providerCallID := req.PathValue("providerCallId")
	if providerCallID == "" {
		http.Error(w, `{"error": "missing call ID"}`, http.StatusBadRequest)
		return
	}

	// Get internal call ID
	callID, err := r.store.GetCallID(req.Context(), providerCallID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error": "call not found"}`, http.StatusNotFound)
			return
		}
		r.logger.Printf("admin: failed to get call ID %s: %v", providerCallID, err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to get call"}`, http.StatusInternalServerError)
		return
	}

	events, err := r.store.ListCallEvents(req.Context(), callID, 1000)
	if err != nil {
		r.logger.Printf("admin: failed to list events for %s: %v", callID, err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to list events"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"events": events})
}

// ============================================================================
// Admin Users Dashboard Handlers
// ============================================================================

// handleAdminListTenantsWithDetails returns all tenants with full config and counts.
func (r *Router) handleAdminListTenantsWithDetails(w http.ResponseWriter, req *http.Request) {
	tenants, err := r.store.ListAllTenantsWithDetails(req.Context())
	if err != nil {
		r.logger.Printf("admin: failed to list tenants with details: %v", err)
		http.Error(w, `{"error": "failed to list tenants"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"tenants": tenants})
}

// handleAdminGetTenantUsers returns users for a specific tenant.
func (r *Router) handleAdminGetTenantUsers(w http.ResponseWriter, req *http.Request) {
	tenantID := req.PathValue("tenantId")
	if tenantID == "" {
		http.Error(w, `{"error": "missing tenant ID"}`, http.StatusBadRequest)
		return
	}

	users, err := r.store.ListUsersByTenant(req.Context(), tenantID)
	if err != nil {
		r.logger.Printf("admin: failed to list users for tenant %s: %v", tenantID, err)
		http.Error(w, `{"error": "failed to list users"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

// handleAdminGetTenantCalls returns calls with transcripts for a specific tenant.
func (r *Router) handleAdminGetTenantCalls(w http.ResponseWriter, req *http.Request) {
	tenantID := req.PathValue("tenantId")
	if tenantID == "" {
		http.Error(w, `{"error": "missing tenant ID"}`, http.StatusBadRequest)
		return
	}

	limit := 20
	if l := req.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	calls, err := r.store.ListCallsByTenantWithDetails(req.Context(), tenantID, limit)
	if err != nil {
		r.logger.Printf("admin: failed to list calls for tenant %s: %v", tenantID, err)
		http.Error(w, `{"error": "failed to list calls"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"calls": calls})
}

// handleAdminUpdateTenant updates a tenant's plan, status, and config settings.
func (r *Router) handleAdminUpdateTenant(w http.ResponseWriter, req *http.Request) {
	tenantID := req.PathValue("tenantId")
	if tenantID == "" {
		http.Error(w, `{"error": "missing tenant ID"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		Plan             string `json:"plan"`
		Status           string `json:"status"`
		MaxTurnTimeoutMs *int   `json:"max_turn_timeout_ms,omitempty"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate plan
	validPlans := []string{"trial", "basic", "pro"}
	if !slices.Contains(validPlans, body.Plan) {
		http.Error(w, `{"error": "invalid plan, must be trial, basic, or pro"}`, http.StatusBadRequest)
		return
	}

	// Validate status
	validStatuses := []string{"active", "suspended", "cancelled"}
	if !slices.Contains(validStatuses, body.Status) {
		http.Error(w, `{"error": "invalid status, must be active, suspended, or cancelled"}`, http.StatusBadRequest)
		return
	}

	// Validate max_turn_timeout_ms if provided
	if body.MaxTurnTimeoutMs != nil {
		if *body.MaxTurnTimeoutMs < 1000 || *body.MaxTurnTimeoutMs > 15000 {
			http.Error(w, `{"error": "max_turn_timeout_ms must be between 1000 and 15000"}`, http.StatusBadRequest)
			return
		}
	}

	rowsAffected, err := r.store.UpdateTenantPlanStatus(req.Context(), tenantID, body.Plan, body.Status)
	if err != nil {
		r.logger.Printf("admin: failed to update tenant %s: %v", tenantID, err)
		http.Error(w, `{"error": "failed to update tenant"}`, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, `{"error": "tenant not found"}`, http.StatusNotFound)
		return
	}

	// Update max_turn_timeout_ms if provided
	if body.MaxTurnTimeoutMs != nil {
		err := r.store.UpdateTenant(req.Context(), tenantID, map[string]any{
			"max_turn_timeout_ms": *body.MaxTurnTimeoutMs,
		})
		if err != nil {
			r.logger.Printf("admin: failed to update tenant config %s: %v", tenantID, err)
			http.Error(w, `{"error": "failed to update tenant config"}`, http.StatusInternalServerError)
			return
		}
	}

	r.logger.Printf("admin: updated tenant %s plan=%s status=%s max_turn_timeout_ms=%v",
		tenantID, body.Plan, body.Status, body.MaxTurnTimeoutMs)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// handleAdminResetUserOnboarding resets a user's onboarding status.
// This is a "smart reset" that preserves the tenant for data/call history
// but releases phone numbers and clears the user's tenant_id and name.
func (r *Router) handleAdminResetUserOnboarding(w http.ResponseWriter, req *http.Request) {
	userID := req.PathValue("userId")
	if userID == "" {
		http.Error(w, `{"error": "missing user ID"}`, http.StatusBadRequest)
		return
	}

	// Verify user exists
	user, err := r.store.GetUserByID(req.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error": "user not found"}`, http.StatusNotFound)
			return
		}
		r.logger.Printf("admin: failed to get user %s: %v", userID, err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to get user"}`, http.StatusInternalServerError)
		return
	}

	// Check if user is already not onboarded
	if user.TenantID == nil {
		http.Error(w, `{"error": "user is not onboarded"}`, http.StatusBadRequest)
		return
	}

	// Perform the reset
	previousTenantID, err := r.store.ResetUserOnboarding(req.Context(), userID)
	if err != nil {
		r.logger.Printf("admin: failed to reset onboarding for user %s: %v", userID, err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to reset onboarding"}`, http.StatusInternalServerError)
		return
	}

	tenantIDStr := "nil"
	if previousTenantID != nil {
		tenantIDStr = *previousTenantID
	}
	r.logger.Printf("admin: reset onboarding for user %s (phone: %s, previous tenant: %s)",
		userID, user.Phone, tenantIDStr)

	writeJSON(w, http.StatusOK, map[string]any{
		"success":            true,
		"previous_tenant_id": previousTenantID,
	})
}

// handleAdminDeleteTenant deletes a tenant and all associated data.
// This permanently removes the tenant, all users, phone numbers, calls, and sessions.
func (r *Router) handleAdminDeleteTenant(w http.ResponseWriter, req *http.Request) {
	tenantID := req.PathValue("tenantId")
	if tenantID == "" {
		http.Error(w, `{"error": "missing tenant ID"}`, http.StatusBadRequest)
		return
	}

	err := r.store.DeleteTenant(req.Context(), tenantID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error": "tenant not found"}`, http.StatusNotFound)
			return
		}
		r.logger.Printf("admin: failed to delete tenant %s: %v", tenantID, err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to delete tenant"}`, http.StatusInternalServerError)
		return
	}

	r.logger.Printf("admin: deleted tenant %s and all associated data", tenantID)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
