package captain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLLMProvider is a mock implementation of LLMProvider for testing
type MockLLMProvider struct {
	mock.Mock
}

func (m *MockLLMProvider) GenerateCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*CompletionResponse), args.Error(1)
}

func (m *MockLLMProvider) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	args := m.Called(ctx, text)
	return args.Get(0).([]float64), args.Error(1)
}

func TestCompletionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CompletionRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CompletionRequest{
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				MaxTokens:   1000,
				Temperature: 0.7,
			},
			wantErr: false,
		},
		{
			name: "empty messages",
			req: CompletionRequest{
				Messages:    []Message{},
				MaxTokens:   1000,
				Temperature: 0.7,
			},
			wantErr: true,
		},
		{
			name: "invalid temperature",
			req: CompletionRequest{
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				MaxTokens:   1000,
				Temperature: 2.0, // should be between 0 and 1
			},
			wantErr: true,
		},
		{
			name: "invalid max tokens",
			req: CompletionRequest{
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				MaxTokens:   -1,
				Temperature: 0.7,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMessage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		msg     Message
		wantErr bool
	}{
		{
			name: "valid user message",
			msg: Message{
				Role:    "user",
				Content: "Hello world",
			},
			wantErr: false,
		},
		{
			name: "valid assistant message",
			msg: Message{
				Role:    "assistant",
				Content: "Hello back!",
			},
			wantErr: false,
		},
		{
			name: "valid system message",
			msg: Message{
				Role:    "system",
				Content: "You are a helpful assistant",
			},
			wantErr: false,
		},
		{
			name: "empty role",
			msg: Message{
				Role:    "",
				Content: "Hello",
			},
			wantErr: true,
		},
		{
			name: "empty content",
			msg: Message{
				Role:    "user",
				Content: "",
			},
			wantErr: true,
		},
		{
			name: "invalid role",
			msg: Message{
				Role:    "invalid",
				Content: "Hello",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}