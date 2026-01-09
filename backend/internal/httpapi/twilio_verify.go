package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// twilioVerifyResponse represents a Twilio Verify API response
type twilioVerifyResponse struct {
	Status string `json:"status"`
	Valid  bool   `json:"valid"`
	SID    string `json:"sid"`
}

// sendTwilioVerifyCode sends a verification code via Twilio Verify
func (r *Router) sendTwilioVerifyCode(ctx context.Context, phone string) error {
	apiURL := fmt.Sprintf(
		"https://verify.twilio.com/v2/Services/%s/Verifications",
		r.cfg.TwilioVerifyServiceID,
	)

	data := url.Values{}
	data.Set("To", phone)
	data.Set("Channel", "sms")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(r.cfg.TwilioAccountSID, r.cfg.TwilioAuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("Twilio API error: %d - %v", resp.StatusCode, errResp)
	}

	var verifyResp twilioVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if verifyResp.Status != "pending" {
		return fmt.Errorf("unexpected verification status: %s", verifyResp.Status)
	}

	return nil
}

// verifyTwilioCode checks a verification code with Twilio Verify
func (r *Router) verifyTwilioCode(ctx context.Context, phone, code string) (bool, error) {
	apiURL := fmt.Sprintf(
		"https://verify.twilio.com/v2/Services/%s/VerificationCheck",
		r.cfg.TwilioVerifyServiceID,
	)

	data := url.Values{}
	data.Set("To", phone)
	data.Set("Code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(r.cfg.TwilioAccountSID, r.cfg.TwilioAuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 404 means code not found or expired
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return false, fmt.Errorf("Twilio API error: %d - %v", resp.StatusCode, errResp)
	}

	var verifyResp twilioVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return verifyResp.Status == "approved" && verifyResp.Valid, nil
}
