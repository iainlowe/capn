package crew

import (
	"context"
	"fmt"
	"time"

	"github.com/iainlowe/capn/internal/agents"
)

// CrewAgentFactory creates crew agents
type CrewAgentFactory struct{}

// NewCrewAgentFactory creates a new crew agent factory
func NewCrewAgentFactory() *CrewAgentFactory {
	return &CrewAgentFactory{}
}

// CreateAgent creates a crew agent of the specified type
func (f *CrewAgentFactory) CreateAgent(id, name string, agentType agents.AgentType) (agents.Agent, error) {
	switch agentType {
	case agents.AgentTypeFile:
		return NewFileAgent(id, name), nil
	case agents.AgentTypeNetwork:
		return NewNetworkAgent(id, name), nil  
	case agents.AgentTypeResearch:
		return NewResearchAgent(id, name), nil
	default:
		return nil, fmt.Errorf("unsupported crew agent type: %s", agentType)
	}
}

// FileAgent handles file system operations
type FileAgent struct {
	*agents.BaseAgent
}

// NewFileAgent creates a new file agent
func NewFileAgent(id, name string) *FileAgent {
	return &FileAgent{
		BaseAgent: agents.NewBaseAgent(id, name, agents.AgentTypeFile),
	}
}

// SetRouter sets the message router for this agent
func (f *FileAgent) SetRouter(router *agents.MessageRouter) {
	f.BaseAgent.SetRouter(router)
}

// GetReceivedMessages returns received messages (for testing)
func (f *FileAgent) GetReceivedMessages() []agents.Message {
	return f.BaseAgent.GetReceivedMessages()
}

// Execute executes file-related tasks
func (f *FileAgent) Execute(ctx context.Context, task agents.Task) agents.Result {
	// Set status to busy during execution
	f.BaseAgent.SetStatus(agents.AgentStatusBusy)
	defer f.BaseAgent.SetStatus(agents.AgentStatusIdle)
	
	startTime := time.Now()
	
	// Extract task data
	pathVal, ok := task.Data["path"]
	path, okStr := pathVal.(string)
	if !ok || !okStr || path == "" {
		return agents.Result{
			TaskID:    task.ID,
			Success:   false,
			Output:    "FileAgent error: missing or invalid 'path' in task data",
			Duration:  0,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"agent_type": "file",
				"operation":  task.Type,
			},
		}
	}
	pattern, _ := task.Data["pattern"].(string)
	
	switch task.Type {
	case "file_analysis":
		output = fmt.Sprintf("FileAgent executed file operation: analyzing files at %s", path)
		if pattern != "" {
			output += fmt.Sprintf(" with pattern %s", pattern)
		}
		// Simulate finding files
		output += " - Found files and performed analysis"
		
	case "file_read":
		output = fmt.Sprintf("FileAgent executed file operation: reading file %s", path)
		
	case "file_write":
		output = fmt.Sprintf("FileAgent executed file operation: writing to file %s", path)
		
	case "file_search":
		query, _ := task.Data["query"].(string)
		output = fmt.Sprintf("FileAgent executed file operation: searching for '%s' in %s", query, path)
		
	default:
		output = fmt.Sprintf("FileAgent executed file operation: %s", task.Description)
	}
	
	return agents.Result{
		TaskID:    task.ID,
		Success:   success,
		Output:    output,
		Duration:  time.Since(startTime),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"agent_type": "file",
			"operation": task.Type,
		},
	}
}

// NetworkAgent handles API interactions and web operations
type NetworkAgent struct {
	*agents.BaseAgent
}

// NewNetworkAgent creates a new network agent
func NewNetworkAgent(id, name string) *NetworkAgent {
	return &NetworkAgent{
		BaseAgent: agents.NewBaseAgent(id, name, agents.AgentTypeNetwork),
	}
}

// SetRouter sets the message router for this agent
func (n *NetworkAgent) SetRouter(router *agents.MessageRouter) {
	n.BaseAgent.SetRouter(router)
}

