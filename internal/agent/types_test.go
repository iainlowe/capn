package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAgentType_String(t *testing.T) {
	tests := []struct {
		name     string
		agentType AgentType
		expected string
	}{
		{"Captain type", Captain, "Captain"},
		{"FileAgent type", FileAgent, "FileAgent"},
		{"NetworkAgent type", NetworkAgent, "NetworkAgent"},
		{"ResearchAgent type", ResearchAgent, "ResearchAgent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.agentType.String())
		})
	}
}

func TestAgentStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   AgentStatus
		expected string
	}{
		{"Idle status", Idle, "Idle"},
		{"Running status", Running, "Running"},
		{"Stopped status", Stopped, "Stopped"},
		{"Error status", Error, "Error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestHealthStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		health   HealthStatus
		expected string
	}{
		{"Healthy status", Healthy, "Healthy"},
		{"Unhealthy status", Unhealthy, "Unhealthy"},
		{"Unknown status", Unknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.health.String())
		})
	}
}

func TestMessage_Validation(t *testing.T) {
	tests := []struct {
		name    string
		message Message
		isValid bool
	}{
		{
			name: "valid message",
			message: Message{
				ID:        "msg-001",
				From:      "Captain",
				To:        "FileAgent-1",
				Content:   "Please analyze files",
				Timestamp: time.Now(),
				Type:      TaskMessage,
			},
			isValid: true,
		},
		{
			name: "empty ID",
			message: Message{
				From:      "Captain",
				To:        "FileAgent-1",
				Content:   "Please analyze files",
				Timestamp: time.Now(),
				Type:      TaskMessage,
			},
			isValid: false,
		},
		{
			name: "empty From",
			message: Message{
				ID:        "msg-001",
				To:        "FileAgent-1",
				Content:   "Please analyze files",
				Timestamp: time.Now(),
				Type:      TaskMessage,
			},
			isValid: false,
		},
		{
			name: "empty To",
			message: Message{
				ID:        "msg-001",
				From:      "Captain",
				Content:   "Please analyze files",
				Timestamp: time.Now(),
				Type:      TaskMessage,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.Validate()
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestTask_Validation(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		isValid bool
	}{
		{
			name: "valid task",
			task: Task{
				ID:          "task-001",
				Description: "Analyze Go files",
				Parameters:  map[string]interface{}{"path": "./src"},
			},
			isValid: true,
		},
		{
			name: "empty ID",
			task: Task{
				Description: "Analyze Go files",
				Parameters:  map[string]interface{}{"path": "./src"},
			},
			isValid: false,
		},
		{
			name: "empty description",
			task: Task{
				ID:         "task-001",
				Parameters: map[string]interface{}{"path": "./src"},
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestResult_Success(t *testing.T) {
	result := Result{
		TaskID:  "task-001",
		Success: true,
		Data:    "Analysis complete",
	}

	assert.True(t, result.IsSuccess())
	assert.False(t, result.IsError())
	assert.NoError(t, result.GetError())
}

func TestResult_Error(t *testing.T) {
	result := Result{
		TaskID:  "task-001",
		Success: false,
		Error:   "File not found",
	}

	assert.False(t, result.IsSuccess())
	assert.True(t, result.IsError())
	assert.Error(t, result.GetError())
	assert.Equal(t, "File not found", result.GetError().Error())
}