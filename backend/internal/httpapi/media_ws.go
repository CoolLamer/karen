package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lukasbauer/karen/internal/llm"
	"github.com/lukasbauer/karen/internal/stt"
	"github.com/lukasbauer/karen/internal/store"
	"github.com/lukasbauer/karen/internal/tts"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// fillerWords are short Czech acknowledgments spoken immediately after user finishes
// to reduce perceived latency while LLM generates response
var fillerWords = []string{
	"Jasně...",   // Sure...
	"Rozumím...", // I understand...
	"Hmm...",     // Hmm...
	"Aha...",     // Aha...
	"Dobře...",   // Okay...
}

// getRandomFiller returns a random filler word from the list
func getRandomFiller() string {
	return fillerWords[rand.Intn(len(fillerWords))]
}

// fillerSkipProbability is the chance (0.0-1.0) to skip speaking a filler word
// to make conversations feel more natural and less repetitive
const fillerSkipProbability = 0.3

// fillerCooldown is the minimum time between filler words to avoid repetition
const fillerCooldown = 10 * time.Second

// shouldSpeakFiller determines if we should speak a filler word based on:
// - Random skip probability (30% chance to skip for variety)
// - Cooldown period since last filler (avoid back-to-back fillers)
func shouldSpeakFiller(lastFillerTime time.Time) bool {
	// Skip randomly for variety
	if rand.Float64() < fillerSkipProbability {
		return false
	}
	// Skip if we spoke a filler recently
	if time.Since(lastFillerTime) < fillerCooldown {
		return false
	}
	return true
}

// isSentenceEnd checks if the text ends with a sentence-ending punctuation
func isSentenceEnd(text string) bool {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return false
	}
	lastChar := text[len(text)-1]
	return lastChar == '.' || lastChar == '!' || lastChar == '?'
}

// extractCompleteSentences extracts complete sentences from buffer,
// returns (complete sentences, remaining buffer)
func extractCompleteSentences(buffer string) (string, string) {
	// Find the last sentence boundary
	lastBoundary := -1
	for i := len(buffer) - 1; i >= 0; i-- {
		c := buffer[i]
		if c == '.' || c == '!' || c == '?' {
			lastBoundary = i
			break
		}
	}

	if lastBoundary == -1 {
		return "", buffer
	}

	return buffer[:lastBoundary+1], buffer[lastBoundary+1:]
}

// Twilio Media Stream message types
type twilioMessage struct {
	Event          string          `json:"event"`
	SequenceNumber string          `json:"sequenceNumber,omitempty"`
	Media          *twilioMedia    `json:"media,omitempty"`
	Start          *twilioStart    `json:"start,omitempty"`
	StreamSid      string          `json:"streamSid,omitempty"`
}

type twilioMedia struct {
	Track     string `json:"track"`
	Chunk     string `json:"chunk"`
	Timestamp string `json:"timestamp"`
	Payload   string `json:"payload"` // Base64 μ-law audio
}

type twilioStart struct {
	StreamSid    string            `json:"streamSid"`
	AccountSid   string            `json:"accountSid"`
	CallSid      string            `json:"callSid"`
	Tracks       []string          `json:"tracks"`
	CustomParams map[string]string `json:"customParameters"`
	MediaFormat  struct {
		Encoding   string `json:"encoding"`
		SampleRate int    `json:"sampleRate"`
		Channels   int    `json:"channels"`
	} `json:"mediaFormat"`
}

// twilioOutboundMedia is the format for sending audio back to Twilio
type twilioOutboundMedia struct {
	Event     string `json:"event"`
	StreamSid string `json:"streamSid"`
	Media     struct {
		Payload string `json:"payload"` // Base64 μ-law audio
	} `json:"media"`
}

// twilioMark sends a mark event to track when audio completes
type twilioMark struct {
	Event     string `json:"event"`
	StreamSid string `json:"streamSid"`
	Mark      struct {
		Name string `json:"name"`
	} `json:"mark"`
}

// twilioClear sends a clear event to stop audio playback (for barge-in)
type twilioClear struct {
	Event     string `json:"event"`
	StreamSid string `json:"streamSid"`
}

