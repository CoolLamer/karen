package httpapi

import (
	"net/http"
)

// MVP placeholder:
// Twilio Media Streams uses WebSocket messages containing JSON events + base64 audio.
// We'll accept the upgrade and keep the connection alive (real STT/TTS loop comes next).
func (r *Router) handleMediaWS(w http.ResponseWriter, req *http.Request) {
	// NOTE: net/http doesn't include websocket support; you typically use gorilla/websocket
	// or nhooyr.io/websocket here. For now, return a clear message so deploy works.
	http.Error(w, "websocket not implemented yet (add gorilla/websocket and handle Twilio Media Streams here)", http.StatusNotImplemented)
}


