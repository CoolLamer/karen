package httpapi

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCuratedVoices(t *testing.T) {
	// Verify curatedVoices is not empty
	if len(curatedVoices) == 0 {
		t.Error("curatedVoices should not be empty")
	}

	// Verify all voices have required fields
	for i, voice := range curatedVoices {
		if voice.ID == "" {
			t.Errorf("voice[%d] ID is empty", i)
		}
		if voice.Name == "" {
			t.Errorf("voice[%d] Name is empty", i)
		}
		if voice.Description == "" {
			t.Errorf("voice[%d] Description is empty", i)
		}
		if voice.Gender != "male" && voice.Gender != "female" {
			t.Errorf("voice[%d] Gender = %q, want 'male' or 'female'", i, voice.Gender)
		}
	}

	// Verify voiceIDSet matches curatedVoices
	if len(voiceIDSet) != len(curatedVoices) {
		t.Errorf("voiceIDSet has %d entries, curatedVoices has %d", len(voiceIDSet), len(curatedVoices))
	}

	for _, voice := range curatedVoices {
		if !voiceIDSet[voice.ID] {
			t.Errorf("voice ID %q not found in voiceIDSet", voice.ID)
		}
	}
}

func TestHandleListVoices(t *testing.T) {
	r := &Router{
		cfg:    RouterConfig{},
		logger: log.New(io.Discard, "", 0),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/voices", nil)
	rec := httptest.NewRecorder()

	r.handleListVoices(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Check content type
	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}

	// Parse response
	var resp struct {
		Voices []Voice `json:"voices"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify response matches curatedVoices
	if len(resp.Voices) != len(curatedVoices) {
		t.Errorf("response has %d voices, want %d", len(resp.Voices), len(curatedVoices))
	}

	// Verify all expected voices are present
	voiceMap := make(map[string]Voice)
	for _, v := range resp.Voices {
		voiceMap[v.ID] = v
	}

	for _, expected := range curatedVoices {
		actual, ok := voiceMap[expected.ID]
		if !ok {
			t.Errorf("expected voice %q not found in response", expected.ID)
			continue
		}
		if actual.Name != expected.Name {
			t.Errorf("voice %q name = %q, want %q", expected.ID, actual.Name, expected.Name)
		}
		if actual.Description != expected.Description {
			t.Errorf("voice %q description = %q, want %q", expected.ID, actual.Description, expected.Description)
		}
		if actual.Gender != expected.Gender {
			t.Errorf("voice %q gender = %q, want %q", expected.ID, actual.Gender, expected.Gender)
		}
	}
}

func TestHandlePreviewVoiceValidation(t *testing.T) {
	r := &Router{
		cfg:    RouterConfig{},
		logger: log.New(io.Discard, "", 0),
	}

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/voices/preview", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handlePreviewVoice(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing voice_id", func(t *testing.T) {
		body := `{}`
		req := httptest.NewRequest(http.MethodPost, "/api/voices/preview", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handlePreviewVoice(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}

		var resp map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if !strings.Contains(resp["error"], "voice_id is required") {
			t.Errorf("error = %q, should mention voice_id is required", resp["error"])
		}
	})

	t.Run("empty voice_id", func(t *testing.T) {
		body := `{"voice_id": ""}`
		req := httptest.NewRequest(http.MethodPost, "/api/voices/preview", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handlePreviewVoice(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}

		var resp map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if !strings.Contains(resp["error"], "voice_id is required") {
			t.Errorf("error = %q, should mention voice_id is required", resp["error"])
		}
	})

	t.Run("invalid voice_id not in curated list", func(t *testing.T) {
		body := `{"voice_id": "invalid-voice-id"}`
		req := httptest.NewRequest(http.MethodPost, "/api/voices/preview", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r.handlePreviewVoice(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}

		var resp map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if !strings.Contains(resp["error"], "invalid voice_id") {
			t.Errorf("error = %q, should mention invalid voice_id", resp["error"])
		}
	})
}

func TestPreviewCache(t *testing.T) {
	// Clear the cache before testing
	previewCache.Lock()
	previewCache.data = make(map[string]cachedAudio)
	previewCache.Unlock()

	// Verify cache is initially empty
	previewCache.RLock()
	if len(previewCache.data) != 0 {
		t.Errorf("cache should be empty initially, has %d entries", len(previewCache.data))
	}
	previewCache.RUnlock()

	// Verify cache duration is reasonable (at least 1 hour)
	if previewCacheDuration < 1*60*60*1e9 {
		t.Errorf("previewCacheDuration = %v, want at least 1 hour", previewCacheDuration)
	}
}

func TestPreviewText(t *testing.T) {
	// Verify preview text is not empty and is in Czech
	if previewText == "" {
		t.Error("previewText should not be empty")
	}

	// Should contain "Karen" (the assistant name)
	if !strings.Contains(previewText, "Karen") {
		t.Errorf("previewText = %q, should contain 'Karen'", previewText)
	}

	// Should be reasonable length (not too short, not too long)
	if len(previewText) < 20 {
		t.Errorf("previewText is too short: %d characters", len(previewText))
	}
	if len(previewText) > 200 {
		t.Errorf("previewText is too long: %d characters", len(previewText))
	}
}

func TestVoiceType(t *testing.T) {
	// Test Voice struct JSON marshaling
	voice := Voice{
		ID:          "test-id",
		Name:        "Test Voice",
		Description: "Test description",
		Gender:      "female",
	}

	data, err := json.Marshal(voice)
	if err != nil {
		t.Fatalf("failed to marshal Voice: %v", err)
	}

	// Verify JSON keys are correct
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"id"`) {
		t.Error("JSON should contain 'id' key")
	}
	if !strings.Contains(jsonStr, `"name"`) {
		t.Error("JSON should contain 'name' key")
	}
	if !strings.Contains(jsonStr, `"description"`) {
		t.Error("JSON should contain 'description' key")
	}
	if !strings.Contains(jsonStr, `"gender"`) {
		t.Error("JSON should contain 'gender' key")
	}

	// Test unmarshaling
	var decoded Voice
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Voice: %v", err)
	}

	if decoded.ID != voice.ID {
		t.Errorf("decoded.ID = %q, want %q", decoded.ID, voice.ID)
	}
	if decoded.Name != voice.Name {
		t.Errorf("decoded.Name = %q, want %q", decoded.Name, voice.Name)
	}
}
