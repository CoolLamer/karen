package eventlog

import (
	"context"
	"testing"
)

func TestEventTypeConstants(t *testing.T) {
	// Verify all event types are defined as expected
	expectedEvents := map[EventType]string{
		EventCallStarted:      "call_started",
		EventSTTResult:        "stt_result",
		EventTurnFinalized:    "turn_finalized",
		EventBargeIn:          "barge_in",
		EventFillerSpoken:     "filler_spoken",
		EventFillerSkipped:    "filler_skipped",
		EventLLMStarted:       "llm_started",
		EventLLMCompleted:     "llm_completed",
		EventLLMError:         "llm_error",
		EventGoodbyeDetected:  "goodbye_detected",
		EventForwardDetected:  "forward_detected",
		EventCallForwarded:    "call_forwarded",
		EventCallHangup:       "call_hangup",
		EventCallEnded:        "call_ended",
		EventVADSpeechStarted: "vad_speech_started",
		EventVADUtteranceEnd:  "vad_utterance_end",
		EventMaxTurnTimeout:   "max_turn_timeout",
		EventTTSStarted:       "tts_started",
		EventTTSCompleted:     "tts_completed",
		EventTTSError:         "tts_error",
	}

	for eventType, expectedValue := range expectedEvents {
		if string(eventType) != expectedValue {
			t.Errorf("EventType %q = %q, want %q", expectedValue, string(eventType), expectedValue)
		}
	}
}

func TestLatencyDebugEventTypes(t *testing.T) {
	// Verify all latency debugging event types are defined
	latencyEvents := map[EventType]string{
		EventLLMFirstToken:     "llm_first_token",
		EventSentenceExtracted: "sentence_extracted",
		EventTTSFirstChunk:     "tts_first_chunk",
		EventFillerDecision:    "filler_decision",
	}

	for eventType, expectedValue := range latencyEvents {
		if string(eventType) != expectedValue {
			t.Errorf("Latency EventType %q = %q, want %q", expectedValue, string(eventType), expectedValue)
		}
	}
}

func TestSTTDiagnosticEventTypes(t *testing.T) {
	// Verify all STT diagnostic event types are defined
	diagnosticEvents := map[EventType]string{
		EventSTTEmptyStreak:       "stt_empty_streak",
		EventAudioSilenceDetected: "audio_silence_detected",
	}

	for eventType, expectedValue := range diagnosticEvents {
		if string(eventType) != expectedValue {
			t.Errorf("Diagnostic EventType %q = %q, want %q", expectedValue, string(eventType), expectedValue)
		}
	}
}

func TestLoggerNew(t *testing.T) {
	// Test that New returns a non-nil logger even with nil DB
	logger := New(nil)
	if logger == nil {
		t.Error("New(nil) should return a non-nil logger")
	}
}

func TestLoggerLogAsyncWithNilDB(t *testing.T) {
	// Test that LogAsync doesn't panic with nil DB
	logger := New(nil)

	// Should not panic
	logger.LogAsync("test-call-id", EventCallStarted, map[string]any{
		"test_key": "test_value",
	})
}

func TestLoggerLogAsyncWithEmptyCallID(t *testing.T) {
	// Test that LogAsync doesn't panic with empty call ID
	logger := New(nil)

	// Should not panic - silently skips
	logger.LogAsync("", EventCallStarted, map[string]any{
		"test_key": "test_value",
	})
}

func TestLoggerLogWithNilDB(t *testing.T) {
	// Test that Log returns nil error with nil DB
	logger := New(nil)

	err := logger.Log(context.Background(), "test-call-id", EventCallStarted, map[string]any{
		"test_key": "test_value",
	})

	if err != nil {
		t.Errorf("Log with nil DB should return nil error, got %v", err)
	}
}

func TestLoggerLogWithEmptyCallID(t *testing.T) {
	// Test that Log returns nil error with empty call ID
	logger := New(nil)

	err := logger.Log(context.Background(), "", EventCallStarted, map[string]any{
		"test_key": "test_value",
	})

	if err != nil {
		t.Errorf("Log with empty call ID should return nil error, got %v", err)
	}
}

func TestEventTypeStringConversion(t *testing.T) {
	// Test that event types can be converted to strings correctly
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventLLMFirstToken, "llm_first_token"},
		{EventSentenceExtracted, "sentence_extracted"},
		{EventTTSFirstChunk, "tts_first_chunk"},
		{EventFillerDecision, "filler_decision"},
	}

	for _, tt := range tests {
		if string(tt.eventType) != tt.expected {
			t.Errorf("string(%v) = %q, want %q", tt.eventType, string(tt.eventType), tt.expected)
		}
	}
}

func TestLatencyEventDataStructures(t *testing.T) {
	// Test that typical latency event data can be constructed
	logger := New(nil)

	// LLM first token event data
	llmFirstTokenData := map[string]any{
		"turn_id":    uint64(1),
		"latency_ms": int64(150),
	}
	logger.LogAsync("test-call", EventLLMFirstToken, llmFirstTokenData)

	// Sentence extracted event data
	sentenceExtractedData := map[string]any{
		"turn_id":        uint64(1),
		"sentence_num":   1,
		"text_length":    42,
		"buffer_wait_ms": int64(1800),
	}
	logger.LogAsync("test-call", EventSentenceExtracted, sentenceExtractedData)

	// TTS first chunk event data
	ttsFirstChunkData := map[string]any{
		"text_length": 42,
		"latency_ms":  int64(180),
	}
	logger.LogAsync("test-call", EventTTSFirstChunk, ttsFirstChunkData)

	// Filler decision event data - spoken
	fillerDecisionSpokenData := map[string]any{
		"turn_id":  uint64(1),
		"decision": "spoken",
		"reason":   "llm_slow",
		"delay_ms": int64(350),
		"filler":   "Jasně...",
	}
	logger.LogAsync("test-call", EventFillerDecision, fillerDecisionSpokenData)

	// Filler decision event data - skipped
	fillerDecisionSkippedData := map[string]any{
		"turn_id":  uint64(1),
		"decision": "skipped",
		"reason":   "llm_fast",
		"delay_ms": int64(100),
	}
	logger.LogAsync("test-call", EventFillerDecision, fillerDecisionSkippedData)
}
