package stt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const deepgramWSURL = "wss://api.deepgram.com/v1/listen"

// DeepgramClient implements the Client interface using Deepgram's streaming API.
type DeepgramClient struct {
	conn       *websocket.Conn
	results    chan TranscriptResult
	errors     chan error
	done       chan struct{}
	closeOnce  sync.Once
	mu         sync.Mutex
	wg         sync.WaitGroup // Wait for readLoop to finish
}

// DeepgramConfig holds configuration for the Deepgram client.
type DeepgramConfig struct {
	APIKey         string
	Language       string // e.g., "cs" for Czech
	Model          string // e.g., "nova-3"
	SampleRate     int    // e.g., 8000 for Twilio μ-law
	Encoding       string // e.g., "mulaw" for Twilio
	Channels       int    // e.g., 1 for mono
	Punctuate      bool
	Endpointing    int // milliseconds of silence for endpointing, 0 for default
	UtteranceEndMs int // hard timeout after last speech, regardless of noise (0 for default)
}

// deepgramResponse represents a Deepgram WebSocket response.
type deepgramResponse struct {
	Type    string `json:"type"`
	Channel struct {
		Alternatives []struct {
			Transcript string  `json:"transcript"`
			Confidence float64 `json:"confidence"`
		} `json:"alternatives"`
	} `json:"channel"`
	IsFinal     bool `json:"is_final"`
	SpeechFinal bool `json:"speech_final"`
}

// NewDeepgramClient creates a new Deepgram streaming STT client.
func NewDeepgramClient(ctx context.Context, cfg DeepgramConfig) (*DeepgramClient, error) {
	// Build WebSocket URL with query parameters
	url := fmt.Sprintf("%s?model=%s&language=%s&encoding=%s&sample_rate=%d&channels=%d&punctuate=%t",
		deepgramWSURL,
		cfg.Model,
		cfg.Language,
		cfg.Encoding,
		cfg.SampleRate,
		cfg.Channels,
		cfg.Punctuate,
	)

	if cfg.Endpointing > 0 {
		url += fmt.Sprintf("&endpointing=%d", cfg.Endpointing)
	}

	if cfg.UtteranceEndMs > 0 {
		url += fmt.Sprintf("&utterance_end_ms=%d", cfg.UtteranceEndMs)
	}

	// Set up headers with API key
	headers := http.Header{}
	headers.Set("Authorization", "Token "+cfg.APIKey)

	// Connect to Deepgram
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Deepgram: %w", err)
	}

	client := &DeepgramClient{
		conn:    conn,
		results: make(chan TranscriptResult, 100),
		errors:  make(chan error, 10),
		done:    make(chan struct{}),
	}

	// Start reading responses
	client.wg.Add(1)
	go client.readLoop()

	return client, nil
}

// StreamAudio sends audio data to Deepgram.
func (c *DeepgramClient) StreamAudio(ctx context.Context, audio []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.done:
		return fmt.Errorf("client is closed")
	default:
	}

	return c.conn.WriteMessage(websocket.BinaryMessage, audio)
}

// Results returns the channel for receiving transcription results.
func (c *DeepgramClient) Results() <-chan TranscriptResult {
	return c.results
}

// Errors returns the channel for receiving errors.
func (c *DeepgramClient) Errors() <-chan error {
	return c.errors
}

// Close closes the Deepgram connection.
func (c *DeepgramClient) Close() error {
	var err error
	c.closeOnce.Do(func() {
		close(c.done)

		// Send close message to Deepgram
		c.mu.Lock()
		closeMsg := []byte(`{"type": "CloseStream"}`)
		_ = c.conn.WriteMessage(websocket.TextMessage, closeMsg)
		c.mu.Unlock()

		err = c.conn.Close()

		// Wait for readLoop to finish before closing channels
		c.wg.Wait()
		close(c.results)
		close(c.errors)
	})
	return err
}

// readLoop reads responses from Deepgram and sends them to the results channel.
func (c *DeepgramClient) readLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.done:
			return
		default:
		}

		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			select {
			case <-c.done:
				return
			case c.errors <- fmt.Errorf("read error: %w", err):
			default:
			}
			return
		}

		var resp deepgramResponse
		if err := json.Unmarshal(msg, &resp); err != nil {
			log.Printf("deepgram: failed to parse response: %v", err)
			continue
		}

		// Skip non-results messages
		if resp.Type != "Results" {
			continue
		}

		// Extract transcript from first alternative (can be empty).
		var transcript string
		var confidence float64
		if len(resp.Channel.Alternatives) > 0 {
			alt := resp.Channel.Alternatives[0]
			transcript = alt.Transcript
			confidence = alt.Confidence
		}

		result := TranscriptResult{
			Text:         transcript,
			Confidence:   confidence,
			SegmentFinal: resp.IsFinal,
			SpeechFinal:  resp.SpeechFinal,
		}

		// Emit events even if transcript is empty when we have boundary signals.
		if result.Text == "" && !result.SegmentFinal && !result.SpeechFinal {
			continue
		}

		select {
		case <-c.done:
			return
		case c.results <- result:
		}
	}
}
