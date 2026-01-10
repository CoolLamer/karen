package httpapi

import (
	"net/http"
	"strings"
)

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


