package tts

import (
	"testing"
)

func TestNewElevenLabsClient_DefaultValues(t *testing.T) {
	// Test that default values are used when -1 (sentinel) is specified
	// This signals "use defaults" since 0.0 is now a valid value
	client := NewElevenLabsClient(ElevenLabsConfig{
		APIKey:     "test-key",
		Stability:  -1, // Sentinel for "use default"
		Similarity: -1, // Sentinel for "use default"
	})

	if client.voiceID != "21m00Tcm4TlvDq8ikWAM" {
		t.Errorf("voiceID = %q, want %q", client.voiceID, "21m00Tcm4TlvDq8ikWAM")
	}
	if client.modelID != "eleven_flash_v2_5" {
		t.Errorf("modelID = %q, want %q", client.modelID, "eleven_flash_v2_5")
	}
	if client.stability != 0.5 {
		t.Errorf("stability = %f, want %f", client.stability, 0.5)
	}
	if client.similarity != 0.75 {
		t.Errorf("similarity = %f, want %f", client.similarity, 0.75)
	}
}

func TestNewElevenLabsClient_CustomStability(t *testing.T) {
	// Test that custom stability is used
	client := NewElevenLabsClient(ElevenLabsConfig{
		APIKey:     "test-key",
		Stability:  0.8,
		Similarity: -1, // Use default for similarity
	})

	if client.stability != 0.8 {
		t.Errorf("stability = %f, want %f", client.stability, 0.8)
	}
	// Similarity should still be default
	if client.similarity != 0.75 {
		t.Errorf("similarity = %f, want %f", client.similarity, 0.75)
	}
}

func TestNewElevenLabsClient_CustomSimilarity(t *testing.T) {
	// Test that custom similarity is used
	client := NewElevenLabsClient(ElevenLabsConfig{
		APIKey:     "test-key",
		Stability:  -1, // Use default for stability
		Similarity: 0.9,
	})

	// Stability should still be default
	if client.stability != 0.5 {
		t.Errorf("stability = %f, want %f", client.stability, 0.5)
	}
	if client.similarity != 0.9 {
		t.Errorf("similarity = %f, want %f", client.similarity, 0.9)
	}
}

func TestNewElevenLabsClient_CustomBoth(t *testing.T) {
	// Test that both custom values are used
	client := NewElevenLabsClient(ElevenLabsConfig{
		APIKey:     "test-key",
		Stability:  0.3,
		Similarity: 0.6,
	})

	if client.stability != 0.3 {
		t.Errorf("stability = %f, want %f", client.stability, 0.3)
	}
	if client.similarity != 0.6 {
		t.Errorf("similarity = %f, want %f", client.similarity, 0.6)
	}
}

func TestNewElevenLabsClient_ZeroValuesAreValid(t *testing.T) {
	// Test that zero values are now valid (0.0 is a valid ElevenLabs setting)
	client := NewElevenLabsClient(ElevenLabsConfig{
		APIKey:     "test-key",
		Stability:  0,
		Similarity: 0,
	})

	// Zero values should be preserved (0.0 is valid for max expressiveness)
	if client.stability != 0 {
		t.Errorf("stability = %f, want %f (zero is now valid)", client.stability, 0.0)
	}
	if client.similarity != 0 {
		t.Errorf("similarity = %f, want %f (zero is now valid)", client.similarity, 0.0)
	}
}

func TestNewElevenLabsClient_NegativeValuesUseDefaults(t *testing.T) {
	// Test that negative values fall back to defaults (use -1 as sentinel)
	client := NewElevenLabsClient(ElevenLabsConfig{
		APIKey:     "test-key",
		Stability:  -1,
		Similarity: -1,
	})

	// Negative values should trigger defaults
	if client.stability != 0.5 {
		t.Errorf("stability = %f, want %f (default for negative)", client.stability, 0.5)
	}
	if client.similarity != 0.75 {
		t.Errorf("similarity = %f, want %f (default for negative)", client.similarity, 0.75)
	}
}

func TestNewElevenLabsClient_CustomVoiceAndModel(t *testing.T) {
	// Test custom voice and model
	client := NewElevenLabsClient(ElevenLabsConfig{
		APIKey:  "test-key",
		VoiceID: "custom-voice-id",
		ModelID: "custom-model-id",
	})

	if client.voiceID != "custom-voice-id" {
		t.Errorf("voiceID = %q, want %q", client.voiceID, "custom-voice-id")
	}
	if client.modelID != "custom-model-id" {
		t.Errorf("modelID = %q, want %q", client.modelID, "custom-model-id")
	}
}