// TenantConfig holds tenant-specific settings for the call
type TenantConfig struct {
	TenantID       string   `json:"tenant_id,omitempty"`
	SystemPrompt   string   `json:"system_prompt,omitempty"`
	GreetingText   *string  `json:"greeting_text,omitempty"`
	VoiceID        *string  `json:"voice_id,omitempty"`
	Language       string   `json:"language,omitempty"`
	Endpointing    *int     `json:"endpointing,omitempty"` // STT endpointing in ms (default 800)
	VIPNames       []string `json:"vip_names,omitempty"`
	MarketingEmail *string  `json:"marketing_email,omitempty"`
	ForwardNumber  *string  `json:"forward_number,omitempty"`
}

// callSession manages a single call's voice AI session
type callSession struct {
	callSid    string
	streamSid  string
	accountSid string
	callID     string // DB call ID

	conn   *websocket.Conn
	connMu sync.Mutex

	sttClient *stt.DeepgramClient
	llmClient *llm.OpenAIClient
	ttsClient *tts.ElevenLabsClient

	store      *store.Store
	logger     *log.Logger
	cfg        RouterConfig
	httpClient *http.Client

	// Tenant-specific configuration
	tenantCfg TenantConfig

	// Conversation state
	messages       []llm.Message
	utteranceSeq   int
	speaking       bool      // True when TTS is playing
	speakingMu     sync.Mutex
	lastFillerTime time.Time // Last time a filler word was spoken

	// Barge-in handling
	bargeInCh chan string // Channel to signal barge-in with the interrupting text

	// Goodbye handling
	pendingGoodbye bool          // True when waiting for goodbye audio to finish
	goodbyeDone    chan struct{} // Signaled when goodbye mark is received

	ctx    context.Context
	cancel context.CancelFunc
}

func (r *Router) handleMediaWS(w http.ResponseWriter, req *http.Request) {
	// Check if we have required API keys
	if r.cfg.DeepgramAPIKey == "" || r.cfg.OpenAIAPIKey == "" || r.cfg.ElevenLabsAPIKey == "" {
		r.logger.Printf("media_ws: missing API keys")
		captureError(req, fmt.Errorf("voice AI not configured: missing API keys"), "media_ws: configuration error")
		http.Error(w, "voice AI not configured", http.StatusServiceUnavailable)
		return
	}

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		r.logger.Printf("media_ws: upgrade failed: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(req.Context())

	session := &callSession{
		conn:        conn,
		store:       r.store,
		logger:      r.logger,
		cfg:         r.cfg,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		messages:    []llm.Message{},
		bargeInCh:   make(chan string, 1), // Buffered channel for barge-in
		goodbyeDone: make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Create LLM client (doesn't require connection)
	session.llmClient = llm.NewOpenAIClient(llm.OpenAIConfig{
		APIKey: r.cfg.OpenAIAPIKey,
		Model:  "gpt-4o-mini",
	})

	// Create TTS client
	session.ttsClient = tts.NewElevenLabsClient(tts.ElevenLabsConfig{
		APIKey:     r.cfg.ElevenLabsAPIKey,
		VoiceID:    r.cfg.TTSVoiceID,
		ModelID:    "eleven_flash_v2_5",
		Stability:  r.cfg.TTSStability,
		Similarity: r.cfg.TTSSimilarity,
	})

	r.logger.Printf("media_ws: connection established, waiting for start message")

	// Handle the WebSocket connection
	session.run()
}

func (s *callSession) run() {
	defer s.cleanup()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		_, msg, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				s.logger.Printf("media_ws: connection closed for call %s", s.callSid)
			} else {
				s.logger.Printf("media_ws: read error for call %s: %v", s.callSid, err)
			}
			return
		}

		var twilioMsg twilioMessage
		if err := json.Unmarshal(msg, &twilioMsg); err != nil {
			s.logger.Printf("media_ws: failed to parse message: %v", err)
			continue
		}

		switch twilioMsg.Event {
		case "connected":
			s.logger.Printf("media_ws: Twilio connected for call %s", s.callSid)

		case "start":
			if err := s.handleStart(twilioMsg.Start); err != nil {
				s.logger.Printf("media_ws: start error: %v", err)
				return
			}

		case "media":
			if err := s.handleMedia(twilioMsg.Media); err != nil {
				s.logger.Printf("media_ws: media error: %v", err)
			}

		case "stop":
			s.logger.Printf("media_ws: stream stopped for call %s", s.callSid)
			return

		case "mark":
			// Audio playback completed
			s.speakingMu.Lock()
			s.speaking = false
			// Signal goodbye completion if we're waiting
			if s.pendingGoodbye {
				s.pendingGoodbye = false
				select {
				case s.goodbyeDone <- struct{}{}:
				default:
				}
			}
			s.speakingMu.Unlock()
		}
	}
}

