package common

import "fmt"

// ConfigBuilder provides a fluent interface for building configurations
type ConfigBuilder[T any] struct {
	config     T
	validators []func(T) error
}

// NewConfigBuilder creates a new configuration builder
func NewConfigBuilder[T any](defaultConfig T) *ConfigBuilder[T] {
	return &ConfigBuilder[T]{
		config:     defaultConfig,
		validators: make([]func(T) error, 0),
	}
}

// With applies a configuration function
func (b *ConfigBuilder[T]) With(fn func(*T)) *ConfigBuilder[T] {
	fn(&b.config)
	return b
}

// Validate adds a validation function
func (b *ConfigBuilder[T]) Validate(fn func(T) error) *ConfigBuilder[T] {
	b.validators = append(b.validators, fn)
	return b
}

// Build validates and returns the final configuration
func (b *ConfigBuilder[T]) Build() (T, error) {
	for _, validator := range b.validators {
		if err := validator(b.config); err != nil {
			var zero T
			return zero, fmt.Errorf("configuration validation failed: %w", err)
		}
	}
	return b.config, nil
}

// Get returns the current configuration without validation (for chaining)
func (b *ConfigBuilder[T]) Get() T {
	return b.config
}
