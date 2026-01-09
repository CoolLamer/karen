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

	call, err := r.store.GetCallDetail(req.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
		return
	}

	// TODO: Add tenant verification - check if call belongs to user's tenant

	writeJSON(w, http.StatusOK, call)
}


