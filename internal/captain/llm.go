package captain

import (
	"context"
	"fmt"

	"github.com/iainlowe/capn/internal/common"
)

// LLMProvider defines the interface for language model providers
type LLMProvider interface {
	GenerateCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float64, error)
}

// Message represents a single message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Validate validates the message
func (m *Message) Validate() error {
	validator := common.NewValidator()
	validator.AddRule("role", common.Required("role"))
	validator.AddRule("content", common.Required("content"))
	validator.AddRule("role_value", common.OneOf("role", "system", "user", "assistant"))

	return validator.Validate(map[string]interface{}{
		"role":       m.Role,
		"content":    m.Content,
		"role_value": m.Role,
	})
}

// CompletionRequest represents a request to generate a completion
type CompletionRequest struct {
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	Model       string    `json:"model,omitempty"`
}

// Validate validates the completion request
func (r *CompletionRequest) Validate() error {
	// First validate basic fields
	validator := common.NewValidator()
	validator.AddRule("max_tokens", common.Positive("max_tokens"))
	validator.AddRule("temperature", common.Range("temperature", 0, 1))

	err := validator.Validate(map[string]interface{}{
		"max_tokens":  r.MaxTokens,
		"temperature": r.Temperature,
	})
	if err != nil {
		return err
	}

	// Validate messages
	if len(r.Messages) == 0 {
		return fmt.Errorf("messages cannot be empty")
	}

	for i, msg := range r.Messages {
		if err := msg.Validate(); err != nil {
			return fmt.Errorf("invalid message at index %d: %w", i, err)
		}
	}

	return nil
}

// CompletionResponse represents a response from a completion request
type CompletionResponse struct {
	Content      string            `json:"content"`
	TokensUsed   int               `json:"tokens_used"`
	Model        string            `json:"model"`
	FinishReason string            `json:"finish_reason"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
