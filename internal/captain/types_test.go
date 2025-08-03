package captain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask_JSONSerialization(t *testing.T) {
	task := &Task{
		ID:          "task-1",
		Title:       "Test Task",
		Description: "A test task for validation",
		Command:     "echo hello",
		Dependencies: []string{"task-0"},
		EstimatedDuration: 30 * time.Second,
		Priority:    1,
		Status:      TaskStatusPending,
		Metadata:    map[string]interface{}{"env": "test"},
	}

	// Test ToJSON
	jsonStr, err := task.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, jsonStr, "task-1")
	assert.Contains(t, jsonStr, "Test Task")

	// Test FromJSON
	newTask := &Task{}
	err = newTask.FromJSON(jsonStr)
	require.NoError(t, err)
	assert.Equal(t, task.ID, newTask.ID)
	assert.Equal(t, task.Title, newTask.Title)
	assert.Equal(t, task.Description, newTask.Description)
	assert.Equal(t, task.Command, newTask.Command)
	assert.Equal(t, task.Dependencies, newTask.Dependencies)
	assert.Equal(t, task.Priority, newTask.Priority)
	assert.Equal(t, task.Status, newTask.Status)
}

func TestTask_FromJSON_InvalidJSON(t *testing.T) {
	task := &Task{}
	err := task.FromJSON("invalid json")
	assert.Error(t, err)
}

func TestExecutionPlan_JSONSerialization(t *testing.T) {
	createdAt := time.Now().UTC()
	plan := &ExecutionPlan{
		ID:          "plan-1",
		Goal:        "Test goal",
		Description: "A test execution plan",
		Tasks: []Task{
			{
				ID:     "task-1",
				Title:  "First task",
				Status: TaskStatusPending,
			},
			{
				ID:     "task-2",
				Title:  "Second task",
				Status: TaskStatusPending,
				Dependencies: []string{"task-1"},
			},
		},
		CreatedAt:         createdAt,
		EstimatedDuration: 5 * time.Minute,
		Reasoning:         "This is a test plan",
		Metadata:          map[string]interface{}{"priority": "high"},
	}

	// Test ToJSON
	jsonStr, err := plan.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, jsonStr, "plan-1")
	assert.Contains(t, jsonStr, "Test goal")
	assert.Contains(t, jsonStr, "task-1")
	assert.Contains(t, jsonStr, "task-2")

	// Test FromJSON
	newPlan := &ExecutionPlan{}
	err = newPlan.FromJSON(jsonStr)
	require.NoError(t, err)
	assert.Equal(t, plan.ID, newPlan.ID)
	assert.Equal(t, plan.Goal, newPlan.Goal)
	assert.Equal(t, plan.Description, newPlan.Description)
	assert.Len(t, newPlan.Tasks, 2)
	assert.Equal(t, plan.Tasks[0].ID, newPlan.Tasks[0].ID)
	assert.Equal(t, plan.Tasks[1].Dependencies, newPlan.Tasks[1].Dependencies)
	assert.Equal(t, plan.Reasoning, newPlan.Reasoning)
}

func TestExecutionPlan_FromJSON_InvalidJSON(t *testing.T) {
	plan := &ExecutionPlan{}
	err := plan.FromJSON("invalid json")
	assert.Error(t, err)
}

func TestTaskStatus_Constants(t *testing.T) {
	assert.Equal(t, TaskStatus("pending"), TaskStatusPending)
	assert.Equal(t, TaskStatus("running"), TaskStatusRunning)
	assert.Equal(t, TaskStatus("completed"), TaskStatusCompleted)
	assert.Equal(t, TaskStatus("failed"), TaskStatusFailed)
	assert.Equal(t, TaskStatus("skipped"), TaskStatusSkipped)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	require.NotNil(t, config)
	
	assert.Equal(t, "gpt-4", config.Model)
	assert.Equal(t, 2000, config.MaxTokens)
	assert.Equal(t, float32(0.7), config.Temperature)
	assert.Equal(t, 5, config.MaxConcurrentAgents)
	assert.Equal(t, 30*time.Second, config.PlanningTimeout)
	assert.Equal(t, 3, config.RetryAttempts)
	assert.Equal(t, time.Second, config.RetryDelay)
}

func TestCompletionRequest_JSONSerialization(t *testing.T) {
	req := CompletionRequest{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
		},
		MaxTokens:   1000,
		Temperature: 0.8,
		Model:       "gpt-3.5-turbo",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var newReq CompletionRequest
	err = json.Unmarshal(data, &newReq)
	require.NoError(t, err)

	assert.Equal(t, req.Messages, newReq.Messages)
	assert.Equal(t, req.MaxTokens, newReq.MaxTokens)
	assert.Equal(t, req.Temperature, newReq.Temperature)
	assert.Equal(t, req.Model, newReq.Model)
}

func TestCompletionResponse_JSONSerialization(t *testing.T) {
	resp := CompletionResponse{
		Content:    "Test response",
		TokensUsed: 150,
		Model:      "gpt-4",
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var newResp CompletionResponse
	err = json.Unmarshal(data, &newResp)
	require.NoError(t, err)

	assert.Equal(t, resp.Content, newResp.Content)
	assert.Equal(t, resp.TokensUsed, newResp.TokensUsed)
	assert.Equal(t, resp.Model, newResp.Model)
}

func TestResult_JSONSerialization(t *testing.T) {
	timestamp := time.Now().UTC()
	result := Result{
		TaskID:    "task-1",
		Success:   true,
		Output:    "Command executed successfully",
		Duration:  2 * time.Second,
		Timestamp: timestamp,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var newResult Result
	err = json.Unmarshal(data, &newResult)
	require.NoError(t, err)

	assert.Equal(t, result.TaskID, newResult.TaskID)
	assert.Equal(t, result.Success, newResult.Success)
	assert.Equal(t, result.Output, newResult.Output)
	assert.Equal(t, result.Error, newResult.Error)
	assert.Equal(t, result.Duration, newResult.Duration)
	// Time comparison with truncation to handle precision differences
	assert.True(t, result.Timestamp.Truncate(time.Second).Equal(newResult.Timestamp.Truncate(time.Second)))
}

func TestResult_WithError(t *testing.T) {
	result := Result{
		TaskID:    "task-1",
		Success:   false,
		Output:    "Failed output",
		Error:     "Command failed with exit code 1",
		Duration:  1 * time.Second,
		Timestamp: time.Now().UTC(),
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var newResult Result
	err = json.Unmarshal(data, &newResult)
	require.NoError(t, err)

	assert.Equal(t, result.TaskID, newResult.TaskID)
	assert.Equal(t, result.Success, newResult.Success)
	assert.Equal(t, result.Error, newResult.Error)
}