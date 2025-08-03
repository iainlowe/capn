package captain

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PlanningEngine handles goal decomposition and execution planning
type PlanningEngine struct {
	llmProvider LLMProvider
}

// NewPlanningEngine creates a new planning engine
func NewPlanningEngine(llmProvider LLMProvider) *PlanningEngine {
	return &PlanningEngine{
		llmProvider: llmProvider,
	}
}

// PlanResponse represents the structured response from the LLM for planning
type PlanResponse struct {
	Tasks             []TaskTemplate `json:"tasks"`
	Strategy          string         `json:"strategy"`
	EstimatedDuration string         `json:"estimated_duration"`
	Reasoning         string         `json:"reasoning,omitempty"`
}

// TaskTemplate represents a task template from LLM response
type TaskTemplate struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Priority     string   `json:"priority"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
}

// CreatePlan creates an execution plan from a goal using LLM-powered reasoning
func (pe *PlanningEngine) CreatePlan(ctx context.Context, goal string) (*ExecutionPlan, error) {
	if goal == "" {
		return nil, fmt.Errorf("goal cannot be empty")
	}

	// Build the planning prompt with chain-of-thought reasoning
	messages := pe.buildPlanningPrompt(goal)

	// Request completion from LLM
	req := CompletionRequest{
		Messages:    messages,
		MaxTokens:   2000,
		Temperature: 0.3, // Lower temperature for more consistent planning
	}

	resp, err := pe.llmProvider.GenerateCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	// Parse the LLM response
	planResp, err := pe.parsePlanResponse(resp.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse plan response: %w", err)
	}

	// Convert to execution plan
	plan, err := pe.convertToPlan(goal, planResp)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to execution plan: %w", err)
	}

	// Validate the generated plan
	if err := pe.ValidatePlan(plan); err != nil {
		return nil, fmt.Errorf("generated plan is invalid: %w", err)
	}

	return plan, nil
}

// ValidatePlan validates an execution plan for correctness
func (pe *PlanningEngine) ValidatePlan(plan *ExecutionPlan) error {
	if plan == nil {
		return fmt.Errorf("plan cannot be nil")
	}

	if plan.ID == "" {
		return fmt.Errorf("plan ID cannot be empty")
	}

	if plan.Goal == "" {
		return fmt.Errorf("goal cannot be empty")
	}

	if len(plan.Tasks) == 0 {
		return fmt.Errorf("plan must contain at least one task")
	}

	// Check for duplicate task IDs
	taskIDs := make(map[string]bool)
	for _, task := range plan.Tasks {
		if task.ID == "" {
			return fmt.Errorf("task ID cannot be empty")
		}
		if taskIDs[task.ID] {
			return fmt.Errorf("duplicate task ID: %s", task.ID)
		}
		taskIDs[task.ID] = true
	}

	// Validate dependencies
	for _, task := range plan.Tasks {
		for _, dep := range task.Dependencies {
			if !taskIDs[dep] {
				return fmt.Errorf("task %s depends on nonexistent task: %s", task.ID, dep)
			}
		}
	}

	// Check for circular dependencies
	if pe.hasCircularDependencies(plan.Tasks) {
		return fmt.Errorf("circular dependency detected")
	}

	return nil
}

// buildPlanningPrompt creates the prompt messages for planning
func (pe *PlanningEngine) buildPlanningPrompt(goal string) []Message {
	systemPrompt := `You are an expert AI planning agent specialized in task decomposition and execution planning. Your role is to analyze complex goals and create detailed, executable plans.

## Planning Principles:
1. Break down complex goals into atomic, executable tasks
2. Identify dependencies between tasks and order them logically
3. Assign appropriate priorities and task types
4. Consider resource requirements and time estimates
5. Use chain-of-thought reasoning to justify your decisions

## Task Types:
- "analysis": Information gathering, research, assessment
- "execution": Action execution, implementation, deployment
- "validation": Testing, verification, quality assurance
- "reporting": Documentation, summarization, communication

## Priority Levels:
- "critical": Must be completed immediately, blocks other work
- "high": Important, should be completed soon
- "medium": Standard priority, normal workflow
- "low": Nice to have, can be delayed

## Response Format:
Respond with a JSON object containing:
{
  "tasks": [
    {
      "id": "task-1",
      "type": "analysis|execution|validation|reporting",
      "priority": "critical|high|medium|low",
      "description": "Clear description of what needs to be done",
      "dependencies": ["task-id-1", "task-id-2"]
    }
  ],
  "strategy": "sequential|parallel|hybrid",
  "estimated_duration": "30m",
  "reasoning": "Brief explanation of the planning approach"
}