func (s *callSession) handleStart(start *twilioStart) error {
	if start == nil {
		return fmt.Errorf("nil start message")
	}

	s.streamSid = start.StreamSid
	s.accountSid = start.AccountSid

	// Get callSid from custom parameters or directly from start message
	s.callSid = start.CallSid
	if s.callSid == "" {
		if cs, ok := start.CustomParams["callSid"]; ok {
			s.callSid = cs
		}
	}

	// Parse tenant config from custom parameters
	if tenantID, ok := start.CustomParams["tenantId"]; ok {
		s.tenantCfg.TenantID = tenantID
	}
	if configJSON, ok := start.CustomParams["tenantConfig"]; ok {
		if err := json.Unmarshal([]byte(configJSON), &s.tenantCfg); err != nil {
			s.logger.Printf("media_ws: failed to parse tenant config: %v", err)
		}
	}

	if s.tenantCfg.TenantID != "" {
		s.logger.Printf("media_ws: stream started - StreamSid: %s, CallSid: %s, TenantID: %s",
			start.StreamSid, s.callSid, s.tenantCfg.TenantID)
	} else {
		s.logger.Printf("media_ws: stream started - StreamSid: %s, CallSid: %s (no tenant)",
			start.StreamSid, s.callSid)
	}

	// Get call ID from database now that we have callSid
	if s.callSid != "" {
		callID, err := s.store.GetCallID(s.ctx, s.callSid)
		if err != nil {
			s.logger.Printf("media_ws: failed to get call ID for %s: %v", s.callSid, err)
		} else {
			s.callID = callID
		}
	}

	// Determine language for STT (from tenant config or default)
	language := "cs"
	if s.tenantCfg.Language != "" {
		language = s.tenantCfg.Language
	}

	// Determine endpointing (from tenant config or default 800ms)
	endpointing := 800
	if s.tenantCfg.Endpointing != nil && *s.tenantCfg.Endpointing > 0 {
		endpointing = *s.tenantCfg.Endpointing
		s.logger.Printf("media_ws: using tenant endpointing: %dms", endpointing)
	}

	// Connect to Deepgram STT
	sttClient, err := stt.NewDeepgramClient(s.ctx, stt.DeepgramConfig{
		APIKey:      s.cfg.DeepgramAPIKey,
		Language:    language,
		Model:       "nova-3",
		SampleRate:  8000,
		Encoding:    "mulaw",
		Channels:    1,
		Punctuate:   true,
		Endpointing: endpointing, // Configurable silence threshold for turn detection
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Deepgram: %w", err)
	}
	s.sttClient = sttClient

	// Update TTS client with tenant's voice ID if specified
	if s.tenantCfg.VoiceID != nil && *s.tenantCfg.VoiceID != "" {
		s.ttsClient = tts.NewElevenLabsClient(tts.ElevenLabsConfig{
			APIKey:     s.cfg.ElevenLabsAPIKey,
			VoiceID:    *s.tenantCfg.VoiceID,
			ModelID:    "eleven_flash_v2_5",
			Stability:  s.cfg.TTSStability,
			Similarity: s.cfg.TTSSimilarity,
		})
	}

	// Set tenant's custom system prompt if available
	if s.tenantCfg.SystemPrompt != "" {
		s.llmClient.SetSystemPrompt(s.tenantCfg.SystemPrompt)
		s.logger.Printf("media_ws: using tenant's custom system prompt")
	}

	// Start processing STT results
	go s.processSTTResults()

	// Speak the greeting using ElevenLabs (same voice as rest of conversation)
	go s.speakGreeting()

	return nil
}

func (s *callSession) handleMedia(media *twilioMedia) error {
	if media == nil || s.sttClient == nil {
		return nil
	}

	// Decode base64 audio
	audio, err := base64.StdEncoding.DecodeString(media.Payload)
	if err != nil {
		return fmt.Errorf("failed to decode audio: %w", err)
	}

	// Forward to STT
	return s.sttClient.StreamAudio(s.ctx, audio)
}

func (s *callSession) processSTTResults() {
	var currentUtterance strings.Builder
	var utteranceStartTime *time.Time

	for {
		select {
		case <-s.ctx.Done():
			return

		case err := <-s.sttClient.Errors():
			s.logger.Printf("media_ws: STT error: %v", err)
			return

		case result, ok := <-s.sttClient.Results():
			if !ok {
				return
			}

			if result.Text == "" {
				continue
			}

			// Track utterance timing
			if utteranceStartTime == nil {
				now := time.Now().UTC()
				utteranceStartTime = &now
			}

			if result.IsFinal {
				// Append final text
				if currentUtterance.Len() > 0 {
					currentUtterance.WriteString(" ")
				}
				currentUtterance.WriteString(result.Text)

				// Complete utterance - process with LLM
				text := strings.TrimSpace(currentUtterance.String())
				if text != "" {
					// Check if this is a barge-in (user speaking while agent is speaking)
					s.speakingMu.Lock()
					isSpeaking := s.speaking
					s.speakingMu.Unlock()

					if isSpeaking {
						// Barge-in detected: clear audio and signal interruption.
						// Note: There's a small timing window where speakText may send a few
						// more audio chunks after clearAudio() is called but before the
						// bargeInCh signal is received. This is acceptable because Twilio's
						// "clear" command will stop playback regardless of pending chunks.
						s.logger.Printf("media_ws: BARGE-IN detected - caller said: %s", text)
						if err := s.clearAudio(); err != nil {
							s.logger.Printf("media_ws: failed to clear audio: %v", err)
						}
						// Signal barge-in via channel (non-blocking)
						select {
						case s.bargeInCh <- text:
						default:
							// Channel full, skip (shouldn't happen with buffered channel)
						}
						// Mark speaking as false since we cleared
						s.speakingMu.Lock()
						s.speaking = false
						s.speakingMu.Unlock()
					}

					s.logger.Printf("media_ws: caller said: %s", text)

					// Store utterance (mark as interruption if barge-in)
					s.utteranceSeq++
					now := time.Now().UTC()
					confidence := result.Confidence
					if s.callID != "" {
						_ = s.store.InsertUtterance(s.ctx, s.callID, store.Utterance{
							Speaker:       "caller",
							Text:          text,
							Sequence:      s.utteranceSeq,
							StartedAt:     utteranceStartTime,
							EndedAt:       &now,
							STTConfidence: &confidence,
							Interrupted:   isSpeaking, // Mark as interruption if it was a barge-in
						})
					}

					// Add to conversation history
					s.messages = append(s.messages, llm.Message{
						Role:    "user",
						Content: text,
					})

					// Speak filler word immediately, then generate and speak response
					go s.speakFillerAndGenerate()
				}

				// Reset for next utterance
				currentUtterance.Reset()
				utteranceStartTime = nil
			}
		}
	}
}

// speakFillerAndGenerate speaks a short filler word immediately to reduce perceived
// latency, then generates and speaks the actual LLM response.
// Filler words are skipped randomly (30% of time) and when spoken recently (cooldown).
func (s *callSession) speakFillerAndGenerate() {
	// Check if already speaking - acquire lock for entire operation to prevent
	// timing window where another goroutine could slip through
	s.speakingMu.Lock()
	if s.speaking {
		s.speakingMu.Unlock()
		return
	}
	s.speaking = true
	lastFiller := s.lastFillerTime
	s.speakingMu.Unlock()

	// Optionally speak filler word (skip randomly or if recently spoken for variety)
	if shouldSpeakFiller(lastFiller) {
		filler := getRandomFiller()
		s.logger.Printf("media_ws: speaking filler: %s", filler)
		if err := s.speakText(filler); err != nil {
			s.logger.Printf("media_ws: filler TTS error: %v", err)
		}
		// Update last filler time
		s.speakingMu.Lock()
		s.lastFillerTime = time.Now()
		s.speakingMu.Unlock()
	} else {
		s.logger.Printf("media_ws: skipping filler (variety/cooldown)")
	}

	// Generate and speak the actual response (keep speaking=true throughout)
	s.generateAndSpeakInternal()
}

func (s *callSession) generateAndSpeak() {
	// Check if already speaking (barge-in handling would go here)
	s.speakingMu.Lock()
	if s.speaking {
		s.speakingMu.Unlock()
		return // Skip if already speaking, let current response finish
	}
	s.speaking = true
	s.speakingMu.Unlock()

	s.generateAndSpeakInternal()
}

// generateAndSpeakInternal generates and speaks the LLM response.
// Assumes speaking is already set to true by caller.
func (s *callSession) generateAndSpeakInternal() {
	defer func() {
		s.speakingMu.Lock()
		s.speaking = false
		s.speakingMu.Unlock()
	}()

	// Generate response with LLM
	responseCh, err := s.llmClient.GenerateResponse(s.ctx, s.messages)
	if err != nil {
		s.logger.Printf("media_ws: LLM error: %v", err)
		return
	}

	// Stream LLM output with sentence-based TTS for lower latency
	var fullResponse strings.Builder
	var buffer strings.Builder
	sentenceCount := 0

	for chunk := range responseCh {
		fullResponse.WriteString(chunk)
		buffer.WriteString(chunk)

		// Check if we have a complete sentence to speak
		completeSentences, remaining := extractCompleteSentences(buffer.String())
		if completeSentences != "" {
			sentenceCount++
			// Speak complete sentences immediately (strip forward marker)
			ttsText := stripForwardMarker(completeSentences)
			if ttsText != "" {
				if sentenceCount == 1 {
					s.logger.Printf("media_ws: streaming first sentence: %s", ttsText)
				}
				if err := s.speakText(ttsText); err != nil {
					s.logger.Printf("media_ws: TTS error: %v", err)
				}
			}
			// Keep only the remaining incomplete sentence
			buffer.Reset()
			buffer.WriteString(remaining)
		}
	}

	// Speak any remaining text that didn't end with punctuation
	remaining := strings.TrimSpace(buffer.String())
	if remaining != "" {
		ttsText := stripForwardMarker(remaining)
		if ttsText != "" {
			if err := s.speakText(ttsText); err != nil {
				s.logger.Printf("media_ws: TTS error: %v", err)
			}
		}
	}

	responseText := strings.TrimSpace(fullResponse.String())
	if responseText == "" {
		return
	}

	s.logger.Printf("media_ws: agent response (full): %s", responseText)

	// Add to conversation history
	s.messages = append(s.messages, llm.Message{
		Role:    "assistant",
		Content: responseText,
	})

	// Store agent utterance
	s.utteranceSeq++
	startTime := time.Now().UTC()
	if s.callID != "" {
		_ = s.store.InsertUtterance(s.ctx, s.callID, store.Utterance{
			Speaker:     "agent",
			Text:        responseText,
			Sequence:    s.utteranceSeq,
			StartedAt:   &startTime,
			Interrupted: false,
		})
	}

	// Check for forwarding (emergency/VIP)
	if isForward(responseText) {
		s.logger.Printf("media_ws: detected forward request, will forward after audio finishes")
		s.speakingMu.Lock()
		s.pendingGoodbye = true // reuse same mechanism
		s.speakingMu.Unlock()
		go s.forwardCall()
		return
	}

	// If this was a goodbye, mark pending and hang up after audio finishes
	if isGoodbye(responseText) {
		s.logger.Printf("media_ws: detected goodbye, will hang up after audio finishes")
		s.speakingMu.Lock()
		s.pendingGoodbye = true
		s.speakingMu.Unlock()
		go s.hangUpCall()
	}
}

func (s *callSession) speakText(text string) error {
	// Get audio from TTS
	audioCh, err := s.ttsClient.SynthesizeStream(s.ctx, text)
	if err != nil {
		return err
	}

	// Send audio chunks to Twilio
	for chunk := range audioCh {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case <-s.bargeInCh:
			// Barge-in detected - stop sending audio
			s.logger.Printf("media_ws: stopping audio send due to barge-in")
			// Drain remaining audio
			for range audioCh {
			}
			return nil
		default:
		}

		// Encode as base64 and send to Twilio
		outMsg := twilioOutboundMedia{
			Event:     "media",
			StreamSid: s.streamSid,
		}
		outMsg.Media.Payload = base64.StdEncoding.EncodeToString(chunk)

		s.connMu.Lock()
		err := s.conn.WriteJSON(outMsg)
		s.connMu.Unlock()

		if err != nil {
			return fmt.Errorf("failed to send audio: %w", err)
		}
	}

	// Send mark to track completion
	mark := twilioMark{
		Event:     "mark",
		StreamSid: s.streamSid,
	}
	mark.Mark.Name = fmt.Sprintf("response-%d", s.utteranceSeq)

	s.connMu.Lock()
	err = s.conn.WriteJSON(mark)
	s.connMu.Unlock()

	return err
}

