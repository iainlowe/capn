package captain

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewPlanningEngine(t *testing.T) {
	mockLLM := &MockLLMProvider{}
	engine := NewPlanningEngine(mockLLM)
	
	assert.NotNil(t, engine)
	assert.Equal(t, mockLLM, engine.llmProvider)
}

func TestPlanningEngine_CreatePlan(t *testing.T) {
	tests := []struct {
		name        string
		goal        string
		mockSetup   func(*MockLLMProvider)
		wantErr     bool
		expectTasks int
	}{
		{
			name: "simple goal decomposition",
			goal: "analyze code quality",
			mockSetup: func(m *MockLLMProvider) {
				response := &CompletionResponse{
					Content: `{
						"tasks": [
							{
								"id": "task-1",
								"type": "analysis",
								"priority": "high",
								"description": "Run static code analysis",
								"dependencies": []
							},
							{
								"id": "task-2", 
								"type": "reporting",
								"priority": "medium",
								"description": "Generate quality report",
								"dependencies": ["task-1"]
							}
						],
						"strategy": "sequential",
						"estimated_duration": "10m"
					}`,
					TokensUsed: 150,
					Model:      "gpt-3.5-turbo",
				}
				m.On("GenerateCompletion", mock.Anything, mock.MatchedBy(func(req CompletionRequest) bool {
					return len(req.Messages) > 0 && 
						   strings.Contains(req.Messages[len(req.Messages)-1].Content, "analyze code quality")
				})).Return(response, nil)
			},
			wantErr:     false,
			expectTasks: 2,
		},
		{
			name: "complex goal with chain of thought",
			goal: "set up CI/CD pipeline with testing and deployment",
			mockSetup: func(m *MockLLMProvider) {
				response := &CompletionResponse{
					Content: `{
						"tasks": [
							{
								"id": "task-1",
								"type": "analysis",
								"priority": "high",
								"description": "Analyze existing codebase structure",
								"dependencies": []
							},
							{
								"id": "task-2",
								"type": "execution", 
								"priority": "high",
								"description": "Set up GitHub Actions workflow",
								"dependencies": ["task-1"]
							},
							{
								"id": "task-3",
								"type": "execution",
								"priority": "medium",
								"description": "Configure test runners",
								"dependencies": ["task-2"]
							},
							{
								"id": "task-4",
								"type": "validation",
								"priority": "high",
								"description": "Test the pipeline",
								"dependencies": ["task-3"]
							}
						],
						"strategy": "sequential",
						"estimated_duration": "45m"
					}`,
					TokensUsed: 300,
					Model:      "gpt-3.5-turbo",
				}
				m.On("GenerateCompletion", mock.Anything, mock.MatchedBy(func(req CompletionRequest) bool {
					return len(req.Messages) > 0
				})).Return(response, nil)
			},
			wantErr:     false,
			expectTasks: 4,
		},
		{
			name: "LLM API error",
			goal: "test goal",
			mockSetup: func(m *MockLLMProvider) {
				m.On("GenerateCompletion", mock.Anything, mock.Anything).Return(
					(*CompletionResponse)(nil), errors.New("API rate limit exceeded"))
			},
			wantErr:     true,
			expectTasks: 0,
		},
		{
			name: "invalid JSON response",
			goal: "test goal",
			mockSetup: func(m *MockLLMProvider) {
				response := &CompletionResponse{
					Content:    "invalid json content",
					TokensUsed: 50,
					Model:      "gpt-3.5-turbo",
				}
				m.On("GenerateCompletion", mock.Anything, mock.Anything).Return(response, nil)
			},
			wantErr:     true,
			expectTasks: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := &MockLLMProvider{}
			tt.mockSetup(mockLLM)
			
			engine := NewPlanningEngine(mockLLM)
			ctx := context.Background()
			
			plan, err := engine.CreatePlan(ctx, tt.goal)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, plan)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, plan)
				assert.NotEmpty(t, plan.ID)
				assert.Equal(t, tt.goal, plan.Goal)
				assert.Len(t, plan.Tasks, tt.expectTasks)
				assert.NotEmpty(t, plan.Strategy.Type)
				
				// Verify task dependencies are valid
				taskIDs := make(map[string]bool)
				for _, task := range plan.Tasks {
					taskIDs[task.ID] = true
				}
				
				for _, task := range plan.Tasks {
					for _, dep := range task.Dependencies {
						assert.True(t, taskIDs[dep], 
							"Task %s has invalid dependency %s", task.ID, dep)
					}
				}
			}
			
			mockLLM.AssertExpectations(t)
		})
	}
}

