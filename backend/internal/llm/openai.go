package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const openaiAPIURL = "https://api.openai.com/v1/chat/completions"

// OpenAIClient implements the Client interface using OpenAI's API.
type OpenAIClient struct {
	apiKey       string
	model        string
	systemPrompt string
	httpClient   *http.Client
}

// OpenAIConfig holds configuration for the OpenAI client.
type OpenAIConfig struct {
	APIKey       string
	Model        string // e.g., "gpt-4o-mini"
	SystemPrompt string // Optional custom system prompt
}

// NewOpenAIClient creates a new OpenAI client.
func NewOpenAIClient(cfg OpenAIConfig) *OpenAIClient {
	model := cfg.Model
	if model == "" {
		model = "gpt-4o-mini"
	}
	systemPrompt := cfg.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = SystemPromptCzech
	}
	return &OpenAIClient{
		apiKey:       cfg.APIKey,
		model:        model,
		systemPrompt: systemPrompt,
		httpClient:   &http.Client{},
	}
}

// SetSystemPrompt sets a custom system prompt for this client.
func (c *OpenAIClient) SetSystemPrompt(prompt string) {
	if prompt != "" {
		c.systemPrompt = prompt
	}
}

// GetSystemPrompt returns the current system prompt.
func (c *OpenAIClient) GetSystemPrompt() string {
	return c.systemPrompt
}

func (c *OpenAIClient) systemPromptWithGuardrails() string {
	// Always include guardrails to keep turn-taking smooth.
	return VoiceGuardrailsCzech + "\n\n" + c.systemPrompt
}

// chatRequest represents an OpenAI chat completion request.
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Stream      bool          `json:"stream,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse represents an OpenAI chat completion response.
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// AnalyzeCall analyzes the conversation and returns a screening result.
func (c *OpenAIClient) AnalyzeCall(ctx context.Context, messages []Message) (*ScreeningResult, error) {
	// Build messages with system prompt and analysis request
	chatMsgs := []chatMessage{
		{Role: "system", Content: c.systemPromptWithGuardrails()},
	}

	for _, m := range messages {
		chatMsgs = append(chatMsgs, chatMessage{Role: m.Role, Content: m.Content})
	}

	// Add analysis request
	chatMsgs = append(chatMsgs, chatMessage{
		Role:    "user",
		Content: AnalysisPromptCzech,
	})

	req := chatRequest{
		Model:       c.model,
		Messages:    chatMsgs,
		Temperature: 0.3,
		MaxTokens:   500,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", openaiAPIURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(respBody))
	}

	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := chatResp.Choices[0].Message.Content

	// Parse JSON from response (handle potential markdown code blocks)
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result ScreeningResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse screening result: %w (content: %s)", err, content)
	}

	return &result, nil
}

// GenerateResponse generates a response based on the conversation.
func (c *OpenAIClient) GenerateResponse(ctx context.Context, messages []Message) (<-chan string, error) {
	// Build messages with system prompt
	chatMsgs := []chatMessage{
		{Role: "system", Content: c.systemPromptWithGuardrails()},
	}

	for _, m := range messages {
		chatMsgs = append(chatMsgs, chatMessage{Role: m.Role, Content: m.Content})
	}

	req := chatRequest{
		Model:       c.model,
		Messages:    chatMsgs,
		Stream:      true,
		Temperature: 0.5,
		MaxTokens:   90,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", openaiAPIURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(respBody))
	}

	ch := make(chan string, 100)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines and non-data lines
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var streamResp chatResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue
			}

			if len(streamResp.Choices) > 0 {
				content := streamResp.Choices[0].Delta.Content
				if content != "" {
					select {
					case <-ctx.Done():
						return
					case ch <- content:
					}
				}
			}
		}
	}()

	return ch, nil
}