// clearAudio sends a clear event to Twilio to stop audio playback (for barge-in)
func (s *callSession) clearAudio() error {
	clearMsg := twilioClear{
		Event:     "clear",
		StreamSid: s.streamSid,
	}

	s.connMu.Lock()
	err := s.conn.WriteJSON(clearMsg)
	s.connMu.Unlock()

	if err != nil {
		return fmt.Errorf("failed to send clear: %w", err)
	}

	s.logger.Printf("media_ws: sent clear command (barge-in)")
	return nil
}

// speakGreeting sends the initial greeting via ElevenLabs TTS
func (s *callSession) speakGreeting() {
	s.speakingMu.Lock()
	s.speaking = true
	s.speakingMu.Unlock()

	defer func() {
		s.speakingMu.Lock()
		s.speaking = false
		s.speakingMu.Unlock()
	}()

	// Use tenant's greeting if available, otherwise fall back to config default
	var greeting string
	if s.tenantCfg.GreetingText != nil && *s.tenantCfg.GreetingText != "" {
		greeting = *s.tenantCfg.GreetingText
	} else if s.cfg.GreetingText != "" {
		greeting = s.cfg.GreetingText
	} else {
		greeting = "Dobrý den, tady Asistentka Karen. Lukáš nemá čas, ale můžu vám pro něj zanechat vzkaz - co od něj potřebujete?"
	}

	s.logger.Printf("media_ws: speaking greeting: %s", greeting)

	// Add greeting to conversation history so LLM knows it was already said
	s.messages = append(s.messages, llm.Message{
		Role:    "assistant",
		Content: greeting,
	})

	if err := s.speakText(greeting); err != nil {
		s.logger.Printf("media_ws: greeting TTS error: %v", err)
	}
}

