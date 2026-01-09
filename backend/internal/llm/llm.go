package llm

import "context"

// ScreeningResult contains the LLM's analysis of a call.
type ScreeningResult struct {
	LegitimacyLabel      string            `json:"legitimacy_label"`      // legitimní, marketing, spam, podvod
	LegitimacyConfidence float64           `json:"legitimacy_confidence"` // 0-1
	IntentCategory       string            `json:"intent_category"`       // obchodní, osobní, servis, etc.
	IntentText           string            `json:"intent_text"`           // Brief description in Czech
	Entities             map[string]string `json:"entities"`              // Extracted entities (name, company, etc.)
	SuggestedResponse    string            `json:"suggested_response"`    // What the agent should say
	ShouldEndCall        bool              `json:"should_end_call"`       // Whether to end the call
}

// Message represents a conversation message.
type Message struct {
	Role    string // "system", "user", "assistant"
	Content string
}

// Client defines the interface for LLM providers.
type Client interface {
	// AnalyzeCall analyzes the conversation and returns a screening result.
	AnalyzeCall(ctx context.Context, messages []Message) (*ScreeningResult, error)

	// GenerateResponse generates a response based on the conversation.
	// Returns the response text streamed through the channel.
	GenerateResponse(ctx context.Context, messages []Message) (<-chan string, error)
}
