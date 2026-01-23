package httpapi

import (
	"testing"
	"time"
)

func TestRobocallDetector_ProlongedSilence(t *testing.T) {
	cfg := RobocallConfig{
		SilenceThreshold:    100 * time.Millisecond, // Short for testing
		BargeInThreshold:    3,
		BargeInWindow:       15 * time.Second,
		RepetitionThreshold: 3,
	}
	d := NewRobocallDetector(cfg)

	// Record enough agent turns
	d.RecordAgentTurn()
	d.RecordAgentTurn()

	// Should not detect yet (threshold not reached)
	result := d.Check()
	if result.IsRobocall {
		t.Error("should not detect robocall before silence threshold")
	}

	// Wait for silence threshold
	time.Sleep(150 * time.Millisecond)

	result = d.Check()
	if !result.IsRobocall {
		t.Error("should detect robocall after prolonged silence")
	}
	if result.Reason != "prolonged_silence" {
		t.Errorf("expected reason 'prolonged_silence', got %q", result.Reason)
	}
}

func TestRobocallDetector_SilenceNotTriggeredWithSpeech(t *testing.T) {
	cfg := RobocallConfig{
		SilenceThreshold:    100 * time.Millisecond,
		BargeInThreshold:    3,
		BargeInWindow:       15 * time.Second,
		RepetitionThreshold: 3,
	}
	d := NewRobocallDetector(cfg)

	// Record agent turns
	d.RecordAgentTurn()
	d.RecordAgentTurn()

	// Record speech - should prevent silence detection
	d.RecordSpeech("Hello, this is a test")

	time.Sleep(150 * time.Millisecond)

	result := d.Check()
	if result.IsRobocall {
		t.Error("should not detect robocall when speech was recorded")
	}
}

func TestRobocallDetector_RapidBargeIns(t *testing.T) {
	cfg := RobocallConfig{
		SilenceThreshold:    30 * time.Second,
		BargeInThreshold:    3,
		BargeInWindow:       5 * time.Second,
		RepetitionThreshold: 3,
	}
	d := NewRobocallDetector(cfg)

	// Record 2 barge-ins - should not trigger yet
	d.RecordBargeIn()
	d.RecordBargeIn()

	result := d.Check()
	if result.IsRobocall {
		t.Error("should not detect robocall with only 2 barge-ins")
	}

	// Record 3rd barge-in - should trigger
	d.RecordBargeIn()

	result = d.Check()
	if !result.IsRobocall {
		t.Error("should detect robocall with 3 rapid barge-ins")
	}
	if result.Reason != "rapid_barge_ins" {
		t.Errorf("expected reason 'rapid_barge_ins', got %q", result.Reason)
	}
}

func TestRobocallDetector_BargeInWindowExpiry(t *testing.T) {
	cfg := RobocallConfig{
		SilenceThreshold:    30 * time.Second,
		BargeInThreshold:    3,
		BargeInWindow:       100 * time.Millisecond, // Short window for testing
		RepetitionThreshold: 3,
	}
	d := NewRobocallDetector(cfg)

	// Record 2 barge-ins
	d.RecordBargeIn()
	d.RecordBargeIn()

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Record 1 more - should not trigger since previous ones are outside window
	d.RecordBargeIn()

	result := d.Check()
	if result.IsRobocall {
		t.Error("should not detect robocall when barge-ins are outside window")
	}
}

func TestRobocallDetector_PhraseRepetition(t *testing.T) {
	cfg := RobocallConfig{
		SilenceThreshold:    30 * time.Second,
		BargeInThreshold:    3,
		BargeInWindow:       15 * time.Second,
		RepetitionThreshold: 3,
	}
	d := NewRobocallDetector(cfg)

	// Record same phrase twice - should not trigger
	d.RecordSpeech("Please wait for the next available agent")
	d.RecordSpeech("Please wait for the next available agent")

	result := d.Check()
	if result.IsRobocall {
		t.Error("should not detect robocall with only 2 repetitions")
	}

	// Record 3rd repetition - should trigger
	d.RecordSpeech("Please wait for the next available agent")

	result = d.Check()
	if !result.IsRobocall {
		t.Error("should detect robocall with 3 phrase repetitions")
	}
	if result.Reason != "phrase_repetition:please wait for the next available agent" {
		t.Errorf("unexpected reason: %q", result.Reason)
	}
}

