package captain

import (
	"context"
	"fmt"
	"strings"
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
	if m.Role == "" {
		return fmt.Errorf("role cannot be empty")
	}
	if m.Content == "" {
		return fmt.Errorf("content cannot be empty")
	}
	
	validRoles := []string{"system", "user", "assistant"}
	for _, validRole := range validRoles {
		if m.Role == validRole {
			return nil
		}
	}
	
	return fmt.Errorf("invalid role: %s, must be one of: %s", m.Role, strings.Join(validRoles, ", "))
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
	if len(r.Messages) == 0 {
		return fmt.Errorf("messages cannot be empty")
	}
	
	for i, msg := range r.Messages {
		if err := msg.Validate(); err != nil {
			return fmt.Errorf("invalid message at index %d: %w", i, err)
		}
	}
	
	if r.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}
	
	if r.Temperature < 0 || r.Temperature > 1 {
		return fmt.Errorf("temperature must be between 0 and 1")
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