package agents

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
		{"Captain type", AgentTypeCaptain, "captain"},
		{"File type", AgentTypeFile, "file"},
		{"Network type", AgentTypeNetwork, "network"},
		{"Research type", AgentTypeResearch, "research"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.agentType))
		})
	}
}

func TestAgentStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   AgentStatus
		expected string
	}{
		{"Idle status", AgentStatusIdle, "idle"},
		{"Busy status", AgentStatusBusy, "busy"},
		{"Stopped status", AgentStatusStopped, "stopped"},
		{"Error status", AgentStatusError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestMessage_Validation(t *testing.T) {
	tests := []struct {
		name      string
		message   Message
		wantError bool
	}{
		{
			name: "valid message",
			message: Message{
				ID:        "msg-1",
				From:      "agent-1",
				To:        "agent-2",
				Content:   "Hello world",
				Type:      MessageTypeText,
				Timestamp: time.Now(),
			},
			wantError: false,
		},
		{
			name: "empty ID",
			message: Message{
				From:      "agent-1",
				To:        "agent-2",
				Content:   "Hello world",
				Type:      MessageTypeText,
				Timestamp: time.Now(),
			},
			wantError: true,
		},
		{
			name: "empty from",
			message: Message{
				ID:        "msg-1",
				To:        "agent-2",
				Content:   "Hello world",
				Type:      MessageTypeText,
				Timestamp: time.Now(),
			},
			wantError: true,
		},
		{
			name: "empty to",
			message: Message{
				ID:        "msg-1",
				From:      "agent-1",
				Content:   "Hello world",
				Type:      MessageTypeText,
				Timestamp: time.Now(),
			},
			wantError: true,
		},
		{
			name: "empty content",
			message: Message{
				ID:        "msg-1",
				From:      "agent-1",
				To:        "agent-2",
				Type:      MessageTypeText,
				Timestamp: time.Now(),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.Validate()
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTask_Validation(t *testing.T) {
	tests := []struct {
		name      string
		task      Task
		wantError bool
	}{
		{
			name: "valid task",
			task: Task{
				ID:          "task-1",
				Type:        "analysis",
				Description: "Analyze files",
				Priority:    PriorityMedium,
				Data:        map[string]interface{}{"path": "/tmp"},
			},
			wantError: false,
		},
		{
			name: "empty ID",
			task: Task{
				Type:        "analysis",
				Description: "Analyze files",
				Priority:    PriorityMedium,
			},
			wantError: true,
		},
		{
			name: "empty type",
			task: Task{
				ID:          "task-1",
				Description: "Analyze files",
				Priority:    PriorityMedium,
			},
			wantError: true,
		},
		{
			name: "empty description",
			task: Task{
				ID:       "task-1",
				Type:     "analysis",
				Priority: PriorityMedium,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResult_Success(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected bool
	}{
		{
			name:     "successful result",
			result:   Result{Success: true},
			expected: true,
		},
		{
			name:     "failed result",
			result:   Result{Success: false, Error: "something went wrong"},
			expected: false,
		},
		{
			name:     "result with error message",
			result:   Result{Success: true, Error: "warning message"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.Success)
		})
	}
}

func TestHealthStatus_IsHealthy(t *testing.T) {
	tests := []struct {
		name     string
		health   HealthStatus
		expected bool
	}{
		{
			name: "healthy status",
			health: HealthStatus{
				Status:    HealthStatusHealthy,
				Timestamp: time.Now(),
			},
			expected: true,
		},
		{
			name: "degraded status",
			health: HealthStatus{
				Status:    HealthStatusDegraded,
				Timestamp: time.Now(),
			},
			expected: false,
		},
		{
			name: "unhealthy status",
			health: HealthStatus{
				Status:    HealthStatusUnhealthy,
				Timestamp: time.Now(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isHealthy := tt.health.Status == HealthStatusHealthy
			assert.Equal(t, tt.expected, isHealthy)
		})
	}
}