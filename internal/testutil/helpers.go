package testutil

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ValidationTestCase represents a test case for validation functions
type ValidationTestCase[T any] struct {
	Name      string
	Input     T
	WantError bool
	ErrorMsg  string // Optional: specific error message to check
}

// RunValidationTests runs a table of validation test cases
func RunValidationTests[T any](t *testing.T, testCases []ValidationTestCase[T], validateFn func(T) error) {
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			err := validateFn(tt.Input)
			if tt.WantError {
				assert.Error(t, err)
				if tt.ErrorMsg != "" && err != nil {
					assert.Contains(t, err.Error(), tt.ErrorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TransformTestCase represents a test case for transformation functions
type TransformTestCase[I, O any] struct {
	Name      string
	Input     I
	Expected  O
	WantError bool
	ErrorMsg  string
}

// RunTransformTests runs a table of transformation test cases
func RunTransformTests[I, O comparable](t *testing.T, testCases []TransformTestCase[I, O], transformFn func(I) (O, error)) {
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			result, err := transformFn(tt.Input)
			if tt.WantError {
				assert.Error(t, err)
				if tt.ErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.ErrorMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.Expected, result)
			}
		})
	}
}

// FactoryTestCase represents a test case for factory functions
type FactoryTestCase[I, O any] struct {
	Name      string
	Input     I
	Validator func(t *testing.T, result O) // Custom validation function
	WantError bool
	ErrorMsg  string
}

// RunFactoryTests runs a table of factory test cases
func RunFactoryTests[I, O any](t *testing.T, testCases []FactoryTestCase[I, O], factoryFn func(I) (O, error)) {
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			result, err := factoryFn(tt.Input)
			if tt.WantError {
				assert.Error(t, err)
				if tt.ErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.ErrorMsg)
				}
			} else {
				require.NoError(t, err)
				if tt.Validator != nil {
					tt.Validator(t, result)
				}
			}
		})
	}
}

// AssertJSONRoundTrip tests that a struct can be marshaled to JSON and back
func AssertJSONRoundTrip[T comparable](t *testing.T, original T) {
	t.Helper()

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Unmarshal back
	var unmarshaled T
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, original, unmarshaled)
}
