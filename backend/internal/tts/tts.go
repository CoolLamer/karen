package tts

import "context"

// Client defines the interface for text-to-speech providers.
type Client interface {
	// Synthesize converts text to speech and returns audio data.
	// The returned audio is in the format specified by the provider config.
	Synthesize(ctx context.Context, text string) ([]byte, error)

	// SynthesizeStream converts text to speech and streams audio chunks.
	// Each chunk is sent to the returned channel.
	SynthesizeStream(ctx context.Context, text string) (<-chan []byte, error)
}