// isGoodbye checks if the response contains goodbye phrases
func isGoodbye(text string) bool {
	lower := strings.ToLower(text)
	goodbyePhrases := []string{
		"na shledanou",
		"nashledanou",
		"mějte se",
		"hezký den",
	}
	for _, phrase := range goodbyePhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

// isForward checks if the response contains the forward marker
func isForward(text string) bool {
	return strings.Contains(text, "[PŘEPOJIT]")
}

// stripForwardMarker removes the [PŘEPOJIT] marker from text for TTS
func stripForwardMarker(text string) string {
	return strings.ReplaceAll(text, "[PŘEPOJIT] ", "")
}

// forwardCall forwards the call to the tenant's forward number
func (s *callSession) forwardCall() {
	if s.callSid == "" || s.accountSid == "" || s.cfg.TwilioAuthToken == "" {
		s.logger.Printf("media_ws: cannot forward - missing callSid, accountSid, or auth token")
		return
	}

	// Determine the forward number (tenant config or default)
	forwardNumber := "+420724794686" // Default fallback
	if s.tenantCfg.ForwardNumber != nil && *s.tenantCfg.ForwardNumber != "" {
		forwardNumber = *s.tenantCfg.ForwardNumber
	}

	// Wait for the forwarding message to finish playing
	select {
	case <-s.goodbyeDone:
		s.logger.Printf("media_ws: forward audio finished, forwarding call to %s", forwardNumber)
	case <-time.After(10 * time.Second):
		s.logger.Printf("media_ws: timeout waiting for forward audio, forwarding anyway to %s", forwardNumber)
	case <-s.ctx.Done():
		return
	}

	// Small delay to ensure audio is flushed
	time.Sleep(500 * time.Millisecond)

	// Forward to the configured number using TwiML
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Calls/%s.json",
		s.accountSid, s.callSid)

	// TwiML to dial the forward number
	twiml := fmt.Sprintf(`<Response><Dial>%s</Dial></Response>`, forwardNumber)

	data := url.Values{}
	data.Set("Twiml", twiml)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		s.logger.Printf("media_ws: failed to create forward request: %v", err)
		return
	}

	req.SetBasicAuth(s.accountSid, s.cfg.TwilioAuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Printf("media_ws: failed to forward call: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.logger.Printf("media_ws: call %s forwarded successfully to %s", s.callSid, forwardNumber)
	} else {
		s.logger.Printf("media_ws: forward returned status %d", resp.StatusCode)
	}
}

