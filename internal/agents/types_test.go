package agents

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/iainlowe/capn/internal/testutil"
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
	testCases := []testutil.ValidationTestCase[Message]{
		{
			Name: "valid message",
			Input: Message{
				ID:        "msg-1",
				From:      "agent-1",
				To:        "agent-2",
				Content:   "Hello world",
				Type:      MessageTypeText,
				Timestamp: time.Now(),
			},
			WantError: false,
		},
		{
			Name: "empty ID",
			Input: Message{
				From:      "agent-1",
				To:        "agent-2",
				Content:   "Hello world",
				Type:      MessageTypeText,
				Timestamp: time.Now(),
			},
			WantError: true,
		},
		{
			Name: "empty from",
			Input: Message{
				ID:        "msg-1",
				To:        "agent-2",
				Content:   "Hello world",
				Type:      MessageTypeText,
				Timestamp: time.Now(),
			},
			WantError: true,
		},
		{
			Name: "empty to",
			Input: Message{
				ID:        "msg-1",
				From:      "agent-1",
				Content:   "Hello world",
				Type:      MessageTypeText,
				Timestamp: time.Now(),
			},
			WantError: true,
		},
		{
			Name: "empty content",
			Input: Message{
				ID:        "msg-1",
				From:      "agent-1",
				To:        "agent-2",
				Type:      MessageTypeText,
				Timestamp: time.Now(),
			},
			WantError: true,
		},
	}

	testutil.RunValidationTests(t, testCases, func(msg Message) error {
		return msg.Validate()
	})
}

func TestTask_Validation(t *testing.T) {
	testCases := []testutil.ValidationTestCase[Task]{
		{
			Name: "valid task",
			Input: Task{
				ID:          "task-1",
				Type:        "analysis",
				Description: "Analyze files",
				Priority:    PriorityMedium,
				Data:        map[string]interface{}{"path": "/tmp"},
			},
			WantError: false,
		},
		{
			Name: "empty ID",
			Input: Task{
				Type:        "analysis",
				Description: "Analyze files",
				Priority:    PriorityMedium,
			},
			WantError: true,
		},
		{
			Name: "empty type",
			Input: Task{
				ID:          "task-1",
				Description: "Analyze files",
				Priority:    PriorityMedium,
			},
			WantError: true,
		},
		{
			Name: "empty description",
			Input: Task{
				ID:       "task-1",
				Type:     "analysis",
				Priority: PriorityMedium,
			},
			WantError: true,
		},
	}

	testutil.RunValidationTests(t, testCases, func(task Task) error {
		return task.Validate()
	})
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