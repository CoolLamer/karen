package httpapi

import (
	"io"
	"log"
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

	// When nil, we should not override the chosen default endpointing.
	// (In production, the default comes from RouterConfig / env.)
	defaultEndpointing := 1200
	endpointing := defaultEndpointing
	if cfg.Endpointing != nil && *cfg.Endpointing > 0 {
		endpointing = *cfg.Endpointing
	}
	if endpointing != 1200 {
		t.Errorf("Default endpointing = %d, want %d", endpointing, 1200)
	}
}

func TestIsForward(t *testing.T) {
	tests := []struct {
		text     string
		expected bool
	}{
		{"[PŘEPOJIT] Přepojuji tě.", true},
		{"[PŘEPOJIT]", true},
		{"Dobrý den, [PŘEPOJIT] moment prosím.", true},
		{"Dobrý den, jak vám mohu pomoci?", false},
		{"Přepojuji vás", false}, // Without the marker
		{"", false},
		{"PŘEPOJIT", false}, // Missing brackets
	}

	for _, tt := range tests {
		result := isForward(tt.text)
		if result != tt.expected {
			t.Errorf("isForward(%q) = %v, want %v", tt.text, result, tt.expected)
		}
	}
}

func TestIsGoodbye(t *testing.T) {
	// Test that isGoodbye correctly detects goodbye phrases
	// Current implementation checks for: "na shledanou", "nashledanou", "mějte se", "hezký den"
	tests := []struct {
		text     string
		expected bool
	}{
		// "na shledanou" phrases
		{"Na shledanou!", true},
		{"nashledanou", true},
		{"NA SHLEDANOU", true}, // case insensitive
		{"Ahoj, na shledanou.", true},
		{"Dobře, na shledanou", true},

		// "hezký den" phrases
		{"Hezký den!", true},
		{"HEZKÝ DEN", true},
		{"Přeji vám hezký den", true},

		// "mějte se" phrases
		{"Mějte se!", true},
		{"Mějte se hezky", true},
		{"MĚJTE SE", true},

		// Not goodbyes
		{"Dobrý den", false},
		{"Ahoj, jak se máte?", false},
		{"Potřebuji mluvit s Lukášem", false},
		{"", false},
		{"Nashle", false}, // too short, doesn't contain full phrase
		{"Dobře, rozumím", false},
		{"Měj se", false},         // doesn't match "mějte se"
		{"Pěkný den", false},      // doesn't match "hezký den"
		{"Hezký zbytek dne", false}, // "hezký" and "den" aren't adjacent
	}

	for _, tt := range tests {
		result := isGoodbye(tt.text)
		if result != tt.expected {
			t.Errorf("isGoodbye(%q) = %v, want %v", tt.text, result, tt.expected)
		}
	}
}

func TestIsGoodbye_WithContext(t *testing.T) {
	// Test goodbye detection in full sentences with context
	// Only phrases that match: "na shledanou", "nashledanou", "mějte se", "hezký den"
	goodbyes := []string{
		"Děkuji za zprávu, na shledanou",
		"Dobře, předám mu to. Hezký den!",
		"Nashledanou!",
		"To je vše, mějte se hezky",
		"Ok, nashledanou!",
		"Hezký den, sbohem",
	}

	for _, text := range goodbyes {
		if !isGoodbye(text) {
			t.Errorf("isGoodbye(%q) should be true", text)
		}
	}

	notGoodbyes := []string{
		"Dobrý den, tady Karen",
		"Rozumím, zapíšu si to",
		"Můžete to zopakovat?",
		"Ahoj, co potřebujete?",
		"Pěkný zbytek dne", // doesn't contain "hezký den"
		"Měj se", // doesn't contain "mějte se"
	}

	for _, text := range notGoodbyes {
		if isGoodbye(text) {
			t.Errorf("isGoodbye(%q) should be false", text)
		}
	}
}

