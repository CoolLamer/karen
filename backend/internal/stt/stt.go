package stt

import "context"

// TranscriptResult represents a speech-to-text transcription result.
type TranscriptResult struct {
	Text       string  // The transcribed text
	Confidence float64 // Confidence score (0-1)
	IsFinal    bool    // Whether this is a final or interim result
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
