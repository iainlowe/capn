package captain

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenAIProvider(t *testing.T) {
	tests := []struct {
		name    string
		config  OpenAIConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: OpenAIConfig{
				APIKey: "test-key",
				Model:  "gpt-3.5-turbo",
			},
			wantErr: false,
		},
		{
			name: "empty API key",
			config: OpenAIConfig{
				APIKey: "",
				Model:  "gpt-3.5-turbo",
			},
			wantErr: true,
		},
		{
			name: "empty model",
			config: OpenAIConfig{
				APIKey: "test-key",
				Model:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewOpenAIProvider(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestOpenAIProvider_GenerateCompletion_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/chat/completions", r.URL.Path)
		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{{
				Message:      openai.ChatCompletionMessage{Role: "assistant", Content: "hello"},
				FinishReason: openai.FinishReasonStop,
			}},
			Usage: openai.Usage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2},
			Model: "gpt-3.5-turbo",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := OpenAIConfig{APIKey: "test", Model: "gpt-3.5-turbo", BaseURL: server.URL}
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	ctx := context.Background()
	req := CompletionRequest{Messages: []Message{{Role: "user", Content: "hi"}}, MaxTokens: 10, Temperature: 0.7}
	resp, err := provider.GenerateCompletion(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "hello", resp.Content)
	assert.Equal(t, "gpt-3.5-turbo", resp.Model)
	assert.Equal(t, 2, resp.TokensUsed)
}

func TestOpenAIProvider_GenerateEmbedding_Mock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/embeddings", r.URL.Path)
		resp := openai.EmbeddingResponse{
			Data: []openai.Embedding{{Embedding: []float32{0.1, 0.2}}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := OpenAIConfig{APIKey: "test", Model: "gpt-3.5-turbo", BaseURL: server.URL}
	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	ctx := context.Background()
	embedding, err := provider.GenerateEmbedding(ctx, "hello")
	require.NoError(t, err)
	assert.InDeltaSlice(t, []float64{0.1, 0.2}, embedding, 1e-6)
}

func TestOpenAIProvider_GenerateCompletion_Integration(t *testing.T) {
	// Skip integration test if no API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}

	config := OpenAIConfig{
		APIKey: apiKey,
		Model:  "gpt-3.5-turbo",
	}

	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	ctx := context.Background()
	req := CompletionRequest{
		Messages: []Message{
			{Role: "user", Content: "Say hello in one word."},
		},
		MaxTokens:   10,
		Temperature: 0.1,
	}

	resp, err := provider.GenerateCompletion(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
	assert.NotEmpty(t, resp.Model)
	assert.Greater(t, resp.TokensUsed, 0)
}

func TestOpenAIProvider_GenerateEmbedding_Integration(t *testing.T) {
	// Skip integration test if no API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: OPENAI_API_KEY not set")
	}

	config := OpenAIConfig{
		APIKey: apiKey,
		Model:  "gpt-3.5-turbo",
	}

	provider, err := NewOpenAIProvider(config)
	require.NoError(t, err)

	ctx := context.Background()
	text := "Hello world"

	embedding, err := provider.GenerateEmbedding(ctx, text)
	require.NoError(t, err)
	assert.NotEmpty(t, embedding)
	assert.Greater(t, len(embedding), 0)
	// OpenAI embeddings typically have 1536 dimensions for text-embedding-ada-002
	assert.Greater(t, len(embedding), 1000)
}

func TestOpenAIConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  OpenAIConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: OpenAIConfig{
				APIKey:      "test-key",
				Model:       "gpt-3.5-turbo",
				BaseURL:     "https://api.openai.com/v1",
				MaxRetries:  3,
				Temperature: 0.7,
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: OpenAIConfig{
				Model: "gpt-3.5-turbo",
			},
			wantErr: true,
		},
		{
			name: "missing model",
			config: OpenAIConfig{
				APIKey: "test-key",
			},
			wantErr: true,
		},
		{
			name: "invalid temperature",
			config: OpenAIConfig{
				APIKey:      "test-key",
				Model:       "gpt-3.5-turbo",
				Temperature: 2.0,
			},
			wantErr: true,
		},
		{
			name: "negative max retries",
			config: OpenAIConfig{
				APIKey:     "test-key",
				Model:      "gpt-3.5-turbo",
				MaxRetries: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
