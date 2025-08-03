package captain

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LLMPlanningEngine implements the PlanningEngine interface using an LLM provider
type LLMPlanningEngine struct {
	llmProvider LLMProvider
	config      *Config
}

// NewLLMPlanningEngine creates a new LLM-powered planning engine
func NewLLMPlanningEngine(llmProvider LLMProvider, config *Config) *LLMPlanningEngine {
	return &LLMPlanningEngine{
		llmProvider: llmProvider,
		config:      config,
	}
}

// AnalyzeGoal analyzes a goal and creates an execution plan using chain-of-thought reasoning
func (p *LLMPlanningEngine) AnalyzeGoal(ctx context.Context, goal string) (*ExecutionPlan, error) {
	if goal == "" {
		return nil, fmt.Errorf("goal cannot be empty")
	}

	// Create a timeout context for planning
	planCtx, cancel := context.WithTimeout(ctx, p.config.PlanningTimeout)
	defer cancel()

	// Build the chain-of-thought prompt
	prompt := p.buildGoalAnalysisPrompt(goal)
	
	req := CompletionRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: p.getSystemPrompt(),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   p.config.MaxTokens,
		Temperature: p.config.Temperature,
		Model:       p.config.Model,
	}

	resp, err := p.llmProvider.GenerateCompletion(planCtx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	// Parse the LLM response to extract the execution plan
	plan, err := p.parseExecutionPlan(resp.Content, goal)
	if err != nil {
		return nil, fmt.Errorf("failed to parse execution plan: %w", err)
	}

	return plan, nil
}

// ValidatePlan validates an execution plan for feasibility and correctness
func (p *LLMPlanningEngine) ValidatePlan(ctx context.Context, plan *ExecutionPlan) error {
	if plan == nil {
		return fmt.Errorf("plan cannot be nil")
	}

	// Basic validation
	if len(plan.Tasks) == 0 {
		return fmt.Errorf("plan must contain at least one task")
	}

	// Validate task dependencies
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

	// Check dependency validity
	for _, task := range plan.Tasks {
		for _, depID := range task.Dependencies {
			if !taskIDs[depID] {
				return fmt.Errorf("task %s depends on non-existent task %s", task.ID, depID)
			}
		}
	}

	// Use LLM for deeper validation
	return p.llmValidatePlan(ctx, plan)
}

// OptimizePlan optimizes an execution plan for better performance
func (p *LLMPlanningEngine) OptimizePlan(ctx context.Context, plan *ExecutionPlan) (*ExecutionPlan, error) {
	if plan == nil {
		return nil, fmt.Errorf("plan cannot be nil")
	}

	// First validate the plan
	if err := p.ValidatePlan(ctx, plan); err != nil {
		return nil, fmt.Errorf("cannot optimize invalid plan: %w", err)
	}

	// Create optimization prompt
	planJSON, err := plan.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize plan: %w", err)
	}

	prompt := p.buildOptimizationPrompt(planJSON)
	
	req := CompletionRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: p.getOptimizationSystemPrompt(),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   p.config.MaxTokens,
		Temperature: p.config.Temperature * 0.8, // Lower temperature for optimization
		Model:       p.config.Model,
	}

	resp, err := p.llmProvider.GenerateCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to optimize plan: %w", err)
	}

	// Parse the optimized plan
	optimizedPlan, err := p.parseExecutionPlan(resp.Content, plan.Goal)
	if err != nil {
		return nil, fmt.Errorf("failed to parse optimized plan: %w", err)
	}

	return optimizedPlan, nil
}

// getSystemPrompt returns the system prompt for goal analysis
func (p *LLMPlanningEngine) getSystemPrompt() string {
	return `You are an expert task planning AI assistant. Your role is to analyze goals and create detailed, actionable execution plans.

Key principles:
1. Break down complex goals into manageable, specific tasks
2. Identify task dependencies and optimal execution order  
3. Provide realistic time estimates
4. Include necessary commands and validation steps
5. Consider potential failure points and mitigation strategies

Response format: Provide your analysis in JSON format with the following structure:
{
  "reasoning": "Your step-by-step analysis and reasoning",
  "tasks": [
    {
      "id": "unique-task-id",
      "title": "Task title",
      "description": "Detailed description",
      "command": "shell command if applicable",
      "dependencies": ["list-of-task-ids"],
      "estimated_duration": "duration in seconds",
      "priority": 1
    }
  ],
  "estimated_duration": "total duration in seconds"
}`
}

// getOptimizationSystemPrompt returns the system prompt for plan optimization
func (p *LLMPlanningEngine) getOptimizationSystemPrompt() string {
	return `You are an expert at optimizing execution plans for maximum efficiency and parallelism.

Focus on:
1. Identifying tasks that can run in parallel
2. Optimizing task order to minimize total execution time
3. Reducing unnecessary dependencies
4. Improving resource utilization
5. Adding checkpoints and validation steps

Maintain the same JSON format while improving the plan structure.`
}

