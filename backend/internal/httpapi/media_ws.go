package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gorilla/websocket"
	"github.com/lukasbauer/karen/internal/eventlog"
	"github.com/lukasbauer/karen/internal/llm"
	"github.com/lukasbauer/karen/internal/notifications"
	"github.com/lukasbauer/karen/internal/store"
	"github.com/lukasbauer/karen/internal/stt"
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
	Mark           *twilioMarkData `json:"mark,omitempty"`
	StreamSid      string          `json:"streamSid,omitempty"`
}

type twilioMarkData struct {
	Name string `json:"name"`
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
	TenantID         string   `json:"tenant_id,omitempty"`
	SystemPrompt     string   `json:"system_prompt,omitempty"`
	GreetingText     *string  `json:"greeting_text,omitempty"`
	VoiceID          *string  `json:"voice_id,omitempty"`
	Language         string   `json:"language,omitempty"`
	Endpointing      *int     `json:"endpointing,omitempty"`         // STT endpointing in ms (default 800)
	UtteranceEnd     *int     `json:"utterance_end,omitempty"`       // Hard timeout after last speech in ms (default 1500)
	MaxTurnTimeoutMs *int     `json:"max_turn_timeout_ms,omitempty"` // Hard cap on waiting for speech_final in ms (default 4000)
	VIPNames         []string `json:"vip_names,omitempty"`
	MarketingEmail   *string  `json:"marketing_email,omitempty"`
	ForwardNumber    *string  `json:"forward_number,omitempty"`
	OwnerPhone       string   `json:"owner_phone,omitempty"` // User's verified phone for forwarding
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
	eventLog   *eventlog.Logger
	cfg        RouterConfig
	httpClient *http.Client
	apns       *notifications.APNsClient

	// Tenant-specific configuration
	tenantCfg TenantConfig

	// Conversation state
	messages   []llm.Message
	messagesMu sync.Mutex

	utteranceSeq   int
	lastFillerTime time.Time // Last time a filler word was spoken
	turnSeq        uint64    // Monotonic user turn id (for logging/debugging)

	// Response control (cancel ongoing response on barge-in / new utterance)
	respMu       sync.Mutex
	respSeq      uint64
	activeRespID uint64
	respCancel   context.CancelFunc

	// Barge-in handling
	bargeInCh chan string // Channel to signal barge-in with the interrupting text

	// Audio playback tracking via Twilio mark events
	audioMu      sync.Mutex
	audioPending int    // number of marks sent but not yet acknowledged by Twilio
	audioSeq     uint64 // monotonically increasing mark id

	// Goodbye/forward handling: wait for a specific mark before acting
	pendingDoneMarkID uint64 // 0 when not waiting
	actionMu          sync.Mutex
	actionCancel      context.CancelFunc

	// Goodbye handling
	goodbyeDone chan struct{} // Signaled when goodbye mark is received
	agentHungUp bool          // True if agent initiated the hangup (prevents overwrite by caller)

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
		sentry.CaptureException(err)
		return
	}

	ctx, cancel := context.WithCancel(req.Context())

	session := &callSession{
		conn:        conn,
		store:       r.store,
		logger:      r.logger,
		eventLog:    r.eventLog,
		cfg:         r.cfg,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		apns:        r.apns,
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
				sentry.CaptureException(err)
				return
			}

		case "media":
			if err := s.handleMedia(twilioMsg.Media); err != nil {
				s.logger.Printf("media_ws: media error: %v", err)
			}

		case "stop":
			s.logger.Printf("media_ws: stream stopped for call %s", s.callSid)

			// Mark call as ended by caller (only if agent didn't initiate the hangup)
			if s.callID != "" && s.callSid != "" && !s.agentHungUp {
				s.logger.Printf("media_ws: caller hung up")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := s.store.UpdateCallEndedBy(ctx, s.callSid, "caller"); err != nil {
					s.logger.Printf("media_ws: failed to update ended_by: %v", err)
					sentry.CaptureException(err)
				}
			}

			return

		case "mark":
			s.handleMark(twilioMsg.Mark)
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
			sentry.CaptureException(err)
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
			sentry.CaptureException(err)
		} else {
			s.callID = callID
		}
	}

	// Determine language for STT (from tenant config or default)
	language := "cs"
	if s.tenantCfg.Language != "" {
		language = s.tenantCfg.Language
	}

	// Determine endpointing (from tenant config or configured default)
	endpointing := s.cfg.STTEndpointingMs
	if endpointing <= 0 {
		endpointing = 1200
	}
	if s.tenantCfg.Endpointing != nil && *s.tenantCfg.Endpointing > 0 {
		endpointing = *s.tenantCfg.Endpointing
		s.logger.Printf("media_ws: using tenant endpointing: %dms", endpointing)
	}

	// Determine utterance_end_ms (hard timeout after last speech, regardless of noise)
	utteranceEnd := s.cfg.STTUtteranceEndMs
	if utteranceEnd <= 0 {
		utteranceEnd = 1500
	}
	if s.tenantCfg.UtteranceEnd != nil && *s.tenantCfg.UtteranceEnd > 0 {
		utteranceEnd = *s.tenantCfg.UtteranceEnd
		s.logger.Printf("media_ws: using tenant utterance_end: %dms", utteranceEnd)
	}

	// Connect to Deepgram STT
	sttClient, err := stt.NewDeepgramClient(s.ctx, stt.DeepgramConfig{
		APIKey:         s.cfg.DeepgramAPIKey,
		Language:       language,
		Model:          "nova-3",
		SampleRate:     8000,
		Encoding:       "mulaw",
		Channels:       1,
		Punctuate:      true,
		Endpointing:    endpointing,  // Silence-based turn detection
		UtteranceEndMs: utteranceEnd, // Hard timeout after last speech (noise-resistant)
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

	// Log call started event
	s.eventLog.LogAsync(s.callID, eventlog.EventCallStarted, map[string]any{
		"stream_sid":       s.streamSid,
		"call_sid":         s.callSid,
		"tenant_id":        s.tenantCfg.TenantID,
		"endpointing_ms":   endpointing,
		"utterance_end_ms": utteranceEnd,
		"language":         language,
	})

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

func (s *callSession) nextAudioMarkID() uint64 {
	return atomic.AddUint64(&s.audioSeq, 1)
}

func (s *callSession) incAudioPending() int {
	s.audioMu.Lock()
	s.audioPending++
	n := s.audioPending
	s.audioMu.Unlock()
	return n
}

func (s *callSession) decAudioPending() int {
	s.audioMu.Lock()
	if s.audioPending > 0 {
		s.audioPending--
	}
	n := s.audioPending
	s.audioMu.Unlock()
	return n
}

func (s *callSession) isAudioPlaying() bool {
	s.audioMu.Lock()
	n := s.audioPending
	s.audioMu.Unlock()
	return n > 0
}

func parseAudioMarkID(name string) (uint64, bool) {
	if !strings.HasPrefix(name, "audio-") {
		return 0, false
	}
	idStr := strings.TrimPrefix(name, "audio-")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (s *callSession) handleMark(mark *twilioMarkData) {
	if mark == nil {
		return
	}
	markID, ok := parseAudioMarkID(mark.Name)
	pending := s.decAudioPending()

	if ok && markID != 0 {
		s.logger.Printf("media_ws: mark received %q (pending=%d)", mark.Name, pending)
	} else {
		s.logger.Printf("media_ws: mark received (unparsed) %q (pending=%d)", mark.Name, pending)
	}

	// Signal goodbye/forward completion only for the final mark we’re waiting on.
	if ok {
		s.audioMu.Lock()
		awaitID := s.pendingDoneMarkID
		if awaitID != 0 && markID == awaitID {
			s.pendingDoneMarkID = 0
			select {
			case s.goodbyeDone <- struct{}{}:
			default:
			}
		}
		s.audioMu.Unlock()
	}
}

func (s *callSession) beginNewResponse() (context.Context, uint64) {
	s.respMu.Lock()
	// Cancel any in-flight response generation/audio sending.
	if s.respCancel != nil {
		s.respCancel()
	}
	respID := atomic.AddUint64(&s.respSeq, 1)
	ctx, cancel := context.WithCancel(s.ctx)
	s.activeRespID = respID
	s.respCancel = cancel
	s.respMu.Unlock()

	// Drain any stale barge-in signals from previous responses.
	// Without this, a barge-in from a previous turn could immediately
	// interrupt the new response's TTS.
	select {
	case <-s.bargeInCh:
	default:
	}

	return ctx, respID
}

func (s *callSession) endResponse(respID uint64) {
	s.respMu.Lock()
	if s.activeRespID == respID {
		s.activeRespID = 0
		s.respCancel = nil
	}
	s.respMu.Unlock()
}

func (s *callSession) cancelResponse() {
	s.respMu.Lock()
	if s.respCancel != nil {
		s.respCancel()
	}
	s.respMu.Unlock()
}

func (s *callSession) isResponseActive() bool {
	s.respMu.Lock()
	active := s.respCancel != nil
	s.respMu.Unlock()
	return active
}

func (s *callSession) beginPendingAction() context.Context {
	s.actionMu.Lock()
	if s.actionCancel != nil {
		s.actionCancel()
	}
	ctx, cancel := context.WithCancel(s.ctx)
	s.actionCancel = cancel
	s.actionMu.Unlock()
	return ctx
}

func (s *callSession) cancelPendingAction() {
	s.actionMu.Lock()
	if s.actionCancel != nil {
		s.actionCancel()
		s.actionCancel = nil
	}
	s.actionMu.Unlock()
}

func (s *callSession) processSTTResults() {
	var currentUtterance strings.Builder
	var utteranceStartTime *time.Time
	var lastConfidence float64

	// We only finalize a user "turn" on Deepgram's end-of-speech signal (`speech_final`),
	// with a short grace window to allow late-arriving tokens/segments.
	const speechFinalGrace = 250 * time.Millisecond
	// Hard cap on waiting for speech_final. If Deepgram is confused by noise, we force
	// finalization after this timeout to avoid multi-second delays.
	// Default is 4 seconds, but can be overridden per-tenant.
	maxTurnTimeout := 4 * time.Second
	if s.tenantCfg.MaxTurnTimeoutMs != nil && *s.tenantCfg.MaxTurnTimeoutMs > 0 {
		maxTurnTimeout = time.Duration(*s.tenantCfg.MaxTurnTimeoutMs) * time.Millisecond
		s.logger.Printf("media_ws: using tenant max_turn_timeout: %v", maxTurnTimeout)
	}

	var finalizeTimer *time.Timer
	var finalizeC <-chan time.Time
	stopFinalizeTimer := func() {
		if finalizeTimer == nil {
			return
		}
		if !finalizeTimer.Stop() {
			// Drain if already fired to avoid spurious wakeups after Reset.
			select {
			case <-finalizeTimer.C:
			default:
			}
		}
	}
	scheduleFinalize := func() {
		if finalizeTimer == nil {
			finalizeTimer = time.NewTimer(speechFinalGrace)
		} else {
			stopFinalizeTimer()
			finalizeTimer.Reset(speechFinalGrace)
		}
		finalizeC = finalizeTimer.C
	}
	cancelFinalize := func() {
		stopFinalizeTimer()
		finalizeC = nil
	}

	// Max turn timeout: force finalization if speech_final doesn't arrive in time.
	var maxTurnTimer *time.Timer
	var maxTurnC <-chan time.Time
	stopMaxTurnTimer := func() {
		if maxTurnTimer == nil {
			return
		}
		if !maxTurnTimer.Stop() {
			select {
			case <-maxTurnTimer.C:
			default:
			}
		}
	}
	scheduleMaxTurn := func() {
		if maxTurnTimer == nil {
			maxTurnTimer = time.NewTimer(maxTurnTimeout)
		} else {
			stopMaxTurnTimer()
			maxTurnTimer.Reset(maxTurnTimeout)
		}
		maxTurnC = maxTurnTimer.C
	}
	cancelMaxTurn := func() {
		stopMaxTurnTimer()
		maxTurnC = nil
	}

	// Track whether we already issued a barge-in clear for the current speaking overlap.
	bargeInSent := false

	finalizeUtterance := func() {
		text := strings.TrimSpace(currentUtterance.String())
		if text == "" {
			// Nothing meaningful captured; reset state and move on.
			currentUtterance.Reset()
			utteranceStartTime = nil
			lastConfidence = 0
			bargeInSent = false
			return
		}

		// Barge-in is defined as the caller speaking while the agent is speaking.
		isSpeaking := bargeInSent || s.isAudioPlaying() || s.isResponseActive()

		if isSpeaking && !bargeInSent {
			// Ensure playback is stopped (safety net). We may have already cleared on interim results.
			s.logger.Printf("media_ws: BARGE-IN detected - caller said: %s", text)
			if err := s.clearAudio(); err != nil {
				s.logger.Printf("media_ws: failed to clear audio: %v", err)
				sentry.CaptureException(err)
			}
			s.cancelResponse()
			s.cancelPendingAction()
			select {
			case s.bargeInCh <- text:
			default:
			}
		}

		turnID := atomic.AddUint64(&s.turnSeq, 1)
		s.logger.Printf("media_ws: caller said (turn=%d): %s", turnID, text)
		s.eventLog.LogAsync(s.callID, eventlog.EventTurnFinalized, map[string]any{
			"turn_id":     turnID,
			"text":        text,
			"confidence":  lastConfidence,
			"interrupted": isSpeaking,
		})

		// Store utterance (mark as interruption if barge-in)
		s.utteranceSeq++
		now := time.Now().UTC()
		confidence := lastConfidence
		if s.callID != "" {
			_ = s.store.InsertUtterance(s.ctx, s.callID, store.Utterance{
				Speaker:       "caller",
				Text:          text,
				Sequence:      s.utteranceSeq,
				StartedAt:     utteranceStartTime,
				EndedAt:       &now,
				STTConfidence: &confidence,
				Interrupted:   isSpeaking,
			})
		}

		// Add to conversation history
		s.messagesMu.Lock()
		s.messages = append(s.messages, llm.Message{
			Role:    "user",
			Content: text,
		})
		s.messagesMu.Unlock()

		// Speak filler word immediately, then generate and speak response
		go s.speakFillerAndGenerate(turnID, text)

		// Reset for next utterance
		currentUtterance.Reset()
		utteranceStartTime = nil
		lastConfidence = 0
		bargeInSent = false
	}

	for {
		select {
		case <-s.ctx.Done():
			cancelFinalize()
			cancelMaxTurn()
			return

		case err := <-s.sttClient.Errors():
			s.logger.Printf("media_ws: STT error: %v", err)
			sentry.CaptureException(err)
			cancelFinalize()
			cancelMaxTurn()
			return

		case result, ok := <-s.sttClient.Results():
			if !ok {
				cancelFinalize()
				cancelMaxTurn()
				return
			}

			// Handle VAD events from Deepgram (for debugging noise issues).
			if result.VADSpeechStarted {
				s.eventLog.LogAsync(s.callID, eventlog.EventVADSpeechStarted, nil)
				continue
			}
			if result.VADUtteranceEnd {
				s.eventLog.LogAsync(s.callID, eventlog.EventVADUtteranceEnd, nil)
				continue
			}

			// Optional instrumentation (keep it light; text is logged at finalize).
			if result.SegmentFinal || result.SpeechFinal {
				s.logger.Printf("media_ws: stt event (segment_final=%t speech_final=%t) text=%q",
					result.SegmentFinal, result.SpeechFinal, strings.TrimSpace(result.Text))
				s.eventLog.LogAsync(s.callID, eventlog.EventSTTResult, map[string]any{
					"text":          strings.TrimSpace(result.Text),
					"confidence":    result.Confidence,
					"segment_final": result.SegmentFinal,
					"speech_final":  result.SpeechFinal,
				})
			}

			// Detect barge-in as early as possible (on any non-empty transcript),
			// not only after endpointing emits a final utterance.
			if strings.TrimSpace(result.Text) != "" {
				isSpeaking := s.isAudioPlaying() || s.isResponseActive()
				if isSpeaking && !bargeInSent {
					s.logger.Printf("media_ws: early BARGE-IN (partial) - caller: %s", strings.TrimSpace(result.Text))
					s.eventLog.LogAsync(s.callID, eventlog.EventBargeIn, map[string]any{
						"partial_text":       strings.TrimSpace(result.Text),
						"agent_was_speaking": true,
					})
					if err := s.clearAudio(); err != nil {
						s.logger.Printf("media_ws: failed to clear audio: %v", err)
						sentry.CaptureException(err)
					}
					s.cancelResponse()
					s.cancelPendingAction()
					select {
					case s.bargeInCh <- strings.TrimSpace(result.Text):
					default:
					}
					bargeInSent = true
				}
			}

			// Track utterance timing once we see any text for this turn.
			if utteranceStartTime == nil && strings.TrimSpace(result.Text) != "" {
				now := time.Now().UTC()
				utteranceStartTime = &now
			}

			// Append only segment-final text. Interim transcripts can be unstable/revised.
			if result.SegmentFinal {
				if strings.TrimSpace(result.Text) != "" {
					if currentUtterance.Len() > 0 {
						currentUtterance.WriteString(" ")
					}
					currentUtterance.WriteString(strings.TrimSpace(result.Text))
					lastConfidence = result.Confidence
					// Start/reset max turn timer - we have text, so ensure we finalize eventually.
					scheduleMaxTurn()
				}
				// If we already saw end-of-speech, keep grace window open for this late segment.
				if finalizeC != nil {
					scheduleFinalize()
				}
			}

			// Schedule finalization on end-of-speech (NOT on segment final).
			if result.SpeechFinal {
				cancelMaxTurn() // No longer need hard timeout; speech_final arrived.
				scheduleFinalize()
			}

		case <-finalizeC:
			// End-of-speech grace window elapsed; finalize a single user turn.
			cancelFinalize()
			cancelMaxTurn()
			finalizeUtterance()

		case <-maxTurnC:
			// Hard timeout: speech_final didn't arrive in time (noisy environment).
			s.logger.Printf("media_ws: MAX TURN TIMEOUT - forcing finalization after 4s")
			s.eventLog.LogAsync(s.callID, eventlog.EventMaxTurnTimeout, map[string]any{
				"pending_text": strings.TrimSpace(currentUtterance.String()),
			})
			cancelMaxTurn()
			cancelFinalize()
			finalizeUtterance()
		}
	}
}

func (s *callSession) isCurrentResponse(respID uint64) bool {
	s.respMu.Lock()
	cur := s.activeRespID
	s.respMu.Unlock()
	return cur == respID
}

// speakFillerAndGenerate starts a new response (cancelling any previous one),
// optionally speaks a short filler after a brief delay, then streams the LLM
// response via sentence-based TTS.
func (s *callSession) speakFillerAndGenerate(turnID uint64, lastUserText string) {
	ctx, respID := s.beginNewResponse()
	defer s.endResponse(respID)
	s.logger.Printf("media_ws: starting response (turn=%d resp=%d)", turnID, respID)

	// Snapshot messages for this response (avoid races with concurrent appends).
	s.messagesMu.Lock()
	msgs := append([]llm.Message(nil), s.messages...)
	lastFiller := s.lastFillerTime
	s.messagesMu.Unlock()

	s.eventLog.LogAsync(s.callID, eventlog.EventLLMStarted, map[string]any{
		"turn_id":       turnID,
		"message_count": len(msgs),
	})

	llmStartTime := time.Now()
	responseCh, err := s.llmClient.GenerateResponse(ctx, msgs)
	if err != nil {
		// Context canceled is expected during barge-in, not a real error
		if !errors.Is(err, context.Canceled) {
			s.logger.Printf("media_ws: LLM error: %v", err)
			sentry.CaptureException(err)
			s.eventLog.LogAsync(s.callID, eventlog.EventLLMError, map[string]any{
				"turn_id": turnID,
				"error":   err.Error(),
			})
		}
		return
	}

	// Buffer LLM chunks so we can optionally speak filler without blocking the stream.
	llmBuf := make(chan string, 200)
	go func() {
		defer close(llmBuf)
		for chunk := range responseCh {
			select {
			case <-ctx.Done():
				return
			case llmBuf <- chunk:
			}
		}
	}()

	var fullResponse strings.Builder
	var buffer strings.Builder
	sentenceCount := 0
	var lastResponseMarkID uint64

	// If the model is slow to produce any output, speak a short acknowledgement.
	// This prevents filler when the model is already fast.
	fillerDelay := 350 * time.Millisecond
	timer := time.NewTimer(fillerDelay)
	defer timer.Stop()

	gotAnyChunk := false
	select {
	case <-ctx.Done():
		return
	case chunk, ok := <-llmBuf:
		if ok {
			gotAnyChunk = true
			fullResponse.WriteString(chunk)
			buffer.WriteString(chunk)
		}
	case <-timer.C:
		// No LLM output yet: consider filler (also skip for very short utterances).
		shortUtterance := len(strings.TrimSpace(lastUserText)) < 8
		if !shortUtterance && shouldSpeakFiller(lastFiller) {
			filler := getRandomFiller()
			s.logger.Printf("media_ws: speaking filler: %s", filler)
			s.eventLog.LogAsync(s.callID, eventlog.EventFillerSpoken, map[string]any{
				"turn_id": turnID,
				"filler":  filler,
			})
			if _, err := s.speakText(ctx, filler); err != nil && !errors.Is(err, context.Canceled) {
				s.logger.Printf("media_ws: filler TTS error: %v", err)
				sentry.CaptureException(err)
			}
			s.messagesMu.Lock()
			s.lastFillerTime = time.Now()
			s.messagesMu.Unlock()
		} else {
			s.logger.Printf("media_ws: skipping filler (variety/cooldown/short-utterance)")
			s.eventLog.LogAsync(s.callID, eventlog.EventFillerSkipped, map[string]any{
				"turn_id": turnID,
				"reason":  "variety/cooldown/short-utterance",
			})
		}
	}

	// Stream remaining LLM output with sentence-based TTS for lower latency.
	for chunk := range llmBuf {
		select {
		case <-ctx.Done():
			return
		default:
		}
		// If a newer response started, stop speaking stale content immediately.
		if !s.isCurrentResponse(respID) {
			return
		}

		gotAnyChunk = true
		fullResponse.WriteString(chunk)
		buffer.WriteString(chunk)

		completeSentences, remaining := extractCompleteSentences(buffer.String())
		if completeSentences != "" {
			sentenceCount++
			ttsText := stripForwardMarker(completeSentences)
			if ttsText != "" {
				if sentenceCount == 1 {
					s.logger.Printf("media_ws: streaming first sentence: %s", ttsText)
				}
				markID, err := s.speakText(ctx, ttsText)
				if err != nil && !errors.Is(err, context.Canceled) {
					s.logger.Printf("media_ws: TTS error: %v", err)
					sentry.CaptureException(err)
				} else if markID != 0 {
					lastResponseMarkID = markID
				}
			}
			buffer.Reset()
			buffer.WriteString(remaining)
		}
	}

	// Speak any remaining text that didn't end with punctuation.
	remaining := strings.TrimSpace(buffer.String())
	if remaining != "" && s.isCurrentResponse(respID) {
		ttsText := stripForwardMarker(remaining)
		if ttsText != "" {
			markID, err := s.speakText(ctx, ttsText)
			if err != nil && !errors.Is(err, context.Canceled) {
				s.logger.Printf("media_ws: TTS error: %v", err)
				sentry.CaptureException(err)
			} else if markID != 0 {
				lastResponseMarkID = markID
			}
		}
	}

	if !gotAnyChunk {
		return
	}

	responseText := strings.TrimSpace(fullResponse.String())
	if responseText == "" {
		return
	}

	s.logger.Printf("media_ws: agent response (full): %s", responseText)
	s.eventLog.LogAsync(s.callID, eventlog.EventLLMCompleted, map[string]any{
		"turn_id":         turnID,
		"response_length": len(responseText),
		"duration_ms":     time.Since(llmStartTime).Milliseconds(),
	})

	// Add to conversation history
	s.messagesMu.Lock()
	s.messages = append(s.messages, llm.Message{
		Role:    "assistant",
		Content: responseText,
	})
	s.messagesMu.Unlock()

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

	// If we need to forward/hang up, wait for the final mark of the response.
	if isForward(responseText) {
		s.logger.Printf("media_ws: detected forward request, will forward after audio finishes")
		s.eventLog.LogAsync(s.callID, eventlog.EventForwardDetected, map[string]any{
			"response_text": responseText,
		})
		if lastResponseMarkID != 0 {
			s.audioMu.Lock()
			s.pendingDoneMarkID = lastResponseMarkID
			s.audioMu.Unlock()
		}
		actionCtx := s.beginPendingAction()
		go s.forwardCall(actionCtx)
		return
	}
	if isGoodbye(responseText) {
		s.logger.Printf("media_ws: detected goodbye, will hang up after audio finishes")
		s.eventLog.LogAsync(s.callID, eventlog.EventGoodbyeDetected, map[string]any{
			"response_text": responseText,
		})
		if lastResponseMarkID != 0 {
			s.audioMu.Lock()
			s.pendingDoneMarkID = lastResponseMarkID
			s.audioMu.Unlock()
		}
		actionCtx := s.beginPendingAction()
		go s.hangUpCall(actionCtx)
	}
}

func (s *callSession) speakText(ctx context.Context, text string) (uint64, error) {
	ttsStartTime := time.Now()
	s.eventLog.LogAsync(s.callID, eventlog.EventTTSStarted, map[string]any{
		"text_length": len(text),
	})

	// Get audio from TTS
	audioCh, err := s.ttsClient.SynthesizeStream(ctx, text)
	if err != nil {
		s.eventLog.LogAsync(s.callID, eventlog.EventTTSError, map[string]any{
			"error":       err.Error(),
			"text_length": len(text),
		})
		return 0, err
	}

	// Send audio chunks to Twilio
	for chunk := range audioCh {
		select {
		case <-ctx.Done():
			s.eventLog.LogAsync(s.callID, eventlog.EventTTSCompleted, map[string]any{
				"text_length": len(text),
				"duration_ms": time.Since(ttsStartTime).Milliseconds(),
				"interrupted": true,
				"reason":      "context_cancelled",
			})
			return 0, ctx.Err()
		case <-s.bargeInCh:
			// Barge-in detected - stop sending audio
			s.logger.Printf("media_ws: stopping audio send due to barge-in")
			// Drain remaining audio
			for range audioCh {
			}
			s.eventLog.LogAsync(s.callID, eventlog.EventTTSCompleted, map[string]any{
				"text_length": len(text),
				"duration_ms": time.Since(ttsStartTime).Milliseconds(),
				"interrupted": true,
				"reason":      "barge_in",
			})
			return 0, nil
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
			s.eventLog.LogAsync(s.callID, eventlog.EventTTSError, map[string]any{
				"error":       err.Error(),
				"text_length": len(text),
				"duration_ms": time.Since(ttsStartTime).Milliseconds(),
			})
			return 0, fmt.Errorf("failed to send audio: %w", err)
		}
	}

	// Send mark to track completion.
	// We track audio in-flight by counting marks we send and marks we receive.
	markID := s.nextAudioMarkID()
	_ = s.incAudioPending()
	mark := twilioMark{
		Event:     "mark",
		StreamSid: s.streamSid,
	}
	mark.Mark.Name = fmt.Sprintf("audio-%d", markID)

	s.connMu.Lock()
	err = s.conn.WriteJSON(mark)
	s.connMu.Unlock()

	if err != nil {
		// If the mark wasn't sent, don't keep the pending counter inflated.
		_ = s.decAudioPending()
	}

	s.eventLog.LogAsync(s.callID, eventlog.EventTTSCompleted, map[string]any{
		"text_length": len(text),
		"duration_ms": time.Since(ttsStartTime).Milliseconds(),
		"interrupted": false,
	})

	return markID, err
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

	// Twilio clears any queued audio. Marks that were already sent may not arrive,
	// so reset our local pending counter to avoid getting stuck in "speaking" mode.
	s.audioMu.Lock()
	s.audioPending = 0
	s.pendingDoneMarkID = 0
	s.audioMu.Unlock()

	// Also cancel any pending post-audio action (hang-up/forward).
	s.cancelPendingAction()

	s.logger.Printf("media_ws: sent clear command (barge-in)")
	return nil
}

// speakGreeting sends the initial greeting via ElevenLabs TTS
func (s *callSession) speakGreeting() {
	// Use tenant's greeting if available, otherwise fall back to config default
	var greeting string
	if s.tenantCfg.GreetingText != nil && *s.tenantCfg.GreetingText != "" {
		greeting = *s.tenantCfg.GreetingText
	} else if s.cfg.GreetingText != "" {
		greeting = s.cfg.GreetingText
	} else {
		greeting = "Dobrý den, tady asistentka Karen. Majitel telefonu teď nemůže přijmout hovor, ale můžu vám pro něj zanechat vzkaz - co od něj potřebujete?"
	}

	s.logger.Printf("media_ws: speaking greeting: %s", greeting)

	// Add greeting to conversation history so LLM knows it was already said
	s.messagesMu.Lock()
	s.messages = append(s.messages, llm.Message{
		Role:    "assistant",
		Content: greeting,
	})
	s.messagesMu.Unlock()

	// Store greeting as first utterance (sequence 1)
	s.utteranceSeq++ // Increments from 0 to 1
	startTime := time.Now().UTC()
	if s.callID != "" {
		if err := s.store.InsertUtterance(s.ctx, s.callID, store.Utterance{
			Speaker:     "agent",
			Text:        greeting,
			Sequence:    s.utteranceSeq,
			StartedAt:   &startTime,
			Interrupted: false,
		}); err != nil {
			s.logger.Printf("media_ws: failed to store greeting utterance: %v", err)
			sentry.CaptureException(err)
		}
	}

	if _, err := s.speakText(s.ctx, greeting); err != nil && !errors.Is(err, context.Canceled) {
		s.logger.Printf("media_ws: greeting TTS error: %v", err)
		sentry.CaptureException(err)
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

// forwardCall forwards the call to the tenant owner's verified phone number
func (s *callSession) forwardCall(ctx context.Context) {
	if s.callSid == "" || s.accountSid == "" || s.cfg.TwilioAuthToken == "" {
		s.logger.Printf("media_ws: cannot forward - missing callSid, accountSid, or auth token")
		return
	}

	// Use owner's verified phone number - no hardcoded fallback for security
	forwardNumber := s.tenantCfg.OwnerPhone
	if forwardNumber == "" {
		s.logger.Printf("media_ws: cannot forward call %s - no owner phone configured for tenant", s.callSid)
		// Instead of forwarding to a random number, just hang up gracefully
		s.hangUpCall(ctx)
		return
	}

	// Wait for the forwarding message to finish playing
	select {
	case <-s.goodbyeDone:
		s.logger.Printf("media_ws: forward audio finished, forwarding call to %s", forwardNumber)
	case <-time.After(10 * time.Second):
		s.logger.Printf("media_ws: timeout waiting for forward audio, forwarding anyway to %s", forwardNumber)
	case <-ctx.Done():
		s.logger.Printf("media_ws: forwarding cancelled")
		return
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
		sentry.CaptureException(err)
		return
	}

	req.SetBasicAuth(s.accountSid, s.cfg.TwilioAuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Printf("media_ws: failed to forward call: %v", err)
		sentry.CaptureException(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.logger.Printf("media_ws: call %s forwarded successfully to %s", s.callSid, forwardNumber)
		s.eventLog.LogAsync(s.callID, eventlog.EventCallForwarded, map[string]any{
			"forward_number": forwardNumber,
			"success":        true,
		})
	} else {
		s.logger.Printf("media_ws: forward returned status %d", resp.StatusCode)
		s.eventLog.LogAsync(s.callID, eventlog.EventCallForwarded, map[string]any{
			"forward_number": forwardNumber,
			"success":        false,
			"status_code":    resp.StatusCode,
		})
	}
}

// hangUpCall terminates the call via Twilio REST API
func (s *callSession) hangUpCall(ctx context.Context) {
	// Mark that agent is initiating hangup (prevents "caller" overwrite in stop handler)
	s.agentHungUp = true

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
	case <-ctx.Done():
		s.logger.Printf("media_ws: hangup cancelled")
		return
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
		sentry.CaptureException(err)
		return
	}

	req.SetBasicAuth(s.accountSid, s.cfg.TwilioAuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Printf("media_ws: failed to hang up call: %v", err)
		sentry.CaptureException(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.logger.Printf("media_ws: call %s hung up successfully (agent initiated)", s.callSid)
		s.eventLog.LogAsync(s.callID, eventlog.EventCallHangup, map[string]any{
			"initiated_by": "agent",
			"success":      true,
		})
		// Update database status to completed
		if s.callID != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.store.UpdateCallStatus(ctx, s.callSid, "completed", time.Now().UTC()); err != nil {
				s.logger.Printf("media_ws: failed to update call status: %v", err)
				sentry.CaptureException(err)
			}

			// Mark call as ended by agent
			if err := s.store.UpdateCallEndedBy(ctx, s.callSid, "agent"); err != nil {
				s.logger.Printf("media_ws: failed to update ended_by: %v", err)
				sentry.CaptureException(err)
			}
		}
	} else {
		s.logger.Printf("media_ws: hang up returned status %d", resp.StatusCode)
		s.eventLog.LogAsync(s.callID, eventlog.EventCallHangup, map[string]any{
			"initiated_by": "agent",
			"success":      false,
			"status_code":  resp.StatusCode,
		})
	}
}

func (s *callSession) analyzeCall() {
	if s.callID == "" {
		return
	}

	// Use background context since call context may be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.messagesMu.Lock()
	msgs := append([]llm.Message(nil), s.messages...)
	s.messagesMu.Unlock()

	result, err := s.llmClient.AnalyzeCall(ctx, msgs)
	if err != nil {
		s.logger.Printf("media_ws: analysis error: %v", err)
		sentry.CaptureException(err)
		return
	}

	// Convert entities to JSON
	entitiesJSON, _ := json.Marshal(result.Entities)

	// Store screening result
	sr := store.ScreeningResult{
		LegitimacyLabel:      result.LegitimacyLabel,
		LegitimacyConfidence: result.LegitimacyConfidence,
		LeadLabel:            result.LeadLabel,
		IntentCategory:       result.IntentCategory,
		IntentText:           result.IntentText,
		EntitiesJSON:         entitiesJSON,
		CreatedAt:            time.Now().UTC(),
	}

	if err := s.store.InsertScreeningResult(ctx, s.callID, sr); err != nil {
		s.logger.Printf("media_ws: failed to store screening result: %v", err)
		sentry.CaptureException(err)
	} else {
		s.logger.Printf("media_ws: call classified as %s (%.0f%% confidence)",
			result.LegitimacyLabel, result.LegitimacyConfidence*100)

		// Send push notifications to tenant devices
		go s.sendPushNotifications(result.LegitimacyLabel, result.IntentText)
	}
}

// sendPushNotifications sends push notifications to all devices registered for the tenant
func (s *callSession) sendPushNotifications(legitimacyLabel, intentText string) {
	if s.apns == nil || s.tenantCfg.TenantID == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the call details to get the from_number
	call, err := s.store.GetCallDetail(ctx, s.callSid)
	if err != nil {
		s.logger.Printf("media_ws: failed to get call detail for push: %v", err)
		return
	}

	// Get all device tokens for the tenant
	tokens, err := s.store.GetTenantPushTokens(ctx, s.tenantCfg.TenantID)
	if err != nil {
		s.logger.Printf("media_ws: failed to get push tokens: %v", err)
		return
	}

	if len(tokens) == 0 {
		return
	}

	notif := notifications.CallNotification{
		CallID:          s.callSid,
		FromNumber:      call.FromNumber,
		IntentSummary:   intentText,
		LegitimacyLabel: legitimacyLabel,
	}

	// Send to all iOS devices
	for _, token := range tokens {
		if token.Platform == "ios" {
			if err := s.apns.SendCallNotification(token.Token, notif); err != nil {
				s.logger.Printf("media_ws: failed to send push to %s...: %v", token.Token[:16], err)
			}
		}
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
	s.messagesMu.Lock()
	msgCount := len(s.messages)
	s.messagesMu.Unlock()
	if msgCount >= 2 {
		s.analyzeCall()
	}

	// Mark call as completed (fallback in case hangUpCall didn't run or failed)
	if s.callID != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.store.UpdateCallStatus(ctx, s.callSid, "completed", time.Now().UTC())
	}

	// Track usage after call completes
	s.trackUsage()

	s.logger.Printf("media_ws: session cleaned up for call %s", s.callSid)
	s.eventLog.LogAsync(s.callID, eventlog.EventCallEnded, map[string]any{
		"call_sid": s.callSid,
	})
}

// trackUsage increments the tenant's usage counters after a call completes
func (s *callSession) trackUsage() {
	if s.callID == "" || s.tenantCfg.TenantID == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get call details to calculate duration
	call, err := s.store.GetCallDetail(ctx, s.callSid)
	if err != nil {
		s.logger.Printf("media_ws: failed to get call for usage tracking: %v", err)
		return
	}

	// Calculate call duration in seconds
	var durationSeconds int
	if call.EndedAt != nil {
		duration := call.EndedAt.Sub(call.StartedAt)
		durationSeconds = int(duration.Seconds())
	}

	// Check if call was spam/marketing (from screening result)
	isSpam := false
	if call.Screening != nil {
		label := strings.ToLower(call.Screening.LegitimacyLabel)
		isSpam = label == "spam" || label == "marketing" || label == "podvod"
	}

	// Increment usage
	if err := s.store.IncrementTenantUsage(ctx, s.tenantCfg.TenantID, durationSeconds, isSpam); err != nil {
		s.logger.Printf("media_ws: failed to track usage: %v", err)
		sentry.CaptureException(err)
	} else {
		s.logger.Printf("media_ws: tracked usage for tenant %s: %ds (spam=%t)", s.tenantCfg.TenantID, durationSeconds, isSpam)
	}

	// Check if we need to send usage warnings
	s.checkUsageWarnings(ctx)
}

// checkUsageWarnings sends push notifications if the tenant is approaching their limit
func (s *callSession) checkUsageWarnings(ctx context.Context) {
	if s.apns == nil {
		return // No APNs client configured
	}

	// Get the tenant to check current usage
	tenant, err := s.store.GetTenantByID(ctx, s.tenantCfg.TenantID)
	if err != nil {
		s.logger.Printf("media_ws: failed to get tenant for usage check: %v", err)
		return
	}

	// Get the plan limit
	limit := store.GetPlanCallLimit(tenant.Plan)
	if limit <= 0 {
		return // Unlimited plan, no warnings needed
	}

	callsUsed := tenant.CurrentPeriodCalls

	// Calculate thresholds
	threshold80 := int(float64(limit) * 0.8)

	// Check if we just crossed the 80% threshold (previous call was below, now at or above)
	// We check if current is at 80% threshold exactly or if we're at limit
	shouldWarn80 := callsUsed == threshold80
	shouldWarnExpired := callsUsed == limit

	if !shouldWarn80 && !shouldWarnExpired {
		return // No warning needed
	}

	// Get all push tokens for users in this tenant
	tokens, err := s.store.GetTenantPushTokens(ctx, tenant.ID)
	if err != nil {
		s.logger.Printf("media_ws: failed to get push tokens for tenant %s: %v", tenant.ID, err)
		return
	}

	if len(tokens) == 0 {
		return // No push tokens registered
	}

	// Determine warning type
	var warningType notifications.UsageWarningType
	if shouldWarnExpired {
		warningType = notifications.UsageWarningExpired
	} else {
		warningType = notifications.UsageWarning80Percent
	}

	// Send notifications to all registered devices
	for _, token := range tokens {
		if token.Platform == "ios" {
			go func(deviceToken string) {
				if err := s.apns.SendUsageWarning(deviceToken, warningType, callsUsed, limit); err != nil {
					s.logger.Printf("media_ws: failed to send usage warning: %v", err)
				}
			}(token.Token)
		}
	}

	s.logger.Printf("media_ws: sent %s warning to %d devices for tenant %s", warningType, len(tokens), tenant.ID)
}