// hangUpCall terminates the call via Twilio REST API
func (s *callSession) hangUpCall() {
	if s.callSid == "" || s.accountSid == "" || s.cfg.TwilioAuthToken == "" {
		s.logger.Printf("media_ws: cannot hang up - missing callSid, accountSid, or auth token")
		return
	}

	// Wait for the goodbye audio to finish playing (signaled by mark event)
	// Use timeout as fallback in case mark never arrives
	select {
	case <-s.goodbyeDone:
		s.logger.Printf("media_ws: goodbye audio finished, hanging up")
	case <-time.After(10 * time.Second):
		s.logger.Printf("media_ws: timeout waiting for goodbye audio, hanging up anyway")
	case <-s.ctx.Done():
		return
	}

	// Small additional delay to ensure Twilio has flushed all audio
	time.Sleep(500 * time.Millisecond)

	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Calls/%s.json",
		s.accountSid, s.callSid)

	data := url.Values{}
	data.Set("Status", "completed")

	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		s.logger.Printf("media_ws: failed to create hang up request: %v", err)
		return
	}

	req.SetBasicAuth(s.accountSid, s.cfg.TwilioAuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Printf("media_ws: failed to hang up call: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.logger.Printf("media_ws: call %s hung up successfully", s.callSid)
		// Update database status to completed
		if s.callID != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.store.UpdateCallStatus(ctx, s.callSid, "completed", time.Now().UTC()); err != nil {
				s.logger.Printf("media_ws: failed to update call status: %v", err)
			}
		}
	} else {
		s.logger.Printf("media_ws: hang up returned status %d", resp.StatusCode)
	}
}

