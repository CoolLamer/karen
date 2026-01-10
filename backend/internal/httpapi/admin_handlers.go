package httpapi

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
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
		http.Error(w, `{"error": "failed to delete phone number"}`, http.StatusInternalServerError)
		return
	}

	r.logger.Printf("admin: deleted phone number %s", id)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
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
		http.Error(w, `{"error": "failed to update phone number"}`, http.StatusInternalServerError)
		return
	}

	r.logger.Printf("admin: updated phone number %s", id)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
