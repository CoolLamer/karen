package httpapi

import (
	"testing"
	"time"
)

func TestGetRandomFiller(t *testing.T) {
	// Test that getRandomFiller returns a valid filler word
	validFillers := map[string]bool{
		"Jasně...":   true,
		"Rozumím...": true,
		"Hmm...":     true,
		"Aha...":     true,
		"Dobře...":   true,
	}

	// Run multiple times to verify randomness works
	for i := 0; i < 100; i++ {
		filler := getRandomFiller()
		if !validFillers[filler] {
			t.Errorf("getRandomFiller() = %q, not in valid fillers", filler)
		}
	}
}

func TestFillerWordsNotEmpty(t *testing.T) {
	// Ensure fillerWords slice is not empty
	if len(fillerWords) == 0 {
		t.Error("fillerWords should not be empty")
	}
}

func TestFillerWordsAllHaveContent(t *testing.T) {
	// Ensure all filler words have content
	for i, filler := range fillerWords {
		if filler == "" {
			t.Errorf("fillerWords[%d] is empty", i)
		}
	}
}

func TestShouldSpeakFiller_Cooldown(t *testing.T) {
	// Test that filler is skipped during cooldown period
	recentTime := time.Now().Add(-5 * time.Second) // 5 seconds ago, within 10s cooldown

	// With recent filler, should always return false
	for i := 0; i < 100; i++ {
		if shouldSpeakFiller(recentTime) {
			t.Error("shouldSpeakFiller should return false during cooldown period")
			return
		}
	}
}

func TestShouldSpeakFiller_AfterCooldown(t *testing.T) {
	// Test that filler can be spoken after cooldown (with some randomness)
	oldTime := time.Now().Add(-15 * time.Second) // 15 seconds ago, past 10s cooldown

	// After cooldown, should return true sometimes (70% of the time on average)
	trueCount := 0
	iterations := 1000
	for i := 0; i < iterations; i++ {
		if shouldSpeakFiller(oldTime) {
			trueCount++
		}
	}

	// Should be roughly 70% (allow 50-90% range for randomness)
	ratio := float64(trueCount) / float64(iterations)
	if ratio < 0.5 || ratio > 0.9 {
		t.Errorf("shouldSpeakFiller returned true %.0f%% of the time, expected ~70%%", ratio*100)
	}
}

func TestShouldSpeakFiller_ZeroTime(t *testing.T) {
	// Test with zero time (no previous filler)
	zeroTime := time.Time{}

	// With no previous filler, should return true sometimes
	trueCount := 0
	iterations := 100
	for i := 0; i < iterations; i++ {
		if shouldSpeakFiller(zeroTime) {
			trueCount++
		}
	}

	// Should get at least some true values (past cooldown, just random skip)
	if trueCount == 0 {
		t.Error("shouldSpeakFiller with zero time should return true sometimes")
	}
}

func TestFillerSkipProbability(t *testing.T) {
	// Verify the skip probability constant is reasonable
	if fillerSkipProbability < 0 || fillerSkipProbability > 1 {
		t.Errorf("fillerSkipProbability = %f, should be between 0 and 1", fillerSkipProbability)
	}
}

func TestFillerCooldown(t *testing.T) {
	// Verify the cooldown constant is reasonable (between 1s and 60s)
	if fillerCooldown < time.Second || fillerCooldown > time.Minute {
		t.Errorf("fillerCooldown = %v, should be between 1s and 60s", fillerCooldown)
	}
}

func TestIsSentenceEnd(t *testing.T) {
	tests := []struct {
		text     string
		expected bool
	}{
		{"Hello.", true},
		{"Hello!", true},
		{"Hello?", true},
		{"Hello", false},
		{"Hello,", false},
		{"", false},
		{"   ", false},
		{"Hello. ", true}, // trailing space should still detect
		{"Jasně...", true}, // Czech filler with ellipsis ends with period
	}

	for _, tt := range tests {
		result := isSentenceEnd(tt.text)
		if result != tt.expected {
			t.Errorf("isSentenceEnd(%q) = %v, want %v", tt.text, result, tt.expected)
		}
	}
}

