package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Voice represents a curated voice option for the AI assistant
type Voice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Gender      string `json:"gender"`
}

// Curated list of voices that work well with Czech
var curatedVoices = []Voice{
	{ID: "21m00Tcm4TlvDq8ikWAM", Name: "Rachel", Description: "Přátelský, profesionální", Gender: "female"},
	{ID: "EXAVITQu4vr4xnSDxMaL", Name: "Sarah", Description: "Měkký, uklidňující", Gender: "female"},
	{ID: "pNInz6obpgDQGcFmaJgB", Name: "Adam", Description: "Hluboký, důvěryhodný", Gender: "male"},
	{ID: "ErXwobaYiN019PkySvjV", Name: "Antoni", Description: "Mladý, energický", Gender: "male"},
	{ID: "MF3mGyEYCl7XYWbV9V6O", Name: "Aria", Description: "Expresivní, přirozený", Gender: "female"},
}

// voiceIDSet for quick validation
var voiceIDSet = func() map[string]bool {
	m := make(map[string]bool)
	for _, v := range curatedVoices {
		m[v.ID] = true
	}
	return m
}()

// previewCache stores cached preview audio to reduce ElevenLabs API calls
var previewCache = struct {
	sync.RWMutex
	data map[string]cachedAudio
}{data: make(map[string]cachedAudio)}

type cachedAudio struct {
	audio     []byte
	expiresAt time.Time
}

const (
	previewCacheDuration = 24 * time.Hour
	previewText          = "Dobrý den, tady Karen. Jak vám mohu pomoci?"
	maxPreviewAudioSize  = 10 * 1024 * 1024 // 10MB max for audio response
)

// HTTP client with timeout for ElevenLabs API calls
var elevenLabsClient = &http.Client{
	Timeout: 30 * time.Second,
}

// handleListVoices returns the curated list of available voices
func (r *Router) handleListVoices(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"voices": curatedVoices,
	})
}

// handlePreviewVoice generates a preview audio clip for a voice
func (r *Router) handlePreviewVoice(w http.ResponseWriter, req *http.Request) {
	var body struct {
		VoiceID string `json:"voice_id"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if body.VoiceID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "voice_id is required"})
		return
	}

	// Validate voice ID is in our curated list
	if !voiceIDSet[body.VoiceID] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid voice_id"})
		return
	}

	// Check cache
	previewCache.RLock()
	cached, found := previewCache.data[body.VoiceID]
	previewCache.RUnlock()

	if found && time.Now().Before(cached.expiresAt) {
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(cached.audio)))
		w.Header().Set("X-Cache", "HIT")
		_, _ = w.Write(cached.audio)
		return
	}

	// Generate preview audio from ElevenLabs
	audio, err := r.generatePreviewAudio(req.Context(), body.VoiceID)
	if err != nil {
		r.logger.Printf("voice: failed to generate preview: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate preview"})
		return
	}

	// Cache the result
	previewCache.Lock()
	previewCache.data[body.VoiceID] = cachedAudio{
		audio:     audio,
		expiresAt: time.Now().Add(previewCacheDuration),
	}
	previewCache.Unlock()

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(audio)))
	w.Header().Set("X-Cache", "MISS")
	_, _ = w.Write(audio)
}

// generatePreviewAudio calls ElevenLabs API to generate MP3 audio
func (r *Router) generatePreviewAudio(ctx context.Context, voiceID string) ([]byte, error) {
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s?output_format=mp3_44100_128", voiceID)

	reqBody := map[string]any{
		"text":     previewText,
		"model_id": "eleven_flash_v2_5",
		"voice_settings": map[string]float64{
			"stability":        0.5,
			"similarity_boost": 0.75,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("xi-api-key", r.cfg.ElevenLabsAPIKey)

	resp, err := elevenLabsClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096)) // Limit error response size
		return nil, fmt.Errorf("ElevenLabs API error: %s - %s", resp.Status, string(respBody))
	}

	return io.ReadAll(io.LimitReader(resp.Body, maxPreviewAudioSize))
}