// buildGoalAnalysisPrompt creates the prompt for goal analysis
func (p *LLMPlanningEngine) buildGoalAnalysisPrompt(goal string) string {
	return fmt.Sprintf(`Please analyze the following goal and create a detailed execution plan:

Goal: %s

Use chain-of-thought reasoning to:
1. Understand what the goal is trying to achieve
2. Break it down into specific, actionable tasks
3. Identify dependencies between tasks
4. Estimate realistic durations
5. Consider potential issues and how to handle them

Provide your analysis and the execution plan in the specified JSON format.`, goal)
}

// buildOptimizationPrompt creates the prompt for plan optimization
func (p *LLMPlanningEngine) buildOptimizationPrompt(planJSON string) string {
	return fmt.Sprintf(`Please optimize the following execution plan for better performance and efficiency:

Current Plan:
%s

Analyze the plan and provide an optimized version that:
1. Maximizes parallelism where possible
2. Reduces total execution time
3. Improves task organization
4. Maintains all necessary dependencies
5. Adds any missing validation or checkpoint tasks

Return the optimized plan in the same JSON format.`, planJSON)
}

// parseExecutionPlan parses the LLM response to extract an execution plan
func (p *LLMPlanningEngine) parseExecutionPlan(content, goal string) (*ExecutionPlan, error) {
	// Try to extract JSON from the response
	jsonStr := p.extractJSON(content)
	if jsonStr == "" {
		return nil, fmt.Errorf("no valid JSON found in response")
	}

	// Parse the JSON structure
	var planData struct {
		Reasoning         string `json:"reasoning"`
		Tasks             []struct {
			ID                string   `json:"id"`
			Title             string   `json:"title"`
			Description       string   `json:"description"`
			Command           string   `json:"command"`
			Dependencies      []string `json:"dependencies"`
			EstimatedDuration string   `json:"estimated_duration"`
			Priority          int      `json:"priority"`
		} `json:"tasks"`
		EstimatedDuration string `json:"estimated_duration"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &planData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to ExecutionPlan
	plan := &ExecutionPlan{
		ID:          uuid.New().String(),
		Goal:        goal,
		Description: fmt.Sprintf("Execution plan for: %s", goal),
		CreatedAt:   time.Now().UTC(),
		Reasoning:   planData.Reasoning,
		Tasks:       make([]Task, len(planData.Tasks)),
	}

	// Parse estimated duration
	if planData.EstimatedDuration != "" {
		if duration, err := time.ParseDuration(planData.EstimatedDuration + "s"); err == nil {
			plan.EstimatedDuration = duration
		}
	}

	// Convert tasks
	for i, taskData := range planData.Tasks {
		task := Task{
			ID:           taskData.ID,
			Title:        taskData.Title,
			Description:  taskData.Description,
			Command:      taskData.Command,
			Dependencies: taskData.Dependencies,
			Priority:     taskData.Priority,
			Status:       TaskStatusPending,
		}

		// Parse estimated duration
		if taskData.EstimatedDuration != "" {
			if duration, err := time.ParseDuration(taskData.EstimatedDuration + "s"); err == nil {
				task.EstimatedDuration = duration
			}
		}

		plan.Tasks[i] = task
	}

	return plan, nil
}

// extractJSON extracts JSON content from a potentially verbose response
func (p *LLMPlanningEngine) extractJSON(content string) string {
	// Look for JSON block markers
	start := strings.Index(content, "{")
	if start == -1 {
		return ""
	}

	// Find the matching closing brace
	braceCount := 0
	end := -1
	for i := start; i < len(content); i++ {
		if content[i] == '{' {
			braceCount++
		} else if content[i] == '}' {
			braceCount--
			if braceCount == 0 {
				end = i + 1
				break
			}
		}
	}

	if end == -1 {
		return ""
	}

	return content[start:end]
}

// llmValidatePlan uses the LLM to validate a plan
func (p *LLMPlanningEngine) llmValidatePlan(ctx context.Context, plan *ExecutionPlan) error {
	planJSON, err := plan.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize plan for validation: %w", err)
	}

	prompt := fmt.Sprintf(`Please validate the following execution plan and identify any issues:

%s

Check for:
1. Logical consistency
2. Feasibility of tasks
3. Proper dependency management
4. Realistic time estimates
5. Missing steps or edge cases

Respond with "VALID" if the plan is good, or explain specific issues if not.`, planJSON)

	req := CompletionRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a plan validation expert. Analyze execution plans for correctness and feasibility.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   1000,
		Temperature: 0.3, // Lower temperature for validation
		Model:       p.config.Model,
	}

	resp, err := p.llmProvider.GenerateCompletion(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to validate plan with LLM: %w", err)
	}

	// Check if the response indicates the plan is valid
	// Look for "VALID" as a standalone word, not as part of other words like "validation"
	upperContent := strings.ToUpper(resp.Content)
	isValid := false
	
	// Split into words and check for exact match
	words := strings.Fields(upperContent)
	for _, word := range words {
		// Remove common punctuation
		cleanWord := strings.Trim(word, ".,!?:;-")
		if cleanWord == "VALID" {
			isValid = true
			break
		}
	}
	
	if !isValid {
		return fmt.Errorf("plan validation failed: %s", resp.Content)
	}

	return nil
}