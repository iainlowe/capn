package captain

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// OpenAIConfig holds configuration for the OpenAI provider
type OpenAIConfig struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	BaseURL     string  `yaml:"base_url,omitempty"`
	MaxRetries  int     `yaml:"max_retries"`
	Temperature float64 `yaml:"temperature"`
}

// Validate validates the OpenAI configuration
func (c *OpenAIConfig) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("api_key cannot be empty")
	}
	if c.Model == "" {
		return fmt.Errorf("model cannot be empty")
	}
	if c.Temperature < 0 || c.Temperature > 1 {
		return fmt.Errorf("temperature must be between 0 and 1")
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}
	return nil
}

// OpenAIProvider implements LLMProvider using OpenAI's API
type OpenAIProvider struct {
	client *openai.Client
	config OpenAIConfig
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config OpenAIConfig) (*OpenAIProvider, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid OpenAI config: %w", err)
	}

	// Set defaults
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	clientConfig := openai.DefaultConfig(config.APIKey)
	if config.BaseURL != "https://api.openai.com/v1" {
		clientConfig.BaseURL = config.BaseURL
	}

	client := openai.NewClientWithConfig(clientConfig)

	return &OpenAIProvider{
		client: client,
		config: config,
	}, nil
}

// GenerateCompletion generates a completion using OpenAI's API
func (p *OpenAIProvider) GenerateCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid completion request: %w", err)
	}

	// Convert our messages to OpenAI format
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Use model from request if specified, otherwise use config default
	model := req.Model
	if model == "" {
		model = p.config.Model
	}

	// Create OpenAI request
	openaiReq := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: float32(req.Temperature),
	}

	// Make the API call
	resp, err := p.client.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no completion choices returned from OpenAI")
	}

	// Extract the response
	choice := resp.Choices[0]
	return &CompletionResponse{
		Content:      choice.Message.Content,
		TokensUsed:   resp.Usage.TotalTokens,
		Model:        resp.Model,
		FinishReason: string(choice.FinishReason),
		Metadata: map[string]string{
			"prompt_tokens":     fmt.Sprintf("%d", resp.Usage.PromptTokens),
			"completion_tokens": fmt.Sprintf("%d", resp.Usage.CompletionTokens),
		},
	}, nil
}

// GenerateEmbedding generates embeddings using OpenAI's API
func (p *OpenAIProvider) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// Create embedding request
	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.AdaEmbeddingV2, // Use the standard embedding model
	}

	// Make the API call
	resp, err := p.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI embedding API error: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned from OpenAI")
	}

	// Convert float32 to float64
	embedding := resp.Data[0].Embedding
	result := make([]float64, len(embedding))
	for i, v := range embedding {
		result[i] = float64(v)
	}

	return result, nil
}