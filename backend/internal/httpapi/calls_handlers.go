package httpapi

import (
	"net/http"
	"strings"
)

func (r *Router) handleListCalls(w http.ResponseWriter, req *http.Request) {
	calls, err := r.store.ListCalls(req.Context(), 100)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, calls)
}

func (r *Router) handleGetCall(w http.ResponseWriter, req *http.Request) {
	// path: /api/calls/{providerCallId}
	id := strings.TrimPrefix(req.URL.Path, "/api/calls/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	call, err := r.store.GetCallDetail(req.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, call)
}