func TestRobocallDetector_PhraseRepetitionNormalization(t *testing.T) {
	cfg := RobocallConfig{
		SilenceThreshold:    30 * time.Second,
		BargeInThreshold:    3,
		BargeInWindow:       15 * time.Second,
		RepetitionThreshold: 3,
	}
	d := NewRobocallDetector(cfg)

	// Record same phrase with different casing/spacing
	d.RecordSpeech("PLEASE WAIT FOR THE NEXT AGENT")
	d.RecordSpeech("please wait  for the next agent") // extra space
	d.RecordSpeech("Please Wait For The Next Agent")

	result := d.Check()
	if !result.IsRobocall {
		t.Error("should detect robocall with normalized phrase repetitions")
	}
}

func TestRobocallDetector_ShortPhrasesIgnored(t *testing.T) {
	cfg := RobocallConfig{
		SilenceThreshold:    30 * time.Second,
		BargeInThreshold:    3,
		BargeInWindow:       15 * time.Second,
		RepetitionThreshold: 3,
	}
	d := NewRobocallDetector(cfg)

	// Record short phrases (less than 3 words)
	d.RecordSpeech("Yes")
	d.RecordSpeech("Yes")
	d.RecordSpeech("Yes")

	result := d.Check()
	if result.IsRobocall {
		t.Error("should not detect robocall from short phrase repetitions")
	}
}

func TestRobocallDetector_HoldKeywords(t *testing.T) {
	cfg := RobocallConfig{
		HoldKeywords: []string{
			"nezavěšujte",
			"please hold",
		},
	}
	d := NewRobocallDetector(cfg)

	// No keyword
	result := d.CheckText("Hello, how can I help you?")
	if result.IsRobocall {
		t.Error("should not detect robocall without keywords")
	}

	// Contains keyword
	result = d.CheckText("Please hold while we connect you")
	if !result.IsRobocall {
		t.Error("should detect robocall with hold keyword")
	}
	if result.Reason != "hold_keyword:please hold" {
		t.Errorf("unexpected reason: %q", result.Reason)
	}

	// Case insensitive
	result = d.CheckText("NEZAVĚŠUJTE prosím")
	if !result.IsRobocall {
		t.Error("should detect robocall with keyword (case insensitive)")
	}
}

func TestRobocallDetector_DisabledThresholds(t *testing.T) {
	cfg := RobocallConfig{
		SilenceThreshold:    30 * time.Second,
		BargeInThreshold:    0, // Disabled
		BargeInWindow:       15 * time.Second,
		RepetitionThreshold: 0, // Disabled
	}
	d := NewRobocallDetector(cfg)

	// Record many barge-ins - should not trigger
	for i := 0; i < 10; i++ {
		d.RecordBargeIn()
	}

	// Record many repetitions - should not trigger
	for i := 0; i < 10; i++ {
		d.RecordSpeech("The same phrase over and over")
	}

	result := d.Check()
	if result.IsRobocall {
		t.Error("should not detect robocall when thresholds are disabled")
	}
}

func TestNormalizePhrase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", ""},                        // Too short
		{"One Two", ""},                            // Too short
		{"One Two Three", "one two three"},         // Just enough
		{"  EXTRA   SPACES  ", ""},                 // Still too short after normalization
		{"Hello This Is Test", "hello this is test"},
		{"  Hello   This    Is   Test  ", "hello this is test"},
	}

	for _, tc := range tests {
		result := normalizePhrase(tc.input)
		if result != tc.expected {
			t.Errorf("normalizePhrase(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}
