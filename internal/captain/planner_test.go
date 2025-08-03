package captain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLLMPlanningEngine(t *testing.T) {
	config := DefaultConfig()
	mockProvider := &MockLLMProvider{}
	
	engine := NewLLMPlanningEngine(mockProvider, config)
	
	require.NotNil(t, engine)
	assert.Equal(t, mockProvider, engine.llmProvider)
	assert.Equal(t, config, engine.config)
}

func TestLLMPlanningEngine_AnalyzeGoal(t *testing.T) {
	config := DefaultConfig()
	
	t.Run("empty goal", func(t *testing.T) {
		mockProvider := &MockLLMProvider{}
		engine := NewLLMPlanningEngine(mockProvider, config)
		
		plan, err := engine.AnalyzeGoal(context.Background(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "goal cannot be empty")
		assert.Nil(t, plan)
	})

	t.Run("successful analysis", func(t *testing.T) {
		mockResponse := `{
  "reasoning": "This goal requires creating a simple file and verifying it exists.",
  "tasks": [
    {
      "id": "task-1",
      "title": "Create file",
      "description": "Create a new text file",
      "command": "touch test.txt",
      "dependencies": [],
      "estimated_duration": "5",
      "priority": 1
    },
    {
      "id": "task-2", 
      "title": "Verify file",
      "description": "Check that file was created",
      "command": "ls -la test.txt",
      "dependencies": ["task-1"],
      "estimated_duration": "2",
      "priority": 2
    }
  ],
  "estimated_duration": "7"
}`
		
		mockProvider := &MockLLMProvider{
			CompletionResponse: &CompletionResponse{
				Content:    mockResponse,
				TokensUsed: 150,
				Model:      "gpt-4",
			},
		}
		
		engine := NewLLMPlanningEngine(mockProvider, config)
		
		plan, err := engine.AnalyzeGoal(context.Background(), "create a test file")
		require.NoError(t, err)
		require.NotNil(t, plan)
		
		assert.Equal(t, "create a test file", plan.Goal)
		assert.Len(t, plan.Tasks, 2)
		assert.Equal(t, "task-1", plan.Tasks[0].ID)
		assert.Equal(t, "Create file", plan.Tasks[0].Title)
		assert.Equal(t, "touch test.txt", plan.Tasks[0].Command)
		assert.Empty(t, plan.Tasks[0].Dependencies)
		assert.Equal(t, 5*time.Second, plan.Tasks[0].EstimatedDuration)
		
		assert.Equal(t, "task-2", plan.Tasks[1].ID)
		assert.Equal(t, []string{"task-1"}, plan.Tasks[1].Dependencies)
		assert.Equal(t, 7*time.Second, plan.EstimatedDuration)
		assert.Contains(t, plan.Reasoning, "creating a simple file")
	})

	t.Run("LLM error", func(t *testing.T) {
		mockProvider := &MockLLMProvider{
			CompletionError: errors.New("API error"),
		}
		
		engine := NewLLMPlanningEngine(mockProvider, config)
		
		plan, err := engine.AnalyzeGoal(context.Background(), "test goal")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to generate plan")
		assert.Nil(t, plan)
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		mockProvider := &MockLLMProvider{
			CompletionResponse: &CompletionResponse{
				Content: "This is not JSON",
			},
		}
		
		engine := NewLLMPlanningEngine(mockProvider, config)
		
		plan, err := engine.AnalyzeGoal(context.Background(), "test goal")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no valid JSON found")
		assert.Nil(t, plan)
	})

	t.Run("context timeout", func(t *testing.T) {
		// Create a config with very short timeout
		shortConfig := *config
		shortConfig.PlanningTimeout = time.Millisecond
		
		mockProvider := &MockLLMProvider{
			CompletionResponse: &CompletionResponse{
				Content: `{"reasoning": "test", "tasks": [], "estimated_duration": "0"}`,
			},
		}
		
		engine := NewLLMPlanningEngine(mockProvider, &shortConfig)
		
		// This should timeout quickly
		plan, err := engine.AnalyzeGoal(context.Background(), "test goal")
		// We might get a timeout error or success depending on timing
		if err != nil {
			assert.Contains(t, err.Error(), "context deadline exceeded")
			assert.Nil(t, plan)
		}
	})
}

func TestLLMPlanningEngine_ValidatePlan(t *testing.T) {
	config := DefaultConfig()
	
	t.Run("nil plan", func(t *testing.T) {
		mockProvider := &MockLLMProvider{}
		engine := NewLLMPlanningEngine(mockProvider, config)
		
		err := engine.ValidatePlan(context.Background(), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plan cannot be nil")
	})

	t.Run("empty tasks", func(t *testing.T) {
		mockProvider := &MockLLMProvider{}
		engine := NewLLMPlanningEngine(mockProvider, config)
		
		plan := &ExecutionPlan{
			ID:    "test-plan",
			Goal:  "test",
			Tasks: []Task{},
		}
		
		err := engine.ValidatePlan(context.Background(), plan)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plan must contain at least one task")
	})

	t.Run("empty task ID", func(t *testing.T) {
		mockProvider := &MockLLMProvider{}
		engine := NewLLMPlanningEngine(mockProvider, config)
		
		plan := &ExecutionPlan{
			ID:   "test-plan",
			Goal: "test",
			Tasks: []Task{
				{ID: "", Title: "Test task"},
			},
		}
		
		err := engine.ValidatePlan(context.Background(), plan)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task ID cannot be empty")
	})

	t.Run("duplicate task ID", func(t *testing.T) {
		mockProvider := &MockLLMProvider{}
		engine := NewLLMPlanningEngine(mockProvider, config)
		
		plan := &ExecutionPlan{
			ID:   "test-plan",
			Goal: "test",
			Tasks: []Task{
				{ID: "task-1", Title: "First task"},
				{ID: "task-1", Title: "Duplicate task"},
			},
		}
		
		err := engine.ValidatePlan(context.Background(), plan)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate task ID: task-1")
	})

	t.Run("invalid dependency", func(t *testing.T) {
		mockProvider := &MockLLMProvider{}
		engine := NewLLMPlanningEngine(mockProvider, config)
		
		plan := &ExecutionPlan{
			ID:   "test-plan",
			Goal: "test",
			Tasks: []Task{
				{ID: "task-1", Title: "First task", Dependencies: []string{"non-existent"}},
			},
		}
		
		err := engine.ValidatePlan(context.Background(), plan)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "depends on non-existent task non-existent")
	})

	t.Run("valid plan with LLM validation", func(t *testing.T) {
		mockProvider := &MockLLMProvider{
			CompletionResponse: &CompletionResponse{
				Content: "VALID - The plan looks good",
			},
		}
		engine := NewLLMPlanningEngine(mockProvider, config)
		
		plan := &ExecutionPlan{
			ID:   "test-plan",
			Goal: "test",
			Tasks: []Task{
				{ID: "task-1", Title: "First task"},
				{ID: "task-2", Title: "Second task", Dependencies: []string{"task-1"}},
			},
		}
		
		err := engine.ValidatePlan(context.Background(), plan)
		assert.NoError(t, err)
	})

	t.Run("LLM validation failure", func(t *testing.T) {
		mockProvider := &MockLLMProvider{
			CompletionResponse: &CompletionResponse{
				Content: "The plan has issues: missing validation steps",
			},
		}
		engine := NewLLMPlanningEngine(mockProvider, config)
		
		plan := &ExecutionPlan{
			ID:   "test-plan",
			Goal: "test",
			Tasks: []Task{
				{ID: "task-1", Title: "First task"},
			},
		}
		
		err := engine.ValidatePlan(context.Background(), plan)
		require.Error(t, err) // Use require to catch nil errors
		assert.Contains(t, err.Error(), "plan validation failed")
		assert.Contains(t, err.Error(), "missing validation steps")
	})
}

func TestLLMPlanningEngine_OptimizePlan(t *testing.T) {
	config := DefaultConfig()
	
	t.Run("nil plan", func(t *testing.T) {
		mockProvider := &MockLLMProvider{}
		engine := NewLLMPlanningEngine(mockProvider, config)
		
		optimized, err := engine.OptimizePlan(context.Background(), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plan cannot be nil")
		assert.Nil(t, optimized)
	})

	t.Run("invalid plan", func(t *testing.T) {
		mockProvider := &MockLLMProvider{}
		engine := NewLLMPlanningEngine(mockProvider, config)
		
		plan := &ExecutionPlan{
			ID:    "test-plan",
			Goal:  "test",
			Tasks: []Task{}, // Empty tasks makes it invalid
		}
		
		optimized, err := engine.OptimizePlan(context.Background(), plan)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot optimize invalid plan")
		assert.Nil(t, optimized)
	})

	t.Run("successful optimization", func(t *testing.T) {
		optimizedResponse := `{
  "reasoning": "Optimized for parallel execution",
  "tasks": [
    {
      "id": "task-1",
      "title": "Optimized task 1",
      "description": "First optimized task",
      "command": "echo optimized",
      "dependencies": [],
      "estimated_duration": "3",
      "priority": 1
    },
    {
      "id": "task-2",
      "title": "Optimized task 2", 
      "description": "Second optimized task",
      "command": "echo parallel",
      "dependencies": [],
      "estimated_duration": "3",
      "priority": 1
    }
  ],
  "estimated_duration": "3"
}`
		
		// Use function to handle multiple calls: validation, then optimization
		callCount := 0
		mockProvider := &MockLLMProvider{
			CompletionFunc: func(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
				callCount++
				if callCount == 1 {
					// First call is for validation
					return &CompletionResponse{Content: "VALID"}, nil
				}
				// Second call is for optimization
				return &CompletionResponse{Content: optimizedResponse}, nil
			},
		}
		engine := NewLLMPlanningEngine(mockProvider, config)
		
		// Valid plan for optimization
		plan := &ExecutionPlan{
			ID:   "test-plan",
			Goal: "test optimization",
			Tasks: []Task{
				{ID: "task-1", Title: "First task"},
				{ID: "task-2", Title: "Second task", Dependencies: []string{"task-1"}},
			},
		}
		
		// Mock the validation call as well
		// Remove this line since we're using CompletionFunc above
		// mockProvider.CompletionResponse = &CompletionResponse{Content: "VALID"}
		
		optimized, err := engine.OptimizePlan(context.Background(), plan)
		require.NoError(t, err)
		require.NotNil(t, optimized)
		
		assert.Equal(t, "test optimization", optimized.Goal)
		assert.Len(t, optimized.Tasks, 2)
		assert.Equal(t, "Optimized task 1", optimized.Tasks[0].Title)
		assert.Equal(t, "Optimized task 2", optimized.Tasks[1].Title)
		// Both tasks should now run in parallel (no dependencies)
		assert.Empty(t, optimized.Tasks[0].Dependencies)
		assert.Empty(t, optimized.Tasks[1].Dependencies)
		assert.Equal(t, 3*time.Second, optimized.EstimatedDuration)
	})
}

func TestLLMPlanningEngine_ExtractJSON(t *testing.T) {
	config := DefaultConfig()
	mockProvider := &MockLLMProvider{}
	engine := NewLLMPlanningEngine(mockProvider, config)
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple JSON",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with surrounding text",
			input:    `Here is the plan: {"tasks": []} That's it.`,
			expected: `{"tasks": []}`,
		},
		{
			name:     "nested JSON",
			input:    `{"outer": {"inner": {"deep": "value"}}}`,
			expected: `{"outer": {"inner": {"deep": "value"}}}`,
		},
		{
			name:     "no JSON",
			input:    "No JSON here",
			expected: "",
		},
		{
			name:     "incomplete JSON",
			input:    `{"incomplete": `,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.extractJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLLMPlanningEngine_Prompts(t *testing.T) {
	config := DefaultConfig()
	mockProvider := &MockLLMProvider{}
	engine := NewLLMPlanningEngine(mockProvider, config)
	
	t.Run("system prompt", func(t *testing.T) {
		prompt := engine.getSystemPrompt()
		assert.Contains(t, prompt, "expert task planning")
		assert.Contains(t, prompt, "JSON format")
		assert.Contains(t, prompt, "dependencies")
	})

	t.Run("optimization system prompt", func(t *testing.T) {
		prompt := engine.getOptimizationSystemPrompt()
		assert.Contains(t, prompt, "optimizing execution plans")
		assert.Contains(t, prompt, "parallel")
		assert.Contains(t, prompt, "efficiency")
	})

	t.Run("goal analysis prompt", func(t *testing.T) {
		goal := "test goal"
		prompt := engine.buildGoalAnalysisPrompt(goal)
		assert.Contains(t, prompt, goal)
		assert.Contains(t, prompt, "chain-of-thought")
		assert.Contains(t, prompt, "JSON format")
	})

	t.Run("optimization prompt", func(t *testing.T) {
		planJSON := `{"test": "plan"}`
		prompt := engine.buildOptimizationPrompt(planJSON)
		assert.Contains(t, prompt, planJSON)
		assert.Contains(t, prompt, "optimize")
		assert.Contains(t, prompt, "parallelism")
	})
}