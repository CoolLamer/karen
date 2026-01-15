package httpapi

import (
	"encoding/json"
	"net/http"
)

// handlePushRegister registers a device push token
func (r *Router) handlePushRegister(w http.ResponseWriter, req *http.Request) {
	user := getAuthUser(req.Context())
	if user == nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var body struct {
		Token    string `json:"token"`
		Platform string `json:"platform"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if body.Token == "" {
		http.Error(w, `{"error": "token is required"}`, http.StatusBadRequest)
		return
	}

	if body.Platform != "ios" && body.Platform != "android" {
		http.Error(w, `{"error": "platform must be 'ios' or 'android'"}`, http.StatusBadRequest)
		return
	}

	if err := r.store.RegisterPushToken(req.Context(), user.ID, body.Token, body.Platform); err != nil {
		r.logger.Printf("push: failed to register token: %v", err)
		http.Error(w, `{"error": "failed to register token"}`, http.StatusInternalServerError)
		return
	}

	r.logger.Printf("push: registered %s token for user %s", body.Platform, user.ID)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// handlePushUnregister removes a device push token
func (r *Router) handlePushUnregister(w http.ResponseWriter, req *http.Request) {
	user := getAuthUser(req.Context())
	if user == nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var body struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if body.Token == "" {
		http.Error(w, `{"error": "token is required"}`, http.StatusBadRequest)
		return
	}

	if err := r.store.UnregisterPushToken(req.Context(), body.Token); err != nil {
		r.logger.Printf("push: failed to unregister token: %v", err)
		http.Error(w, `{"error": "failed to unregister token"}`, http.StatusInternalServerError)
		return
	}

	r.logger.Printf("push: unregistered token for user %s", user.ID)
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
