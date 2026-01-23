package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5"
)

// ============================================================================
// AI Debug API Handlers
// These endpoints are designed for Claude CLI to remotely query call logs
// and update configuration settings for debugging purposes.
// ============================================================================

// handleAIListCalls returns recent calls with basic info.
// Query params: ?tenant_id=, ?limit=, ?since= (ISO timestamp)
func (r *Router) handleAIListCalls(w http.ResponseWriter, req *http.Request) {
	limit := 20
	if l := req.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// For now, use the existing ListCalls which returns all calls
	// TODO: Add filtering by tenant_id and since timestamp
	calls, err := r.store.ListCalls(req.Context(), limit)
	if err != nil {
		r.logger.Printf("ai: failed to list calls: %v", err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to list calls"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"calls": calls,
		"count": len(calls),
	})
}

// handleAIGetCallEvents returns all events for a specific call by provider call SID.
func (r *Router) handleAIGetCallEvents(w http.ResponseWriter, req *http.Request) {
	callSid := req.PathValue("callSid")
	if callSid == "" {
		http.Error(w, `{"error": "missing call SID"}`, http.StatusBadRequest)
		return
	}

	// Get internal call ID from provider call SID
	callID, err := r.store.GetCallID(req.Context(), callSid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error": "call not found"}`, http.StatusNotFound)
			return
		}
		r.logger.Printf("ai: failed to get call ID for %s: %v", callSid, err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to get call"}`, http.StatusInternalServerError)
		return
	}

	// Get call details
	call, err := r.store.GetCallDetail(req.Context(), callSid)
	if err != nil {
		r.logger.Printf("ai: failed to get call detail for %s: %v", callSid, err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to get call detail"}`, http.StatusInternalServerError)
		return
	}

	// Get events
	events, err := r.store.ListCallEvents(req.Context(), callID, 1000)
	if err != nil {
		r.logger.Printf("ai: failed to list events for %s: %v", callID, err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to list events"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"call":   call,
		"events": events,
	})
}

// handleAIGetTenantCalls returns calls for a specific tenant with events.
// Query params: ?limit=, ?since= (ISO timestamp)
func (r *Router) handleAIGetTenantCalls(w http.ResponseWriter, req *http.Request) {
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

	// Get calls for this tenant
	calls, err := r.store.ListCallsByTenantWithDetails(req.Context(), tenantID, limit)
	if err != nil {
		r.logger.Printf("ai: failed to list calls for tenant %s: %v", tenantID, err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to list tenant calls"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"calls":     calls,
		"count":     len(calls),
		"tenant_id": tenantID,
	})
}

// handleAIGetStats returns aggregate statistics for debugging.
// Query params: ?tenant_id=, ?since= (ISO timestamp)
func (r *Router) handleAIGetStats(w http.ResponseWriter, req *http.Request) {
	// Get time range
	since := time.Now().Add(-24 * time.Hour) // Default: last 24 hours
	if s := req.URL.Query().Get("since"); s != "" {
		if parsed, err := time.Parse(time.RFC3339, s); err == nil {
			since = parsed
		}
	}

	// Get event statistics from database
	stats, err := r.store.GetEventStats(req.Context(), since)
	if err != nil {
		r.logger.Printf("ai: failed to get event stats: %v", err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to get stats"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stats": stats,
		"since": since.Format(time.RFC3339),
	})
}

// handleAIListConfig returns all global config values.
func (r *Router) handleAIListConfig(w http.ResponseWriter, req *http.Request) {
	entries, err := r.store.ListGlobalConfig(req.Context())
	if err != nil {
		r.logger.Printf("ai: failed to list global config: %v", err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to list config"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"config": entries})
}

// handleAIUpdateConfig updates a global config value.
func (r *Router) handleAIUpdateConfig(w http.ResponseWriter, req *http.Request) {
	key := req.PathValue("key")
	if key == "" {
		http.Error(w, `{"error": "missing config key"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		Value string `json:"value"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate the key exists
	_, err := r.store.GetGlobalConfig(req.Context(), key)
	if err != nil {
		http.Error(w, `{"error": "config key not found"}`, http.StatusNotFound)
		return
	}

	// Validate numeric values for known numeric keys
	numericKeys := map[string]bool{
		"max_turn_timeout_ms":            true,
		"adaptive_min_timeout_ms":        true,
		"adaptive_text_decay_rate_ms":    true,
		"adaptive_sentence_end_bonus_ms": true,
		"robocall_max_call_duration_ms":  true,
		"robocall_silence_threshold_ms":  true,
		"robocall_barge_in_threshold":    true,
		"robocall_barge_in_window_ms":    true,
		"robocall_repetition_threshold":  true,
	}
	if numericKeys[key] {
		if _, err := strconv.Atoi(body.Value); err != nil {
			http.Error(w, `{"error": "value must be a number"}`, http.StatusBadRequest)
			return
		}
	}

	// Validate boolean values for known boolean keys
	boolKeys := map[string]bool{
		"adaptive_turn_enabled":      true,
		"robocall_detection_enabled": true,
		"stt_debug_enabled":          true,
	}
	if boolKeys[key] {
		if body.Value != "true" && body.Value != "false" {
			http.Error(w, `{"error": "value must be 'true' or 'false'"}`, http.StatusBadRequest)
			return
		}
	}

	if err := r.store.SetGlobalConfig(req.Context(), key, body.Value); err != nil {
		r.logger.Printf("ai: failed to update global config %s: %v", key, err)
		sentry.CaptureException(err)
		http.Error(w, `{"error": "failed to update config"}`, http.StatusInternalServerError)
		return
	}

	r.logger.Printf("ai: updated global config %s = %s", key, body.Value)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"key":     key,
		"value":   body.Value,
	})
}