func (s *callSession) analyzeCall() {
	if s.callID == "" {
		return
	}

	// Use background context since call context may be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.llmClient.AnalyzeCall(ctx, s.messages)
	if err != nil {
		s.logger.Printf("media_ws: analysis error: %v", err)
		return
	}

	// Convert entities to JSON
	entitiesJSON, _ := json.Marshal(result.Entities)

	// Store screening result
	sr := store.ScreeningResult{
		LegitimacyLabel:      result.LegitimacyLabel,
		LegitimacyConfidence: result.LegitimacyConfidence,
		IntentCategory:       result.IntentCategory,
		IntentText:           result.IntentText,
		EntitiesJSON:         entitiesJSON,
		CreatedAt:            time.Now().UTC(),
	}

	if err := s.store.InsertScreeningResult(ctx, s.callID, sr); err != nil {
		s.logger.Printf("media_ws: failed to store screening result: %v", err)
	} else {
		s.logger.Printf("media_ws: call classified as %s (%.0f%% confidence)",
			result.LegitimacyLabel, result.LegitimacyConfidence*100)
	}
}

func (s *callSession) cleanup() {
	s.cancel()

	if s.sttClient != nil {
		s.sttClient.Close()
	}

	s.connMu.Lock()
	s.conn.Close()
	s.connMu.Unlock()

	// Analyze the call at the end (only if we had a conversation)
	if len(s.messages) >= 2 {
		s.analyzeCall()
	}

	// Mark call as completed (fallback in case hangUpCall didn't run or failed)
	if s.callID != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.store.UpdateCallStatus(ctx, s.callSid, "completed", time.Now().UTC())
	}

	s.logger.Printf("media_ws: session cleaned up for call %s", s.callSid)
}