Think step by step and create a comprehensive plan.`

	userPrompt := fmt.Sprintf("Create an execution plan for the following goal:\n\n%s", goal)

	return []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
}

// parsePlanResponse parses the LLM response into a structured plan
func (pe *PlanningEngine) parsePlanResponse(content string) (*PlanResponse, error) {
	// Clean up the response - sometimes LLMs add extra formatting
	content = strings.TrimSpace(content)
	
	// Extract JSON if it's wrapped in code blocks
	if strings.Contains(content, "```json") {
		start := strings.Index(content, "```json") + 7
		end := strings.LastIndex(content, "```")
		if start < end {
			content = content[start:end]
		}
	} else if strings.Contains(content, "```") {
		start := strings.Index(content, "```") + 3
		end := strings.LastIndex(content, "```")
		if start < end {
			content = content[start:end]
		}
	}

	var planResp PlanResponse
	if err := json.Unmarshal([]byte(content), &planResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plan response: %w", err)
	}

	return &planResp, nil
}

// convertToPlan converts a plan response to an ExecutionPlan
func (pe *PlanningEngine) convertToPlan(goal string, planResp *PlanResponse) (*ExecutionPlan, error) {
	planID := uuid.New().String()

	// Convert tasks
	tasks := make([]Task, len(planResp.Tasks))
	for i, taskTemplate := range planResp.Tasks {
		taskType := TaskType(taskTemplate.Type)
		priority := Priority(taskTemplate.Priority)

		tasks[i] = Task{
			ID:           taskTemplate.ID,
			Type:         taskType,
			Priority:     priority,
			Dependencies: taskTemplate.Dependencies,
			Payload: map[string]any{
				"description": taskTemplate.Description,
			},
			Metadata: map[string]string{
				"generated_by": "planning_engine",
			},
		}
	}

	// Parse estimated duration
	duration, err := pe.parseEstimatedDuration(planResp.EstimatedDuration)
	if err != nil {
		// Default to 30 minutes if parsing fails
		duration = 30 * time.Minute
	}

	// Determine strategy type
	strategyType := StrategyType(planResp.Strategy)
	if strategyType != StrategySequential && strategyType != StrategyParallel && strategyType != StrategyHybrid {
		strategyType = StrategySequential // Default to sequential
	}

	plan := &ExecutionPlan{
		ID:    planID,
		Goal:  goal,
		Tasks: tasks,
		Timeline: ExecutionTimeline{
			EstimatedDuration: duration,
		},
		Resources: ResourceAllocation{
			MaxAgents: 5, // Default max agents
		},
		Strategy: ExecutionStrategy{
			Type:        strategyType,
			Description: planResp.Reasoning,
		},
	}

	return plan, nil
}

// parseEstimatedDuration parses duration strings like "30m", "1h", "2h30m"
func (pe *PlanningEngine) parseEstimatedDuration(durationStr string) (time.Duration, error) {
	if durationStr == "" {
		return 30 * time.Minute, nil
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		// Try to parse common formats manually
		durationStr = strings.ToLower(strings.TrimSpace(durationStr))
		
		if strings.HasSuffix(durationStr, "min") || strings.HasSuffix(durationStr, "minutes") {
			// Extract number
			numStr := strings.TrimSuffix(strings.TrimSuffix(durationStr, "minutes"), "min")
			numStr = strings.TrimSpace(numStr)
			
			var minutes int
			if _, err := fmt.Sscanf(numStr, "%d", &minutes); err == nil {
				return time.Duration(minutes) * time.Minute, nil
			}
		}
		
		return 0, err
	}

	return duration, nil
}

// hasCircularDependencies checks if there are circular dependencies in the task list
func (pe *PlanningEngine) hasCircularDependencies(tasks []Task) bool {
	// Build adjacency list
	graph := make(map[string][]string)
	for _, task := range tasks {
		graph[task.ID] = task.Dependencies
	}

	// Track visit states: 0 = unvisited, 1 = visiting, 2 = visited
	visited := make(map[string]int)

	// DFS to detect cycles
	var hasCycle func(string) bool
	hasCycle = func(taskID string) bool {
		if visited[taskID] == 1 {
			return true // Found a back edge (cycle)
		}
		if visited[taskID] == 2 {
			return false // Already processed
		}

		visited[taskID] = 1 // Mark as visiting
		for _, dep := range graph[taskID] {
			if hasCycle(dep) {
				return true
			}
		}
		visited[taskID] = 2 // Mark as visited
		return false
	}

	// Check each task
	for _, task := range tasks {
		if visited[task.ID] == 0 && hasCycle(task.ID) {
			return true
		}
	}

	return false
}