package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/lukasbauer/karen/internal/app"
	"github.com/lukasbauer/karen/internal/store"
)

type Router struct {
	cfg    app.Config
	logger *log.Logger
	store  *store.Store
	mux    *http.ServeMux
}

func NewRouter(cfg app.Config, logger *log.Logger, s *store.Store) http.Handler {
	r := &Router{
		cfg:    cfg,
		logger: logger,
		store:  s,
		mux:    http.NewServeMux(),
	}

	r.routes()
	return withCORS(r.mux)
}

func (r *Router) routes() {
	r.mux.HandleFunc("GET /healthz", r.handleHealthz)
	r.mux.HandleFunc("POST /telephony/inbound", r.handleTwilioInbound)
	r.mux.HandleFunc("POST /telephony/status", r.handleTwilioStatus)
	r.mux.HandleFunc("GET /api/calls", r.handleListCalls)
	r.mux.HandleFunc("GET /api/calls/", r.handleGetCall)
	r.mux.HandleFunc("GET /media", r.handleMediaWS)
}

func (r *Router) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func nowUTC() time.Time { return time.Now().UTC() }

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


