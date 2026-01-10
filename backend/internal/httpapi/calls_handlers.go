package httpapi

import (
	"net/http"
	"strings"
)

// handleCallPatch dispatches PATCH requests for calls based on path suffix
func (r *Router) handleCallPatch(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/calls/")
	switch {
	case strings.HasSuffix(path, "/viewed"):
		r.handleMarkCallViewed(w, req)
	case strings.HasSuffix(path, "/resolve"):
		r.handleMarkCallResolved(w, req)
	default:
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
	}
}

// handleCallDelete dispatches DELETE requests for calls based on path suffix
func (r *Router) handleCallDelete(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/calls/")
	switch {
	case strings.HasSuffix(path, "/resolve"):
		r.handleMarkCallUnresolved(w, req)
	default:
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
	}
}

func (r *Router) handleListCalls(w http.ResponseWriter, req *http.Request) {
	authUser := getAuthUser(req.Context())
	if authUser == nil {
		http.Error(w, `{"error": "not authenticated"}`, http.StatusUnauthorized)
		return
	}

	// If user has a tenant, filter by tenant
	if authUser.TenantID != nil {
		calls, err := r.store.ListCallsByTenant(req.Context(), *authUser.TenantID, 100)
		if err != nil {
			http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, calls)
		return
	}

	// User without tenant sees no calls
	writeJSON(w, http.StatusOK, []any{})
}

func (r *Router) handleGetCall(w http.ResponseWriter, req *http.Request) {
	authUser := getAuthUser(req.Context())
	if authUser == nil {
		http.Error(w, `{"error": "not authenticated"}`, http.StatusUnauthorized)
		return
	}

	// path: /api/calls/{providerCallId}
	id := strings.TrimPrefix(req.URL.Path, "/api/calls/")
	if id == "" {
		http.Error(w, `{"error": "missing id"}`, http.StatusBadRequest)
		return
	}

	call, callTenantID, err := r.store.GetCallDetailWithTenantCheck(req.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
		return
	}

	// Security: verify call belongs to user's tenant
	// Return 404 (not 403) to prevent information leakage about call existence
	if authUser.TenantID == nil {
		// User has no tenant, cannot access any calls
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
		return
	}
	if callTenantID == nil || *callTenantID != *authUser.TenantID {
		// Call belongs to a different tenant or has no tenant
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, call)
}

// handleMarkCallViewed marks a call as viewed (sets first_viewed_at if NULL)
func (r *Router) handleMarkCallViewed(w http.ResponseWriter, req *http.Request) {
	authUser := getAuthUser(req.Context())
	if authUser == nil {
		http.Error(w, `{"error": "not authenticated"}`, http.StatusUnauthorized)
		return
	}

	// path: /api/calls/{providerCallId}/viewed
	path := strings.TrimPrefix(req.URL.Path, "/api/calls/")
	id := strings.TrimSuffix(path, "/viewed")
	if id == "" {
		http.Error(w, `{"error": "missing id"}`, http.StatusBadRequest)
		return
	}

	// Security: verify call belongs to user's tenant
	callTenantID, err := r.store.GetCallTenantID(req.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
		return
	}
	if authUser.TenantID == nil || callTenantID == nil || *callTenantID != *authUser.TenantID {
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
		return
	}

	_, err = r.store.MarkCallViewed(req.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// handleMarkCallResolved marks a call as resolved
func (r *Router) handleMarkCallResolved(w http.ResponseWriter, req *http.Request) {
	authUser := getAuthUser(req.Context())
	if authUser == nil {
		http.Error(w, `{"error": "not authenticated"}`, http.StatusUnauthorized)
		return
	}

	// path: /api/calls/{providerCallId}/resolve
	path := strings.TrimPrefix(req.URL.Path, "/api/calls/")
	id := strings.TrimSuffix(path, "/resolve")
	if id == "" {
		http.Error(w, `{"error": "missing id"}`, http.StatusBadRequest)
		return
	}

	// Security: verify call belongs to user's tenant
	callTenantID, err := r.store.GetCallTenantID(req.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
		return
	}
	if authUser.TenantID == nil || callTenantID == nil || *callTenantID != *authUser.TenantID {
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
		return
	}

	err = r.store.MarkCallResolved(req.Context(), id, authUser.ID)
	if err != nil {
		http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// handleMarkCallUnresolved marks a call as unresolved
func (r *Router) handleMarkCallUnresolved(w http.ResponseWriter, req *http.Request) {
	authUser := getAuthUser(req.Context())
	if authUser == nil {
		http.Error(w, `{"error": "not authenticated"}`, http.StatusUnauthorized)
		return
	}

	// path: /api/calls/{providerCallId}/resolve
	path := strings.TrimPrefix(req.URL.Path, "/api/calls/")
	id := strings.TrimSuffix(path, "/resolve")
	if id == "" {
		http.Error(w, `{"error": "missing id"}`, http.StatusBadRequest)
		return
	}

	// Security: verify call belongs to user's tenant
	callTenantID, err := r.store.GetCallTenantID(req.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
		return
	}
	if authUser.TenantID == nil || callTenantID == nil || *callTenantID != *authUser.TenantID {
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
		return
	}

	err = r.store.MarkCallUnresolved(req.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// handleGetUnresolvedCount returns the count of unresolved calls for the user's tenant
func (r *Router) handleGetUnresolvedCount(w http.ResponseWriter, req *http.Request) {
	authUser := getAuthUser(req.Context())
	if authUser == nil {
		http.Error(w, `{"error": "not authenticated"}`, http.StatusUnauthorized)
		return
	}

	if authUser.TenantID == nil {
		writeJSON(w, http.StatusOK, map[string]int{"count": 0})
		return
	}

	count, err := r.store.CountUnresolvedCalls(req.Context(), *authUser.TenantID)
	if err != nil {
		http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}


