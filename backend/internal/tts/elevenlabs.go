package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const elevenLabsAPIURL = "https://api.elevenlabs.io/v1/text-to-speech"

// ElevenLabsClient implements the Client interface using ElevenLabs' API.
type ElevenLabsClient struct {
	apiKey     string
	voiceID    string
	modelID    string
	httpClient *http.Client
}

// ElevenLabsConfig holds configuration for the ElevenLabs client.
type ElevenLabsConfig struct {
	APIKey  string
	VoiceID string // ElevenLabs voice ID
	ModelID string // e.g., "eleven_flash_v2_5" for low latency
}

// NewElevenLabsClient creates a new ElevenLabs client.
func NewElevenLabsClient(cfg ElevenLabsConfig) *ElevenLabsClient {
	modelID := cfg.ModelID
	if modelID == "" {
		modelID = "eleven_flash_v2_5" // Low latency model with Czech support
	}
	voiceID := cfg.VoiceID
	if voiceID == "" {
		voiceID = "21m00Tcm4TlvDq8ikWAM" // Rachel - default voice
	}
	return &ElevenLabsClient{
		apiKey:     cfg.APIKey,
		voiceID:    voiceID,
		modelID:    modelID,
		httpClient: &http.Client{},
	}
}

// ttsRequest represents an ElevenLabs TTS request.
type ttsRequest struct {
	Text          string       `json:"text"`
	ModelID       string       `json:"model_id"`
	VoiceSettings voiceSettings `json:"voice_settings,omitempty"`
}

type voiceSettings struct {
	Stability       float64 `json:"stability"`
	SimilarityBoost float64 `json:"similarity_boost"`
}

// Synthesize converts text to speech and returns audio data in μ-law format.
func (c *ElevenLabsClient) Synthesize(ctx context.Context, text string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s?output_format=ulaw_8000", elevenLabsAPIURL, c.voiceID)

	req := ttsRequest{
		Text:    text,
		ModelID: c.modelID,
		VoiceSettings: voiceSettings{
			Stability:       0.5,
			SimilarityBoost: 0.75,
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("xi-api-key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ElevenLabs API error: %s - %s", resp.Status, string(respBody))
	}

	return io.ReadAll(resp.Body)
}

// SynthesizeStream converts text to speech and streams audio chunks.
func (c *ElevenLabsClient) SynthesizeStream(ctx context.Context, text string) (<-chan []byte, error) {
	url := fmt.Sprintf("%s/%s/stream?output_format=ulaw_8000", elevenLabsAPIURL, c.voiceID)

	req := ttsRequest{
		Text:    text,
		ModelID: c.modelID,
		VoiceSettings: voiceSettings{
			Stability:       0.5,
			SimilarityBoost: 0.75,
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("xi-api-key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ElevenLabs API error: %s - %s", resp.Status, string(respBody))
	}

	ch := make(chan []byte, 100)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		// Read in chunks of 640 bytes (80ms of μ-law audio at 8kHz)
		buf := make([]byte, 640)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				select {
				case <-ctx.Done():
					return
				case ch <- chunk:
				}
			}
			if err == io.EOF {
				return
			}
			if err != nil {
				return
			}
		}
	}()

	return ch, nil
}