func TestPlanningEngine_ValidatePlan(t *testing.T) {
	engine := NewPlanningEngine(&MockLLMProvider{})
	
	tests := []struct {
		name    string
		plan    *ExecutionPlan
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid plan",
			plan: &ExecutionPlan{
				ID:   "plan-1",
				Goal: "test goal",
				Tasks: []Task{
					{ID: "task-1", Type: TaskTypeAnalysis, Priority: PriorityHigh},
					{ID: "task-2", Type: TaskTypeExecution, Priority: PriorityMedium, Dependencies: []string{"task-1"}},
				},
				Strategy: ExecutionStrategy{Type: StrategySequential},
			},
			wantErr: false,
		},
		{
			name: "empty plan ID",
			plan: &ExecutionPlan{
				Goal: "test goal",
				Tasks: []Task{
					{ID: "task-1", Type: TaskTypeAnalysis, Priority: PriorityHigh},
				},
			},
			wantErr: true,
			errMsg:  "plan ID cannot be empty",
		},
		{
			name: "empty goal",
			plan: &ExecutionPlan{
				ID: "plan-1",
				Tasks: []Task{
					{ID: "task-1", Type: TaskTypeAnalysis, Priority: PriorityHigh},
				},
			},
			wantErr: true,
			errMsg:  "goal cannot be empty",
		},
		{
			name: "no tasks",
			plan: &ExecutionPlan{
				ID:    "plan-1",
				Goal:  "test goal",
				Tasks: []Task{},
			},
			wantErr: true,
			errMsg:  "plan must contain at least one task",
		},
		{
			name: "duplicate task IDs",
			plan: &ExecutionPlan{
				ID:   "plan-1",
				Goal: "test goal",
				Tasks: []Task{
					{ID: "task-1", Type: TaskTypeAnalysis, Priority: PriorityHigh},
					{ID: "task-1", Type: TaskTypeExecution, Priority: PriorityMedium},
				},
			},
			wantErr: true,
			errMsg:  "duplicate task ID: task-1",
		},
		{
			name: "invalid dependency",
			plan: &ExecutionPlan{
				ID:   "plan-1",
				Goal: "test goal",
				Tasks: []Task{
					{ID: "task-1", Type: TaskTypeAnalysis, Priority: PriorityHigh, Dependencies: []string{"nonexistent"}},
				},
			},
			wantErr: true,
			errMsg:  "task task-1 depends on nonexistent task: nonexistent",
		},
		{
			name: "circular dependency",
			plan: &ExecutionPlan{
				ID:   "plan-1",
				Goal: "test goal",
				Tasks: []Task{
					{ID: "task-1", Type: TaskTypeAnalysis, Priority: PriorityHigh, Dependencies: []string{"task-2"}},
					{ID: "task-2", Type: TaskTypeExecution, Priority: PriorityMedium, Dependencies: []string{"task-1"}},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidatePlan(tt.plan)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanningEngine_buildPlanningPrompt(t *testing.T) {
	engine := NewPlanningEngine(&MockLLMProvider{})
	
	goal := "analyze code quality"
	messages := engine.buildPlanningPrompt(goal)
	
	require.Len(t, messages, 2)
	
	// Check system message
	assert.Equal(t, "system", messages[0].Role)
	assert.Contains(t, messages[0].Content, "You are an expert")
	assert.Contains(t, messages[0].Content, "task decomposition")
	
	// Check user message
	assert.Equal(t, "user", messages[1].Role)
	assert.Contains(t, messages[1].Content, goal)
}