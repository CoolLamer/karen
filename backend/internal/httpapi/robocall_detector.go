package httpapi

import (
	"strings"
	"sync"
	"time"
)

// RobocallConfig holds configuration for robocall detection.
type RobocallConfig struct {
	SilenceThreshold    time.Duration // Time without speech to trigger detection (default 30s)
	BargeInThreshold    int           // Number of rapid barge-ins to trigger (default 3)
	BargeInWindow       time.Duration // Window for counting barge-ins (default 15s)
	RepetitionThreshold int           // Number of identical phrases to trigger (default 3)
	HoldKeywords        []string      // Keywords that indicate hold music/robocall
}

// DefaultRobocallConfig returns sensible defaults for robocall detection.
func DefaultRobocallConfig() RobocallConfig {
	return RobocallConfig{
		SilenceThreshold:    30 * time.Second,
		BargeInThreshold:    3,
		BargeInWindow:       15 * time.Second,
		RepetitionThreshold: 3,
		HoldKeywords: []string{
			"nezavěšujte",
			"do not hang up",
			"please hold",
			"moment prosím",
			"čekejte prosím",
			"please stay on the line",
			"your call is important",
			"vaše volání je důležité",
		},
	}
}

// RobocallDetector detects patterns indicative of robocalls or IVR systems.
type RobocallDetector struct {
	cfg RobocallConfig

	mu              sync.Mutex
	callStartTime   time.Time
	firstSpeechTime *time.Time
	lastSpeechTime  *time.Time
	bargeInTimes    []time.Time
	agentTurnCount  int
	phraseCount     map[string]int // normalized phrase -> count
}

// NewRobocallDetector creates a new detector with the given configuration.
func NewRobocallDetector(cfg RobocallConfig) *RobocallDetector {
	return &RobocallDetector{
		cfg:           cfg,
		callStartTime: time.Now(),
		phraseCount:   make(map[string]int),
		bargeInTimes:  make([]time.Time, 0, 10),
	}
}

// RecordSpeech records that speech was detected from the caller.
func (d *RobocallDetector) RecordSpeech(text string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	if d.firstSpeechTime == nil {
		d.firstSpeechTime = &now
	}
	d.lastSpeechTime = &now

	// Track phrase repetition
	normalized := normalizePhrase(text)
	if normalized != "" {
		d.phraseCount[normalized]++
	}
}

// RecordBargeIn records that the caller interrupted the agent.
func (d *RobocallDetector) RecordBargeIn() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.bargeInTimes = append(d.bargeInTimes, time.Now())
}

// RecordAgentTurn records that the agent completed a response turn.
func (d *RobocallDetector) RecordAgentTurn() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.agentTurnCount++
}

// DetectionResult contains the result of robocall detection.
type DetectionResult struct {
	IsRobocall bool
	Reason     string
}

// Check analyzes the accumulated signals and returns whether this appears to be a robocall.
func (d *RobocallDetector) Check() DetectionResult {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 1. Check for prolonged silence (no speech after call started)
	if d.agentTurnCount >= 2 && d.firstSpeechTime == nil {
		elapsed := time.Since(d.callStartTime)
		if elapsed >= d.cfg.SilenceThreshold {
			return DetectionResult{
				IsRobocall: true,
				Reason:     "prolonged_silence",
			}
		}
	}

	// 2. Check for rapid barge-ins (robocalls often have fixed timing)
	if d.cfg.BargeInThreshold > 0 {
		recentBargeIns := d.countRecentBargeIns()
		if recentBargeIns >= d.cfg.BargeInThreshold {
			return DetectionResult{
				IsRobocall: true,
				Reason:     "rapid_barge_ins",
			}
		}
	}

	// 3. Check for phrase repetition (robocalls repeat the same message)
	if d.cfg.RepetitionThreshold > 0 {
		for phrase, count := range d.phraseCount {
			if count >= d.cfg.RepetitionThreshold {
				return DetectionResult{
					IsRobocall: true,
					Reason:     "phrase_repetition:" + phrase,
				}
			}
		}
	}

	return DetectionResult{IsRobocall: false}
}

// CheckText analyzes the given text for robocall keywords.
// This is a separate check that can be called on any speech segment.
func (d *RobocallDetector) CheckText(text string) DetectionResult {
	d.mu.Lock()
	keywords := d.cfg.HoldKeywords
	d.mu.Unlock()

	lower := strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(lower, strings.ToLower(keyword)) {
			return DetectionResult{
				IsRobocall: true,
				Reason:     "hold_keyword:" + keyword,
			}
		}
	}

	return DetectionResult{IsRobocall: false}
}

// countRecentBargeIns counts barge-ins within the configured window.
func (d *RobocallDetector) countRecentBargeIns() int {
	if len(d.bargeInTimes) == 0 {
		return 0
	}

	cutoff := time.Now().Add(-d.cfg.BargeInWindow)
	count := 0
	for _, t := range d.bargeInTimes {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}

// normalizePhrase normalizes text for phrase comparison.
func normalizePhrase(text string) string {
	// Lowercase, trim whitespace, remove extra spaces
	text = strings.ToLower(strings.TrimSpace(text))
	// Collapse multiple spaces
	words := strings.Fields(text)
	if len(words) < 3 {
		// Don't track very short phrases
		return ""
	}
	return strings.Join(words, " ")
}
