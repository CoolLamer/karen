package llm

import (
	"strings"
	"testing"
)

func TestNewOpenAIClient(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		client := NewOpenAIClient(OpenAIConfig{
			APIKey: "test-key",
		})

		if client.model != "gpt-4o-mini" {
			t.Errorf("model = %q, want %q", client.model, "gpt-4o-mini")
		}

		if client.systemPrompt != SystemPromptCzech {
			t.Error("systemPrompt should default to SystemPromptCzech")
		}

		if client.apiKey != "test-key" {
			t.Errorf("apiKey = %q, want %q", client.apiKey, "test-key")
		}
	})

	t.Run("custom model", func(t *testing.T) {
		client := NewOpenAIClient(OpenAIConfig{
			APIKey: "test-key",
			Model:  "gpt-4o",
		})

		if client.model != "gpt-4o" {
			t.Errorf("model = %q, want %q", client.model, "gpt-4o")
		}
	})

	t.Run("custom system prompt", func(t *testing.T) {
		customPrompt := "Custom system prompt for testing"
		client := NewOpenAIClient(OpenAIConfig{
			APIKey:       "test-key",
			SystemPrompt: customPrompt,
		})

		if client.systemPrompt != customPrompt {
			t.Errorf("systemPrompt = %q, want %q", client.systemPrompt, customPrompt)
		}
	})
}

func TestSetSystemPrompt(t *testing.T) {
	client := NewOpenAIClient(OpenAIConfig{
		APIKey: "test-key",
	})

	originalPrompt := client.systemPrompt

	t.Run("set new prompt", func(t *testing.T) {
		newPrompt := "New custom prompt"
		client.SetSystemPrompt(newPrompt)

		if client.systemPrompt != newPrompt {
			t.Errorf("systemPrompt = %q, want %q", client.systemPrompt, newPrompt)
		}
	})

	t.Run("empty prompt does not change", func(t *testing.T) {
		currentPrompt := client.systemPrompt
		client.SetSystemPrompt("")

		if client.systemPrompt != currentPrompt {
			t.Error("empty prompt should not change current prompt")
		}
	})

	t.Run("restore original", func(t *testing.T) {
		client.SetSystemPrompt(originalPrompt)
		if client.systemPrompt != originalPrompt {
			t.Error("should be able to restore original prompt")
		}
	})
}

func TestGetSystemPrompt(t *testing.T) {
	customPrompt := "Test prompt for getter"
	client := NewOpenAIClient(OpenAIConfig{
		APIKey:       "test-key",
		SystemPrompt: customPrompt,
	})

	got := client.GetSystemPrompt()
	if got != customPrompt {
		t.Errorf("GetSystemPrompt() = %q, want %q", got, customPrompt)
	}
}

func TestSystemPromptCzech(t *testing.T) {
	// Verify the default system prompt contains expected elements
	// Note: This is now a generic fallback prompt without hardcoded user details
	prompt := SystemPromptCzech

	expectedPhrases := []string{
		"Karen",     // Agent name
		"TVŮJ ÚKOL", // Task section
		"PRAVIDLA",  // Rules section
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(prompt, phrase) {
			t.Errorf("SystemPromptCzech should contain %q", phrase)
		}
	}

	// The generic prompt should NOT contain hardcoded user-specific values
	unexpectedPhrases := []string{
		"Lukáš",                 // Specific owner name (should be tenant-specific)
		"nabidky@bauerlukas.cz", // Specific email (should be tenant-specific)
	}

	for _, phrase := range unexpectedPhrases {
		if strings.Contains(prompt, phrase) {
			t.Errorf("SystemPromptCzech (generic fallback) should NOT contain hardcoded value %q", phrase)
		}
	}
}

func TestAnalysisPromptCzech(t *testing.T) {
	// Verify the analysis prompt contains JSON structure
	prompt := AnalysisPromptCzech

	expectedFields := []string{
		"legitimacy_label",
		"legitimacy_confidence",
		"intent_category",
		"intent_text",
		"entities",
	}

	for _, field := range expectedFields {
		if !strings.Contains(prompt, field) {
			t.Errorf("AnalysisPromptCzech should contain field %q", field)
		}
	}
}

func TestClientInterface(t *testing.T) {
	// Verify OpenAIClient implements Client interface
	var _ Client = (*OpenAIClient)(nil)
}

func TestMessage(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Hello",
	}

	if msg.Role != "user" {
		t.Errorf("Role = %q, want %q", msg.Role, "user")
	}
	if msg.Content != "Hello" {
		t.Errorf("Content = %q, want %q", msg.Content, "Hello")
	}
}

func TestScreeningResult(t *testing.T) {
	result := ScreeningResult{
		LegitimacyLabel:      "legitimní",
		LegitimacyConfidence: 0.95,
		IntentCategory:       "obchodní",
		IntentText:           "Test intent",
		Entities: map[string]string{
			"name":  "Jan Novák",
			"phone": "+420777123456",
		},
		SuggestedResponse: "Děkuji",
		ShouldEndCall:     true,
	}

	if result.LegitimacyLabel != "legitimní" {
		t.Errorf("LegitimacyLabel = %q, want %q", result.LegitimacyLabel, "legitimní")
	}
	if result.LegitimacyConfidence != 0.95 {
		t.Errorf("LegitimacyConfidence = %f, want %f", result.LegitimacyConfidence, 0.95)
	}
	if result.Entities["name"] != "Jan Novák" {
		t.Errorf("Entities[name] = %q, want %q", result.Entities["name"], "Jan Novák")
	}
	if !result.ShouldEndCall {
		t.Error("ShouldEndCall should be true")
	}
}
