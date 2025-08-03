package captain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewCaptain(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultConfig()
	mockProvider := &MockLLMProvider{}

	tests := []struct {
		name        string
		id          string
		config      *Config
		provider    LLMProvider
		logger      *zap.Logger
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty ID",
			id:          "",
			config:      config,
			provider:    mockProvider,
			logger:      logger,
			expectError: true,
			errorMsg:    "captain ID cannot be empty",
		},
		{
			name:        "nil config",
			id:          "test-captain",
			config:      nil,
			provider:    mockProvider,
			logger:      logger,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name:        "nil provider",
			id:          "test-captain",
			config:      config,
			provider:    nil,
			logger:      logger,
			expectError: true,
			errorMsg:    "LLM provider cannot be nil",
		},
		{
			name:        "nil logger",
			id:          "test-captain",
			config:      config,
			provider:    mockProvider,
			logger:      nil,
			expectError: true,
			errorMsg:    "logger cannot be nil",
		},
		{
			name:        "valid parameters",
			id:          "test-captain",
			config:      config,
			provider:    mockProvider,
			logger:      logger,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			captain, err := NewCaptain(tt.id, tt.config, tt.provider, tt.logger)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, captain)
			} else {
				require.NoError(t, err)
				require.NotNil(t, captain)
				assert.Equal(t, tt.id, captain.GetID())
				assert.Equal(t, tt.config, captain.GetConfig())
				assert.Equal(t, tt.config.MaxConcurrentAgents, captain.GetMaxConcurrentAgents())
				assert.False(t, captain.IsShutdown())
			}
		})
	}
}

func TestCaptain_PlanGoal(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultConfig()

	t.Run("successful planning", func(t *testing.T) {
		mockResponse := `{
  "reasoning": "Simple file creation task",
  "tasks": [
    {
      "id": "task-1",
      "title": "Create file",
      "description": "Create test file",
      "command": "touch test.txt",
      "dependencies": [],
      "estimated_duration": "5",
      "priority": 1
    }
  ],
  "estimated_duration": "5"
}`
		
		mockProvider := &MockLLMProvider{
			CompletionResponse: &CompletionResponse{
				Content: mockResponse,
			},
		}
		
		captain, err := NewCaptain("test-captain", config, mockProvider, logger)
		require.NoError(t, err)
		
		plan, err := captain.PlanGoal(context.Background(), "create a test file")
		require.NoError(t, err)
		require.NotNil(t, plan)
		
		assert.Equal(t, "create a test file", plan.Goal)
		assert.Len(t, plan.Tasks, 1)
		assert.Equal(t, "task-1", plan.Tasks[0].ID)
		assert.Equal(t, "Create file", plan.Tasks[0].Title)
	})

	t.Run("planning failure", func(t *testing.T) {
		mockProvider := &MockLLMProvider{
			CompletionError: errors.New("API error"),
		}
		
		captain, err := NewCaptain("test-captain", config, mockProvider, logger)
		require.NoError(t, err)
		
		plan, err := captain.PlanGoal(context.Background(), "test goal")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "goal analysis failed")
		assert.Nil(t, plan)
	})
}

func TestCaptain_ValidatePlan(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultConfig()

	t.Run("valid plan", func(t *testing.T) {
		mockProvider := &MockLLMProvider{
			CompletionResponse: &CompletionResponse{
				Content: "VALID - Plan looks good",
			},
		}
		
		captain, err := NewCaptain("test-captain", config, mockProvider, logger)
		require.NoError(t, err)
		
		plan := &ExecutionPlan{
			ID:   "test-plan",
			Goal: "test",
			Tasks: []Task{
				{ID: "task-1", Title: "Test task"},
			},
		}
		
		err = captain.ValidatePlan(context.Background(), plan)
		assert.NoError(t, err)
	})

	t.Run("invalid plan", func(t *testing.T) {
		mockProvider := &MockLLMProvider{}
		
		captain, err := NewCaptain("test-captain", config, mockProvider, logger)
		require.NoError(t, err)
		
		plan := &ExecutionPlan{
			ID:    "test-plan",
			Goal:  "test",
			Tasks: []Task{}, // Empty tasks makes it invalid
		}
		
		err = captain.ValidatePlan(context.Background(), plan)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plan validation failed")
	})
}

