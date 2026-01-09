package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
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

	// Conversation state
	messages     []llm.Message
	utteranceSeq int
	speaking     bool // True when TTS is playing
	speakingMu   sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
}

func (r *Router) handleMediaWS(w http.ResponseWriter, req *http.Request) {
	// Check if we have required API keys
	if r.cfg.DeepgramAPIKey == "" || r.cfg.OpenAIAPIKey == "" || r.cfg.ElevenLabsAPIKey == "" {
		r.logger.Printf("media_ws: missing API keys")
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
		conn:       conn,
		store:      r.store,
		logger:     r.logger,
		cfg:        r.cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		messages:   []llm.Message{},
		ctx:        ctx,
		cancel:     cancel,
	}

	// Create LLM client (doesn't require connection)
	session.llmClient = llm.NewOpenAIClient(llm.OpenAIConfig{
		APIKey: r.cfg.OpenAIAPIKey,
		Model:  "gpt-4o-mini",
	})

	// Create TTS client
	session.ttsClient = tts.NewElevenLabsClient(tts.ElevenLabsConfig{
		APIKey:  r.cfg.ElevenLabsAPIKey,
		VoiceID: r.cfg.TTSVoiceID,
		ModelID: "eleven_flash_v2_5",
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

	s.logger.Printf("media_ws: stream started - StreamSid: %s, CallSid: %s", start.StreamSid, s.callSid)

	// Get call ID from database now that we have callSid
	if s.callSid != "" {
		callID, err := s.store.GetCallID(s.ctx, s.callSid)
		if err != nil {
			s.logger.Printf("media_ws: failed to get call ID for %s: %v", s.callSid, err)
		} else {
			s.callID = callID
		}
	}

	// Connect to Deepgram STT
	sttClient, err := stt.NewDeepgramClient(s.ctx, stt.DeepgramConfig{
		APIKey:      s.cfg.DeepgramAPIKey,
		Language:    "cs", // Czech
		Model:       "nova-3",
		SampleRate:  8000,
		Encoding:    "mulaw",
		Channels:    1,
		Punctuate:   true,
		Endpointing: 1200, // 1200ms silence for turn detection (longer for natural Czech speech)
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Deepgram: %w", err)
	}
	s.sttClient = sttClient

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
					s.logger.Printf("media_ws: caller said: %s", text)

					// Store utterance
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
							Interrupted:   false,
						})
					}

					// Add to conversation history
					s.messages = append(s.messages, llm.Message{
						Role:    "user",
						Content: text,
					})

					// Generate and speak response
					go s.generateAndSpeak()
				}

				// Reset for next utterance
				currentUtterance.Reset()
				utteranceStartTime = nil
			}
		}
	}
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

	// Collect full response
	var response strings.Builder
	for chunk := range responseCh {
		response.WriteString(chunk)
	}

	responseText := strings.TrimSpace(response.String())
	if responseText == "" {
		return
	}

	s.logger.Printf("media_ws: agent response: %s", responseText)

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

	// Synthesize and send audio
	if err := s.speakText(responseText); err != nil {
		s.logger.Printf("media_ws: TTS error: %v", err)
	}

	// If this was a goodbye, hang up the call
	if isGoodbye(responseText) {
		s.logger.Printf("media_ws: detected goodbye, hanging up call")
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

	greeting := s.cfg.GreetingText
	if greeting == "" {
		greeting = "Dobrý den, tady Asistentka Karen. Lukáš nemá čas, ale můžu vám pro něj zanechat vzkaz - co od něj potřebujete?"
	}

	s.logger.Printf("media_ws: speaking greeting: %s", greeting)

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
		"ahoj",
		"čau",
	}
	for _, phrase := range goodbyePhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

// hangUpCall terminates the call via Twilio REST API
func (s *callSession) hangUpCall() {
	if s.callSid == "" || s.accountSid == "" || s.cfg.TwilioAuthToken == "" {
		s.logger.Printf("media_ws: cannot hang up - missing callSid, accountSid, or auth token")
		return
	}

	// Wait for the TTS audio to finish playing (goodbye message takes ~4-5 seconds)
	time.Sleep(5 * time.Second)

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
	} else {
		s.logger.Printf("media_ws: hang up returned status %d", resp.StatusCode)
	}
}

func (s *callSession) analyzeCall() {
	if s.callID == "" {
		return
	}

	result, err := s.llmClient.AnalyzeCall(s.ctx, s.messages)
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

	if err := s.store.InsertScreeningResult(s.ctx, s.callID, sr); err != nil {
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

	s.logger.Printf("media_ws: session cleaned up for call %s", s.callSid)
}
