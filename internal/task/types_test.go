package task

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTaskStatus_String(t *testing.T) {
	tests := []struct {
		status   TaskStatus
		expected string
	}{
		{TaskStatusQueued, "queued"},
		{TaskStatusRunning, "running"},
		{TaskStatusCompleted, "completed"},
		{TaskStatusFailed, "failed"},
		{TaskStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestTaskStatus_IsActive(t *testing.T) {
	tests := []struct {
		status   TaskStatus
		expected bool
	}{
		{TaskStatusQueued, true},
		{TaskStatusRunning, true},
		{TaskStatusCompleted, false},
		{TaskStatusFailed, false},
		{TaskStatusCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsActive())
		})
	}
}

func TestTaskSummary_Creation(t *testing.T) {
	now := time.Now()
	summary := &TaskSummary{
		ID:      "test-001",
		Status:  TaskStatusRunning,
		Goal:    "test goal",
		Started: now,
	}

	assert.Equal(t, "test-001", summary.ID)
	assert.Equal(t, TaskStatusRunning, summary.Status)
	assert.Equal(t, "test goal", summary.Goal)
	assert.Equal(t, now, summary.Started)
}

func TestTaskDetails_Creation(t *testing.T) {
	now := time.Now()
	details := &TaskDetails{
		TaskSummary: TaskSummary{
			ID:      "test-001",
			Status:  TaskStatusRunning,
			Goal:    "test goal",
			Started: now,
		},
		Plan:         "test plan",
		Progress:     75.0,
		CurrentStep:  4,
		TotalSteps:   5,
		ActiveAgents: []string{"Agent-1", "Agent-2"},
	}

	assert.Equal(t, "test-001", details.ID)
	assert.Equal(t, TaskStatusRunning, details.Status)
	assert.Equal(t, "test plan", details.Plan)
	assert.Equal(t, 75.0, details.Progress)
	assert.Equal(t, 4, details.CurrentStep)
	assert.Equal(t, 5, details.TotalSteps)
	assert.Equal(t, []string{"Agent-1", "Agent-2"}, details.ActiveAgents)
}

func TestLogEntry_Creation(t *testing.T) {
	now := time.Now()
	entry := LogEntry{
		Timestamp: now,
		Level:     LogLevelInfo,
		Agent:     "TestAgent",
		Message:   "test message",
		Metadata:  map[string]interface{}{"key": "value"},
	}

	assert.Equal(t, now, entry.Timestamp)
	assert.Equal(t, LogLevelInfo, entry.Level)
	assert.Equal(t, "TestAgent", entry.Agent)
	assert.Equal(t, "test message", entry.Message)
	assert.Equal(t, "value", entry.Metadata["key"])
}

func TestTaskFilter_MatchesStatus(t *testing.T) {
	filter := TaskFilter{
		Status: []TaskStatus{TaskStatusRunning, TaskStatusCompleted},
	}

	summary1 := &TaskSummary{Status: TaskStatusRunning}
	summary2 := &TaskSummary{Status: TaskStatusCompleted}
	summary3 := &TaskSummary{Status: TaskStatusFailed}

	assert.True(t, filter.Matches(summary1))
	assert.True(t, filter.Matches(summary2))
	assert.False(t, filter.Matches(summary3))
}

func TestTaskFilter_MatchesKeywords(t *testing.T) {
	filter := TaskFilter{
		Keywords: []string{"security", "analysis"},
	}

	summary1 := &TaskSummary{Goal: "security analysis of codebase"}
	summary2 := &TaskSummary{Goal: "deploy to production"}
	summary3 := &TaskSummary{Goal: "analyze security issues"}

	assert.True(t, filter.Matches(summary1))
	assert.False(t, filter.Matches(summary2))
	assert.True(t, filter.Matches(summary3))
}

func TestTaskFilter_MatchesDateRange(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	filter := TaskFilter{
		DateRange: DateRange{
			Start: &yesterday,
			End:   &tomorrow,
		},
	}

	summary1 := &TaskSummary{Started: now}
	summary2 := &TaskSummary{Started: now.Add(-48 * time.Hour)}

	assert.True(t, filter.Matches(summary1))
	assert.False(t, filter.Matches(summary2))
}

func TestTaskFilter_EmptyFilter(t *testing.T) {
	filter := TaskFilter{}
	summary := &TaskSummary{
		Status: TaskStatusRunning,
		Goal:   "any task",
	}

	// Empty filter should match everything
	assert.True(t, filter.Matches(summary))
}

func TestDateRange_Contains(t *testing.T) {
	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Hour)

	tests := []struct {
		name     string
		dr       DateRange
		t        time.Time
		expected bool
	}{
		{
			name:     "within range",
			dr:       DateRange{Start: &start, End: &end},
			t:        now,
			expected: true,
		},
		{
			name:     "before range",
			dr:       DateRange{Start: &start, End: &end},
			t:        start.Add(-1 * time.Hour),
			expected: false,
		},
		{
			name:     "after range",
			dr:       DateRange{Start: &start, End: &end},
			t:        end.Add(1 * time.Hour),
			expected: false,
		},
		{
			name:     "no start time",
			dr:       DateRange{End: &end},
			t:        start.Add(-1 * time.Hour),
			expected: true,
		},
		{
			name:     "no end time",
			dr:       DateRange{Start: &start},
			t:        end.Add(1 * time.Hour),
			expected: true,
		},
		{
			name:     "no range",
			dr:       DateRange{},
			t:        now,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.dr.Contains(tt.t))
		})
	}
}