func TestCaptain_OptimizePlan(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultConfig()

	t.Run("successful optimization", func(t *testing.T) {
		optimizedResponse := `{
  "reasoning": "Optimized for parallel execution", 
  "tasks": [
    {
      "id": "task-1-opt",
      "title": "Optimized task",
      "description": "Optimized task",
      "command": "echo optimized",
      "dependencies": [],
      "estimated_duration": "3",
      "priority": 1
    }
  ],
  "estimated_duration": "3"
}`
		
		// Create mock provider that returns VALID for validation then optimization response
		callCount := 0
		mockProvider := &MockLLMProvider{
			CompletionFunc: func(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
				callCount++
				if callCount == 1 {
					// First call is validation
					return &CompletionResponse{Content: "VALID"}, nil
				}
				// Second call is optimization
				return &CompletionResponse{Content: optimizedResponse}, nil
			},
		}
		
		captain, err := NewCaptain("test-captain", config, mockProvider, logger)
		require.NoError(t, err)
		
		plan := &ExecutionPlan{
			ID:   "test-plan",
			Goal: "test",
			Tasks: []Task{
				{ID: "task-1", Title: "Original task"},
			},
		}
		
		optimized, err := captain.OptimizePlan(context.Background(), plan)
		require.NoError(t, err)
		require.NotNil(t, optimized)
		
		assert.Equal(t, "test", optimized.Goal)
		assert.Len(t, optimized.Tasks, 1)
		assert.Equal(t, "task-1-opt", optimized.Tasks[0].ID)
		assert.Equal(t, 3*time.Second, optimized.EstimatedDuration)
	})

	t.Run("optimization failure", func(t *testing.T) {
		mockProvider := &MockLLMProvider{
			CompletionError: errors.New("optimization failed"),
		}
		
		captain, err := NewCaptain("test-captain", config, mockProvider, logger)
		require.NoError(t, err)
		
		plan := &ExecutionPlan{
			ID:   "test-plan",
			Goal: "test",
			Tasks: []Task{
				{ID: "task-1", Title: "Test task"},
			},
		}
		
		optimized, err := captain.OptimizePlan(context.Background(), plan)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plan optimization failed")
		assert.Nil(t, optimized)
	})
}

func TestCaptain_ExecutePlan(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultConfig()
	mockProvider := &MockLLMProvider{
		CompletionResponse: &CompletionResponse{
			Content: "VALID - Plan is good",
		},
	}

	captain, err := NewCaptain("test-captain", config, mockProvider, logger)
	require.NoError(t, err)

	plan := &ExecutionPlan{
		ID:   "test-plan",
		Goal: "test execution",
		Tasks: []Task{
			{
				ID:      "task-1",
				Title:   "Test task",
				Command: "echo hello",
				EstimatedDuration: 5 * time.Second,
			},
		},
	}

	t.Run("dry run execution", func(t *testing.T) {
		results, err := captain.ExecutePlan(context.Background(), plan, true)
		require.NoError(t, err)
		require.Len(t, results, 1)
		
		assert.Equal(t, "task-1", results[0].TaskID)
		assert.True(t, results[0].Success)
		assert.Contains(t, results[0].Output, "DRY RUN")
		assert.Contains(t, results[0].Output, "Test task")
		assert.Equal(t, time.Duration(0), results[0].Duration) // Dry run has no duration
	})

	t.Run("normal execution", func(t *testing.T) {
		results, err := captain.ExecutePlan(context.Background(), plan, false)
		require.NoError(t, err)
		require.Len(t, results, 1)
		
		assert.Equal(t, "task-1", results[0].TaskID)
		assert.True(t, results[0].Success)
		assert.Contains(t, results[0].Output, "would be executed")
		assert.Equal(t, 5*time.Second, results[0].Duration)
	})

	t.Run("invalid plan execution", func(t *testing.T) {
		invalidPlan := &ExecutionPlan{
			ID:    "invalid-plan",
			Goal:  "test",
			Tasks: []Task{}, // Empty tasks makes it invalid
		}
		
		results, err := captain.ExecutePlan(context.Background(), invalidPlan, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot execute invalid plan")
		assert.Nil(t, results)
	})
}

func TestCaptain_Shutdown(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultConfig()
	mockProvider := &MockLLMProvider{}

	captain, err := NewCaptain("test-captain", config, mockProvider, logger)
	require.NoError(t, err)

	t.Run("normal shutdown", func(t *testing.T) {
		assert.False(t, captain.IsShutdown())
		
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		
		err := captain.Shutdown(ctx)
		assert.NoError(t, err)
		assert.True(t, captain.IsShutdown())
	})

	t.Run("already shutdown", func(t *testing.T) {
		// Captain is already shut down from previous test
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		
		err := captain.Shutdown(ctx)
		assert.NoError(t, err)
		assert.True(t, captain.IsShutdown())
	})
}

func TestCaptain_GetMethods(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultConfig()
	mockProvider := &MockLLMProvider{}

	captain, err := NewCaptain("test-captain", config, mockProvider, logger)
	require.NoError(t, err)

	assert.Equal(t, "test-captain", captain.GetID())
	assert.Equal(t, config, captain.GetConfig())
	assert.Equal(t, config.MaxConcurrentAgents, captain.GetMaxConcurrentAgents())
	assert.Equal(t, 0, captain.GetActiveAgentCount()) // No active agents initially
}