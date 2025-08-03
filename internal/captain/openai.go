package captain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the LLMProvider interface using OpenAI's API
type OpenAIProvider struct {
	client       *openai.Client
	config       *Config
	retryAttempts int
	retryDelay    time.Duration
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config *Config) (*OpenAIProvider, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	
	if config.OpenAIAPIKey == "" {
		return nil, errors.New("OpenAI API key is required")
	}

	client := openai.NewClient(config.OpenAIAPIKey)
	
	return &OpenAIProvider{
		client:       client,
		config:       config,
		retryAttempts: config.RetryAttempts,
		retryDelay:    config.RetryDelay,
	}, nil
}

// GenerateCompletion generates a completion using OpenAI's API
func (p *OpenAIProvider) GenerateCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if len(req.Messages) == 0 {
		return nil, errors.New("messages cannot be empty")
	}

	// Convert our messages to OpenAI format
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Set defaults if not provided
	model := req.Model
	if model == "" {
		model = p.config.Model
	}
	
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.config.MaxTokens
	}
	
	temperature := req.Temperature
	if temperature == 0 {
		temperature = p.config.Temperature
	}

	openaiReq := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	var resp openai.ChatCompletionResponse
	var err error

	// Retry logic
	for attempt := 0; attempt <= p.retryAttempts; attempt++ {
		resp, err = p.client.CreateChatCompletion(ctx, openaiReq)
		if err == nil {
			break
		}

		if attempt < p.retryAttempts {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(p.retryDelay * time.Duration(attempt+1)):
				// Exponential backoff
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate completion after %d attempts: %w", p.retryAttempts+1, err)
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no completion choices returned")
	}

	return &CompletionResponse{
		Content:    resp.Choices[0].Message.Content,
		TokensUsed: resp.Usage.TotalTokens,
		Model:      resp.Model,
	}, nil
}

// GenerateEmbedding generates an embedding for the given text
func (p *OpenAIProvider) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if text == "" {
		return nil, errors.New("text cannot be empty")
	}

	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.AdaEmbeddingV2,
	}

	var resp openai.EmbeddingResponse
	var err error

	// Retry logic
	for attempt := 0; attempt <= p.retryAttempts; attempt++ {
		resp, err = p.client.CreateEmbeddings(ctx, req)
		if err == nil {
			break
		}

		if attempt < p.retryAttempts {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(p.retryDelay * time.Duration(attempt+1)):
				// Exponential backoff
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding after %d attempts: %w", p.retryAttempts+1, err)
	}

	if len(resp.Data) == 0 {
		return nil, errors.New("no embedding data returned")
	}

	// Convert []float32 to []float64
	embedding32 := resp.Data[0].Embedding
	embedding64 := make([]float64, len(embedding32))
	for i, v := range embedding32 {
		embedding64[i] = float64(v)
	}

	return embedding64, nil
}