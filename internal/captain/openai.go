package captain

import (
	"context"
	"fmt"

	"github.com/iainlowe/capn/internal/common"
)

// OpenAIConfig holds configuration for the OpenAI provider
type OpenAIConfig struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	BaseURL     string  `yaml:"base_url,omitempty"`
	MaxRetries  int     `yaml:"max_retries"`
	Temperature float64 `yaml:"temperature"`
}

// Validate validates the OpenAI configuration using the validation framework
func (c *OpenAIConfig) Validate() error {
	validator := common.NewValidator()
	validator.AddRule("api_key", common.Required("api_key"))
	validator.AddRule("model", common.Required("model"))
	validator.AddRule("temperature", common.Range("temperature", 0, 1))
	validator.AddRule("max_retries", common.Positive("max_retries"))

	return validator.Validate(map[string]interface{}{
		"api_key":     c.APIKey,
		"model":       c.Model,
		"temperature": c.Temperature,
		"max_retries": c.MaxRetries,
	})
}

// OpenAIProvider implements LLMProvider using OpenAI's API
// TODO: Update to work with the official openai-go library
type OpenAIProvider struct {
	// client *openai.Client  // Commented out until API compatibility is resolved
	config OpenAIConfig
}

// NewOpenAIProvider creates a new OpenAI provider using the configuration builder pattern
func NewOpenAIProvider(config OpenAIConfig) (*OpenAIProvider, error) {
	// Use the configuration builder pattern with validation
	validatedConfig, err := common.NewConfigBuilder(config).
		With(func(c *OpenAIConfig) {
			// Set defaults
			if c.BaseURL == "" {
				c.BaseURL = "https://api.openai.com/v1"
			}
			if c.MaxRetries == 0 {
				c.MaxRetries = 3
			}
			if c.Temperature == 0 {
				c.Temperature = 0.7
			}
		}).
		Validate(func(c OpenAIConfig) error {
			return c.Validate()
		}).
		Build()

	if err != nil {
		return nil, fmt.Errorf("invalid OpenAI config: %w", err)
	}

	// TODO: Create actual OpenAI client when API compatibility is resolved
	return &OpenAIProvider{
		config: validatedConfig,
	}, nil
}

// GenerateCompletion generates a completion using OpenAI's API
func (p *OpenAIProvider) GenerateCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid completion request: %w", err)
	}

	// TODO: Implement actual OpenAI API call when library compatibility is resolved
	return nil, fmt.Errorf("OpenAI provider not yet implemented with official openai-go library")
}

// GenerateEmbedding generates embeddings using OpenAI's API
func (p *OpenAIProvider) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// TODO: Implement actual OpenAI API call when library compatibility is resolved
	return nil, fmt.Errorf("OpenAI embedding provider not yet implemented with official openai-go library")
}
