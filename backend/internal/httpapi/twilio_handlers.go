package httpapi

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/lukasbauer/karen/internal/store"
)

// Minimal TwiML (enough to start Media Streams).
// Twilio expects Content-Type: text/xml.
type twimlResponse struct {
	XMLName xml.Name      `xml:"Response"`
	Say     *twimlSay     `xml:"Say,omitempty"`
	Connect *twimlConnect `xml:"Connect,omitempty"`
	Reject  *twimlReject  `xml:"Reject,omitempty"`
}

type twimlSay struct {
	Voice string `xml:"voice,attr,omitempty"`
	Text  string `xml:",chardata"`
}

type twimlReject struct {
	Reason string `xml:"reason,attr,omitempty"` // "rejected" or "busy"
}

type twimlConnect struct {
	Stream twimlStream `xml:"Stream"`
}

type twimlStream struct {
	URL        string           `xml:"url,attr"`
	Parameters []twimlParameter `xml:"Parameter,omitempty"`
}

type twimlParameter struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
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
	forwardedFrom := req.FormValue("ForwardedFrom")

	if callSid == "" {
		http.Error(w, "missing CallSid", http.StatusBadRequest)
		return
	}

	// Look up tenant by Twilio number
	var tenant *store.Tenant
	var err error

	// Primary lookup: by "To" number (the Twilio number that received the call)
	tenant, err = r.store.GetTenantByTwilioNumber(req.Context(), to)
	if err != nil && forwardedFrom != "" {
		// Fallback: try forwarding source lookup
		tenant, _ = r.store.GetTenantByForwardingSource(req.Context(), forwardedFrom)
	}

	// Determine tenant ID for the call record
	var tenantID *string
	if tenant != nil {
		tenantID = &tenant.ID
		r.logger.Printf("inbound: call %s routed to tenant %s (%s)", callSid, tenant.ID, tenant.Name)
	} else {
		r.logger.Printf("inbound: call %s has no tenant (to=%s, forwardedFrom=%s)", callSid, to, forwardedFrom)
	}

	// Check if tenant has exceeded their call limit (trial expired or limit reached)
	if tenant != nil {
		callStatus := store.CanTenantReceiveCalls(tenant)
		if !callStatus.CanReceive {
			r.logger.Printf("inbound: call %s rejected for tenant %s: %s (calls: %d/%d)",
				callSid, tenant.ID, callStatus.Reason, callStatus.CallsUsed, callStatus.CallsLimit)

			// Store call record as rejected
			_ = r.store.UpsertCallWithTenant(req.Context(), store.Call{
				TenantID:       tenantID,
				Provider:       "twilio",
				ProviderCallID: callSid,
				FromNumber:     from,
				ToNumber:       to,
				Status:         "rejected_limit",
				StartedAt:      nowUTC(),
			})

			// Return TwiML that simply hangs up (don't answer, let it ring through to voicemail)
			// We don't play any message to the caller - that would be unprofessional
			resp := twimlResponse{Reject: &twimlReject{Reason: "busy"}}
			out, _ := xml.MarshalIndent(resp, "", "  ")
			w.Header().Set("Content-Type", "text/xml; charset=utf-8")
			_, _ = w.Write([]byte(xml.Header))
			_, _ = w.Write(out)
			return
		}
	}

	// Store call record with tenant ID
	_ = r.store.UpsertCallWithTenant(req.Context(), store.Call{
		TenantID:       tenantID,
		Provider:       "twilio",
		ProviderCallID: callSid,
		FromNumber:     from,
		ToNumber:       to,
		Status:         "in_progress",
		StartedAt:      nowUTC(),
	})

	// Start a media stream to our websocket.
	wsBase := wsURLFromPublicBase(r.cfg.PublicBaseURL)
	mediaURL := strings.TrimRight(wsBase, "/") + "/media"

	// Build stream parameters
	params := []twimlParameter{
		{Name: "callSid", Value: callSid},
	}

	// Pass tenant config to media stream if tenant found
	if tenant != nil {
		params = append(params, twimlParameter{Name: "tenantId", Value: tenant.ID})

		// Get tenant owner's phone number for call forwarding
		ownerPhone, _ := r.store.GetTenantOwnerPhone(req.Context(), tenant.ID)

		// Pass tenant config as JSON for the call session
		tenantConfig := map[string]any{
			"system_prompt":       tenant.SystemPrompt,
			"greeting_text":       tenant.GreetingText,
			"voice_id":            tenant.VoiceID,
			"language":            tenant.Language,
			"vip_names":           tenant.VIPNames,
			"marketing_email":     tenant.MarketingEmail,
			"forward_number":      tenant.ForwardNumber,
			"max_turn_timeout_ms": tenant.MaxTurnTimeoutMs,
			"owner_phone":         ownerPhone, // User's verified phone for forwarding
		}
		configJSON, _ := json.Marshal(tenantConfig)
		params = append(params, twimlParameter{Name: "tenantConfig", Value: string(configJSON)})
	}

	resp := twimlResponse{
		Connect: &twimlConnect{
			Stream: twimlStream{
				URL:        mediaURL,
				Parameters: params,
			},
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