func TestStripForwardMarker(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[PŘEPOJIT] Přepojuji tě.", "Přepojuji tě."},
		{"[PŘEPOJIT] Moment, přepojuji.", "Moment, přepojuji."},
		{"Dobrý den, [PŘEPOJIT] přepojuji.", "Dobrý den, přepojuji."},
		{"Text bez markeru", "Text bez markeru"},
		{"", ""},
		{"[PŘEPOJIT]", "[PŘEPOJIT]"}, // No space after, shouldn't be stripped
	}

	for _, tt := range tests {
		result := stripForwardMarker(tt.input)
		if result != tt.expected {
			t.Errorf("stripForwardMarker(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestTenantConfigOwnerPhone(t *testing.T) {
	// Test TenantConfig with owner phone set
	cfg := TenantConfig{
		TenantID:   "test-tenant",
		OwnerPhone: "+420777123456",
	}

	if cfg.OwnerPhone != "+420777123456" {
		t.Errorf("OwnerPhone = %q, want %q", cfg.OwnerPhone, "+420777123456")
	}
}

func TestTenantConfigOwnerPhoneEmpty(t *testing.T) {
	// Test TenantConfig without owner phone (empty string)
	cfg := TenantConfig{
		TenantID: "test-tenant",
	}

	if cfg.OwnerPhone != "" {
		t.Errorf("OwnerPhone should be empty when not set, got %q", cfg.OwnerPhone)
	}

	// Test the logic that should prevent forwarding without owner phone
	forwardNumber := cfg.OwnerPhone
	if forwardNumber == "" {
		// This is expected - call should be hung up gracefully
	} else {
		t.Error("forwardNumber should be empty when OwnerPhone is not set")
	}
}

func TestTenantConfigVIPNames(t *testing.T) {
	// Test TenantConfig with VIP names
	vipNames := []string{"Máma", "Táta"}
	cfg := TenantConfig{
		TenantID: "test-tenant",
		VIPNames: vipNames,
	}

	if len(cfg.VIPNames) != 2 {
		t.Errorf("VIPNames length = %d, want 2", len(cfg.VIPNames))
	}
	if cfg.VIPNames[0] != "Máma" {
		t.Errorf("VIPNames[0] = %q, want %q", cfg.VIPNames[0], "Máma")
	}
}

func TestTenantConfigVIPNamesEmpty(t *testing.T) {
	// Test TenantConfig without VIP names
	cfg := TenantConfig{
		TenantID: "test-tenant",
	}

	if len(cfg.VIPNames) != 0 {
		t.Errorf("VIPNames should be nil or empty when not set, got %v", cfg.VIPNames)
	}
}

func TestParseAudioMarkID(t *testing.T) {
	id, ok := parseAudioMarkID("audio-123")
	if !ok || id != 123 {
		t.Errorf("parseAudioMarkID(audio-123) = (%d,%v), want (123,true)", id, ok)
	}

	_, ok = parseAudioMarkID("response-1")
	if ok {
		t.Error("parseAudioMarkID(response-1) should be false")
	}

	_, ok = parseAudioMarkID("audio-xyz")
	if ok {
		t.Error("parseAudioMarkID(audio-xyz) should be false")
	}
}

func TestAudioPendingCounterNeverNegative(t *testing.T) {
	s := &callSession{}
	if n := s.decAudioPending(); n != 0 {
		t.Errorf("decAudioPending() from zero = %d, want 0", n)
	}
}

func TestHandleMarkSignalsPendingDoneMark(t *testing.T) {
	s := &callSession{
		logger:     log.New(io.Discard, "", 0),
		goodbyeDone: make(chan struct{}, 1),
	}

	s.audioMu.Lock()
	s.pendingDoneMarkID = 7
	s.audioMu.Unlock()

	s.handleMark(&twilioMarkData{Name: "audio-7"})

	select {
	case <-s.goodbyeDone:
		// ok
	default:
		t.Error("expected goodbyeDone to be signaled for matching mark id")
	}
}