// GetReceivedMessages returns received messages (for testing)
func (n *NetworkAgent) GetReceivedMessages() []agents.Message {
	return n.BaseAgent.GetReceivedMessages()
}

// Execute executes network-related tasks
func (n *NetworkAgent) Execute(ctx context.Context, task agents.Task) agents.Result {
	// Set status to busy during execution
	n.BaseAgent.SetStatus(agents.AgentStatusBusy)
	defer n.BaseAgent.SetStatus(agents.AgentStatusIdle)
	
	startTime := time.Now()
	
	// Extract task data
	url, _ := task.Data["url"].(string)
	method, _ := task.Data["method"].(string)
	
	// Simulate network operation based on task type
	var output string
	var success bool = true
	
	switch task.Type {
	case "api_call":
		output = fmt.Sprintf("NetworkAgent executed network operation: %s request to %s", method, url)
		// Simulate API response
		output += " - Received successful response"
		
	case "web_scrape":
		output = fmt.Sprintf("NetworkAgent executed network operation: scraping data from %s", url)
		
	case "download":
		output = fmt.Sprintf("NetworkAgent executed network operation: downloading from %s", url)
		
	case "upload":
		output = fmt.Sprintf("NetworkAgent executed network operation: uploading to %s", url)
		
	default:
		output = fmt.Sprintf("NetworkAgent executed network operation: %s", task.Description)  
	}
	
	return agents.Result{
		TaskID:    task.ID,
		Success:   success,
		Output:    output,
		Duration:  time.Since(startTime),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"agent_type": "network",
			"operation": task.Type,
		},
	}
}

// ResearchAgent handles information gathering and analysis
type ResearchAgent struct {
	*agents.BaseAgent
}

// NewResearchAgent creates a new research agent
func NewResearchAgent(id, name string) *ResearchAgent {
	return &ResearchAgent{
		BaseAgent: agents.NewBaseAgent(id, name, agents.AgentTypeResearch),
	}
}

// SetRouter sets the message router for this agent
func (r *ResearchAgent) SetRouter(router *agents.MessageRouter) {
	r.BaseAgent.SetRouter(router)
}

// GetReceivedMessages returns received messages (for testing)
func (r *ResearchAgent) GetReceivedMessages() []agents.Message {
	return r.BaseAgent.GetReceivedMessages()
}

// Execute executes research-related tasks
func (r *ResearchAgent) Execute(ctx context.Context, task agents.Task) agents.Result {
	// Set status to busy during execution
	r.BaseAgent.SetStatus(agents.AgentStatusBusy)
	defer r.BaseAgent.SetStatus(agents.AgentStatusIdle)
	
	startTime := time.Now()
	
	// Extract task data
	topic, _ := task.Data["topic"].(string)
	depth, _ := task.Data["depth"].(string)
	
	// Simulate research operation based on task type
	var output string
	var success bool = true
	
	switch task.Type {
	case "research":
		output = fmt.Sprintf("ResearchAgent executed research operation: researching '%s'", topic)
		if depth != "" {
			output += fmt.Sprintf(" with %s analysis", depth)
		}
		// Simulate research results
		output += " - Found relevant information and patterns"
		
	case "analysis":
		output = fmt.Sprintf("ResearchAgent executed research operation: analyzing %s", topic)
		
	case "documentation":
		output = fmt.Sprintf("ResearchAgent executed research operation: documenting %s", topic)
		
	case "best_practices":
		output = fmt.Sprintf("ResearchAgent executed research operation: finding best practices for %s", topic)
		
	default:
		output = fmt.Sprintf("ResearchAgent executed research operation: %s", task.Description)
	}
	
	return agents.Result{
		TaskID:    task.ID,
		Success:   success,
		Output:    output,
		Duration:  time.Since(startTime),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"agent_type": "research",
			"operation": task.Type,
		},
	}
}