package captain

import (
	"context"
	"encoding/json"
	"time"
)

// Task represents a single task in an execution plan
type Task struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Command      string            `json:"command,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	EstimatedDuration time.Duration `json:"estimated_duration"`
	Priority     int               `json:"priority"`
	Status       TaskStatus        `json:"status"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusSkipped    TaskStatus = "skipped"
)

// ExecutionPlan represents a complete execution plan for achieving a goal
type ExecutionPlan struct {
	ID          string            `json:"id"`
	Goal        string            `json:"goal"`
	Description string            `json:"description"`
	Tasks       []Task            `json:"tasks"`
	CreatedAt   time.Time         `json:"created_at"`
	EstimatedDuration time.Duration `json:"estimated_duration"`
	Reasoning   string            `json:"reasoning"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CompletionRequest represents a request to the LLM provider
type CompletionRequest struct {
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	Model       string    `json:"model,omitempty"`
}

// CompletionResponse represents a response from the LLM provider
type CompletionResponse struct {
	Content   string `json:"content"`
	TokensUsed int   `json:"tokens_used"`
	Model     string `json:"model"`
}

// Message represents a single message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Result represents the result of executing a task or plan
type Result struct {
	TaskID    string    `json:"task_id"`
	Success   bool      `json:"success"`
	Output    string    `json:"output"`
	Error     string    `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time `json:"timestamp"`
}

// LLMProvider defines the interface for LLM providers
type LLMProvider interface {
	GenerateCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float64, error)
}

// PlanningEngine defines the interface for planning engines
type PlanningEngine interface {
	AnalyzeGoal(ctx context.Context, goal string) (*ExecutionPlan, error)
	ValidatePlan(ctx context.Context, plan *ExecutionPlan) error
	OptimizePlan(ctx context.Context, plan *ExecutionPlan) (*ExecutionPlan, error)
}

// Config represents the captain-specific configuration
type Config struct {
	OpenAIAPIKey        string        `yaml:"openai_api_key"`
	Model               string        `yaml:"model"`
	MaxTokens           int           `yaml:"max_tokens"`
	Temperature         float32       `yaml:"temperature"`
	MaxConcurrentAgents int           `yaml:"max_concurrent_agents"`
	PlanningTimeout     time.Duration `yaml:"planning_timeout"`
	RetryAttempts       int           `yaml:"retry_attempts"`
	RetryDelay          time.Duration `yaml:"retry_delay"`
}

// DefaultConfig returns a default captain configuration
func DefaultConfig() *Config {
	return &Config{
		Model:               "gpt-4",
		MaxTokens:           2000,
		Temperature:         0.7,
		MaxConcurrentAgents: 5,
		PlanningTimeout:     30 * time.Second,
		RetryAttempts:       3,
		RetryDelay:          time.Second,
	}
}

// ToJSON converts a struct to JSON string
func (p *ExecutionPlan) ToJSON() (string, error) {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON creates an ExecutionPlan from JSON string
func (p *ExecutionPlan) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), p)
}

// ToJSON converts a task to JSON string
func (t *Task) ToJSON() (string, error) {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON creates a Task from JSON string
func (t *Task) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), t)
}