func TestExtractCompleteSentences(t *testing.T) {
	tests := []struct {
		buffer            string
		expectedComplete  string
		expectedRemaining string
	}{
		// Single complete sentence
		{"Hello world.", "Hello world.", ""},
		// Sentence with remaining text
		{"Hello world. How are", "Hello world.", " How are"},
		// Multiple sentences
		{"First. Second. Third", "First. Second.", " Third"},
		// No complete sentence
		{"Hello world", "", "Hello world"},
		// Empty buffer
		{"", "", ""},
		// Only punctuation
		{".", ".", ""},
		// Question mark
		{"Is it working?", "Is it working?", ""},
		// Exclamation mark
		{"Great!", "Great!", ""},
		// Mixed punctuation
		{"Hello! How are you? I'm fine.", "Hello! How are you? I'm fine.", ""},
	}

	for _, tt := range tests {
		complete, remaining := extractCompleteSentences(tt.buffer)
		if complete != tt.expectedComplete {
			t.Errorf("extractCompleteSentences(%q) complete = %q, want %q", tt.buffer, complete, tt.expectedComplete)
		}
		if remaining != tt.expectedRemaining {
			t.Errorf("extractCompleteSentences(%q) remaining = %q, want %q", tt.buffer, remaining, tt.expectedRemaining)
		}
	}
}

func TestExtractCompleteSentences_StreamingSimulation(t *testing.T) {
	// Simulate streaming chunks arriving
	chunks := []string{
		"Dobr",
		"ý den, ",
		"jak se máte?",
		" Já jsem Kar",
		"en.",
	}

	var buffer string
	var allComplete []string

	for _, chunk := range chunks {
		buffer += chunk
		complete, remaining := extractCompleteSentences(buffer)
		if complete != "" {
			allComplete = append(allComplete, complete)
		}
		buffer = remaining
	}

	// Should have extracted two complete sentences
	if len(allComplete) != 2 {
		t.Errorf("expected 2 complete sentences, got %d: %v", len(allComplete), allComplete)
	}

	expectedSentences := []string{
		"Dobrý den, jak se máte?",
		" Já jsem Karen.",
	}

	for i, expected := range expectedSentences {
		if i < len(allComplete) && allComplete[i] != expected {
			t.Errorf("sentence %d = %q, want %q", i, allComplete[i], expected)
		}
	}
}

func TestTwilioClearStructure(t *testing.T) {
	// Test that twilioClear struct is correctly formatted for JSON
	clear := twilioClear{
		Event:     "clear",
		StreamSid: "MZ123456",
	}

	if clear.Event != "clear" {
		t.Errorf("Event = %q, want %q", clear.Event, "clear")
	}
	if clear.StreamSid != "MZ123456" {
		t.Errorf("StreamSid = %q, want %q", clear.StreamSid, "MZ123456")
	}
}

func TestBargeInChannelBuffered(t *testing.T) {
	// Test that barge-in channel is buffered and non-blocking
	ch := make(chan string, 1) // Same as in session creation

	// First send should not block
	select {
	case ch <- "interrupting text":
		// Expected
	default:
		t.Error("buffered channel should not block on first send")
	}

	// Channel should now be full, second send should go to default
	select {
	case ch <- "second text":
		t.Error("channel should be full, second send should not succeed")
	default:
		// Expected - channel is full
	}

	// Read the value
	text := <-ch
	if text != "interrupting text" {
		t.Errorf("got %q, want %q", text, "interrupting text")
	}
}

func TestTenantConfigEndpointing(t *testing.T) {
	// Test TenantConfig with custom endpointing
	endpointing := 600
	cfg := TenantConfig{
		TenantID:    "test-tenant",
		Endpointing: &endpointing,
	}

	if cfg.Endpointing == nil {
		t.Error("Endpointing should not be nil")
	}
	if *cfg.Endpointing != 600 {
		t.Errorf("Endpointing = %d, want %d", *cfg.Endpointing, 600)
	}
}

func TestTenantConfigEndpointingNil(t *testing.T) {
	// Test TenantConfig without endpointing (should use default)
	cfg := TenantConfig{
		TenantID: "test-tenant",
	}

	if cfg.Endpointing != nil {
		t.Error("Endpointing should be nil when not set")
	}

	// When nil, default 800ms should be used in handleStart
	defaultEndpointing := 800
	endpointing := defaultEndpointing
	if cfg.Endpointing != nil && *cfg.Endpointing > 0 {
		endpointing = *cfg.Endpointing
	}
	if endpointing != 800 {
		t.Errorf("Default endpointing = %d, want %d", endpointing, 800)
	}
}
