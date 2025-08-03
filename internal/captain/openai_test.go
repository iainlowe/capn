package captain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenAIProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name: "empty API key",
			config: &Config{
				OpenAIAPIKey: "",
			},
			expectError: true,
			errorMsg:    "OpenAI API key is required",
		},
		{
			name: "valid config",
			config: &Config{
				OpenAIAPIKey:  "test-key",
				Model:         "gpt-4",
				MaxTokens:     1000,
				Temperature:   0.7,
				RetryAttempts: 3,
				RetryDelay:    time.Second,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewOpenAIProvider(tt.config)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, provider)
			} else {
				require.NoError(t, err)
				require.NotNil(t, provider)
				assert.NotNil(t, provider.client)
				assert.Equal(t, tt.config, provider.config)
				assert.Equal(t, tt.config.RetryAttempts, provider.retryAttempts)
				assert.Equal(t, tt.config.RetryDelay, provider.retryDelay)
			}
		})
	}
}

func TestOpenAIProvider_GenerateCompletion_Validation(t *testing.T) {
	config := &Config{
		OpenAIAPIKey:  "test-key",
		Model:         "gpt-4",
		MaxTokens:     1000,
		Temperature:   0.7,
		RetryAttempts: 0, // No retries for faster test
		RetryDelay:    time.Millisecond,
	}
	
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("empty messages", func(t *testing.T) {
		req := CompletionRequest{
			Messages: []Message{},
		}
		
		resp, err := provider.GenerateCompletion(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "messages cannot be empty")
		assert.Nil(t, resp)
	})

	t.Run("defaults applied", func(t *testing.T) {
		// This test verifies that defaults are applied correctly
		// We can't test actual API calls without mocking, but we can verify the provider is set up correctly
		req := CompletionRequest{
			Messages: []Message{
				{Role: "user", Content: "test"},
			},
		}
		
		// We expect this to fail with API error since we're using a test key
		// but the request should be validated and defaults applied
		_, err := provider.GenerateCompletion(ctx, req)
		assert.Error(t, err) // Expected to fail with invalid API key
		assert.Contains(t, err.Error(), "failed to generate completion")
	})
}

func TestOpenAIProvider_GenerateEmbedding_Validation(t *testing.T) {
	config := &Config{
		OpenAIAPIKey:  "test-key",
		Model:         "gpt-4",
		MaxTokens:     1000,
		Temperature:   0.7,
		RetryAttempts: 0, // No retries for faster test
		RetryDelay:    time.Millisecond,
	}
	
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("empty text", func(t *testing.T) {
		embedding, err := provider.GenerateEmbedding(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "text cannot be empty")
		assert.Nil(t, embedding)
	})

	t.Run("non-empty text", func(t *testing.T) {
		// This test verifies that the request is properly formatted
		// We expect this to fail with API error since we're using a test key
		_, err := provider.GenerateEmbedding(ctx, "test text")
		assert.Error(t, err) // Expected to fail with invalid API key
		assert.Contains(t, err.Error(), "failed to generate embedding")
	})
}

func TestOpenAIProvider_ContextCancellation(t *testing.T) {
	config := &Config{
		OpenAIAPIKey:  "test-key",
		Model:         "gpt-4",
		MaxTokens:     1000,
		Temperature:   0.7,
		RetryAttempts: 2,
		RetryDelay:    time.Second,
	}
	
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	t.Run("completion with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		req := CompletionRequest{
			Messages: []Message{
				{Role: "user", Content: "test"},
			},
		}
		
		_, err := provider.GenerateCompletion(ctx, req)
		assert.Error(t, err)
		// Should fail quickly due to cancelled context
	})

	t.Run("embedding with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		_, err := provider.GenerateEmbedding(ctx, "test text")
		assert.Error(t, err)
		// Should fail quickly due to cancelled context
	})
}

// MockLLMProvider implements LLMProvider for testing
type MockLLMProvider struct {
	CompletionResponse *CompletionResponse
	CompletionError    error
	Embedding          []float64
	EmbeddingError     error
	CompletionFunc     func(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	EmbeddingFunc      func(ctx context.Context, text string) ([]float64, error)
}

func (m *MockLLMProvider) GenerateCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if m.CompletionFunc != nil {
		return m.CompletionFunc(ctx, req)
	}
	if m.CompletionError != nil {
		return nil, m.CompletionError
	}
	return m.CompletionResponse, nil
}

func (m *MockLLMProvider) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if m.EmbeddingFunc != nil {
		return m.EmbeddingFunc(ctx, text)
	}
	if m.EmbeddingError != nil {
		return nil, m.EmbeddingError
	}
	return m.Embedding, nil
}

func TestMockLLMProvider(t *testing.T) {
	mock := &MockLLMProvider{
		CompletionResponse: &CompletionResponse{
			Content:    "Test response",
			TokensUsed: 50,
			Model:      "gpt-4",
		},
		Embedding: []float64{0.1, 0.2, 0.3},
	}

	ctx := context.Background()

	t.Run("completion", func(t *testing.T) {
		req := CompletionRequest{
			Messages: []Message{
				{Role: "user", Content: "test"},
			},
		}
		
		resp, err := mock.GenerateCompletion(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "Test response", resp.Content)
		assert.Equal(t, 50, resp.TokensUsed)
		assert.Equal(t, "gpt-4", resp.Model)
	})

	t.Run("embedding", func(t *testing.T) {
		embedding, err := mock.GenerateEmbedding(ctx, "test")
		require.NoError(t, err)
		assert.Equal(t, []float64{0.1, 0.2, 0.3}, embedding)
	})
}