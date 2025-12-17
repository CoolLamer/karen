package httpapi

import (
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/lukasbauer/karen/internal/store"
)

// Minimal TwiML (enough to start Media Streams).
// Twilio expects Content-Type: text/xml.
type twimlResponse struct {
	XMLName xml.Name `xml:"Response"`
	Say     *twimlSay `xml:"Say,omitempty"`
	Connect *twimlConnect `xml:"Connect,omitempty"`
}

type twimlSay struct {
	Voice string `xml:"voice,attr,omitempty"`
	Text  string `xml:",chardata"`
}

type twimlConnect struct {
	Stream twimlStream `xml:"Stream"`
}

type twimlStream struct {
	URL string `xml:"url,attr"`
}

func (r *Router) handleTwilioInbound(w http.ResponseWriter, req *http.Request) {
	// Twilio sends application/x-www-form-urlencoded by default.
	if err := req.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	callSid := req.FormValue("CallSid")
	from := req.FormValue("From")
	to := req.FormValue("To")

	if callSid == "" {
		http.Error(w, "missing CallSid", http.StatusBadRequest)
		return
	}

	// Store call record (id = provider_call_id for MVP simplicity).
	_ = r.store.UpsertCall(req.Context(), store.Call{
		Provider:        "twilio",
		ProviderCallID:  callSid,
		FromNumber:      from,
		ToNumber:        to,
		Status:          "in_progress",
		StartedAt:       nowUTC(),
	})

	// Start a media stream to our websocket.
	wsBase := wsURLFromPublicBase(r.cfg.PublicBaseURL)
	mediaURL := strings.TrimRight(wsBase, "/") + "/media?callSid=" + callSid

	resp := twimlResponse{
		Say: &twimlSay{
			Voice: "alice",
			Text:  "Hi. Please tell me what this call is about.",
		},
		Connect: &twimlConnect{
			Stream: twimlStream{URL: mediaURL},
		},
	}

	out, _ := xml.MarshalIndent(resp, "", "  ")
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	_, _ = w.Write([]byte(xml.Header))
	_, _ = w.Write(out)
}

func (r *Router) handleTwilioStatus(w http.ResponseWriter, req *http.Request) {
	_ = req.ParseForm()
	callSid := req.FormValue("CallSid")
	status := req.FormValue("CallStatus") // queued/ringing/in-progress/completed/...

	if callSid != "" && status != "" {
		_ = r.store.UpdateCallStatus(req.Context(), callSid, status, nowUTC())
	}

	w.WriteHeader(http.StatusNoContent)
}


