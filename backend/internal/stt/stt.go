package stt

import "context"

// TranscriptResult represents a speech-to-text transcription result.
type TranscriptResult struct {
	Text       string  // The transcribed text
	Confidence float64 // Confidence score (0-1)
	// SegmentFinal means Deepgram marked this transcript segment as final (`is_final=true`).
	// Note: multiple SegmentFinal segments can occur within a single user turn.
	SegmentFinal bool
	// SpeechFinal means Deepgram detected end-of-speech (`speech_final=true`).
	// This is the signal we should use to finalize a user turn.
	SpeechFinal bool
}

// Client defines the interface for speech-to-text providers.
type Client interface {
	// StreamAudio sends audio data to the STT service.
	// Audio should be in the format expected by the provider.
	StreamAudio(ctx context.Context, audio []byte) error

	// Results returns a channel that receives transcription results.
	Results() <-chan TranscriptResult

	// Errors returns a channel that receives errors.
	Errors() <-chan error

	// Close closes the connection to the STT service.
	Close() error
}
