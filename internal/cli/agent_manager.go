package cli

import (
	"fmt"
	"strings"

	"github.com/iainlowe/capn/internal/agent"
	"github.com/iainlowe/capn/internal/communication"
	"github.com/iainlowe/capn/internal/crew"
)

// AgentManager manages the lifecycle of agents and their communication
type AgentManager struct {
	router *communication.MessageRouter
	logger communication.CommunicationLogger
}

// NewAgentManager creates a new agent manager
func NewAgentManager() *AgentManager {
	router := communication.NewMessageRouter()
	logger := communication.NewInMemoryLogger()
	router.SetLogger(logger)
	
	return &AgentManager{
		router: router,
		logger: logger,
	}
}

// ListAgents returns information about all registered agents
func (am *AgentManager) ListAgents() string {
	var result strings.Builder
	
	registeredAgents := am.router.GetRegisteredAgents()
	if len(registeredAgents) == 0 {
		result.WriteString("No agents currently registered.\n")
		return result.String()
	}
	
	result.WriteString(fmt.Sprintf("Registered Agents (%d):\n", len(registeredAgents)))
	result.WriteString("═══════════════════════════════════════════\n")
	
	for _, agentID := range registeredAgents {
		a, err := am.router.GetAgent(agentID)
		if err != nil {
			continue
		}
		
		result.WriteString(fmt.Sprintf("• %s\n", a.Name()))
		result.WriteString(fmt.Sprintf("  ID: %s\n", a.ID()))
		result.WriteString(fmt.Sprintf("  Type: %s\n", a.Type().String()))
		result.WriteString(fmt.Sprintf("  Status: %s\n", a.Status().String()))
		result.WriteString(fmt.Sprintf("  Health: %s\n", a.Health().String()))
		result.WriteString("\n")
	}
	
	return result.String()
}

// GetCommunicationHistory returns formatted communication history
func (am *AgentManager) GetCommunicationHistory(limit int) string {
	var result strings.Builder
	
	allMessages := am.logger.GetAllMessages()
	if len(allMessages) == 0 {
		result.WriteString("No communication history available.\n")
		return result.String()
	}
	
	result.WriteString("Communication History:\n")
	result.WriteString("═══════════════════════════════════════════\n")
	
	// Limit the number of messages shown
	start := 0
	if limit > 0 && len(allMessages) > limit {
		start = len(allMessages) - limit
		result.WriteString(fmt.Sprintf("(Showing last %d messages)\n\n", limit))
	}
	
	for i := start; i < len(allMessages); i++ {
		msg := allMessages[i]
		formatted := am.logger.FormatMessage(agent.Message{
			From:      msg.From,
			To:        msg.To,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
		})
		result.WriteString(formatted + "\n")
	}
	
	return result.String()
}

// SearchMessages searches for messages containing the query
func (am *AgentManager) SearchMessages(query string) string {
	var result strings.Builder
	
	searchResults := am.logger.SearchMessages(query)
	if len(searchResults) == 0 {
		result.WriteString(fmt.Sprintf("No messages found containing '%s'.\n", query))
		return result.String()
	}
	
	result.WriteString(fmt.Sprintf("Messages containing '%s' (%d results):\n", query, len(searchResults)))
	result.WriteString("═══════════════════════════════════════════\n")
	
	for _, msg := range searchResults {
		formatted := am.logger.FormatMessage(agent.Message{
			From:      msg.From,
			To:        msg.To,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
		})
		result.WriteString(formatted + "\n")
	}
	
	return result.String()
}

// CreateExampleAgents creates some example agents for demonstration
func (am *AgentManager) CreateExampleAgents() error {
	// Create sample agents
	captain := agent.NewBaseAgent("captain-001", "Captain", agent.Captain)
	fileAgent := crew.NewFileAgent("file-001", "FileAgent-1")
	networkAgent := crew.NewNetworkAgent("network-001", "NetworkAgent-1")
	researchAgent := crew.NewResearchAgent("research-001", "ResearchAgent-1")
	
	// Set up routing
	captain.SetRouter(am.router)
	fileAgent.SetRouter(am.router)
	networkAgent.SetRouter(am.router)
	researchAgent.SetRouter(am.router)
	
	// Register agents
	if err := am.router.RegisterAgent(captain); err != nil {
		return err
	}
	if err := am.router.RegisterAgent(fileAgent); err != nil {
		return err
	}
	if err := am.router.RegisterAgent(networkAgent); err != nil {
		return err
	}
	if err := am.router.RegisterAgent(researchAgent); err != nil {
		return err
	}
	
	return nil
}

// SimulateConversation creates a sample conversation between agents
func (am *AgentManager) SimulateConversation() error {
	captain, err := am.router.GetAgent("captain-001")
	if err != nil {
		return fmt.Errorf("captain not found: %w", err)
	}
	
	fileAgent, err := am.router.GetAgent("file-001")
	if err != nil {
		return fmt.Errorf("file agent not found: %w", err)
	}
	
	researchAgent, err := am.router.GetAgent("research-001")
	if err != nil {
		return fmt.Errorf("research agent not found: %w", err)
	}
	
	// Simulate conversation
	msg1 := agent.Message{
		ID:      "msg-001",
		Content: "Please analyze the Go files in ./src directory",
		Type:    agent.TaskMessage,
	}
	captain.SendMessage("file-001", msg1)
	
	msg2 := agent.Message{
		ID:      "msg-002",
		Content: "Found 23 Go files, analyzing structure...",
		Type:    agent.ResponseMessage,
	}
	fileAgent.SendMessage("captain-001", msg2)
	
	msg3 := agent.Message{
		ID:      "msg-003",
		Content: "Hey Research, can you look up best practices for this pattern?",
		Type:    agent.TaskMessage,
	}
	fileAgent.SendMessage("research-001", msg3)
	
	msg4 := agent.Message{
		ID:      "msg-004",
		Content: "Sure! Found 5 relevant patterns in Go documentation",
		Type:    agent.ResponseMessage,
	}
	researchAgent.SendMessage("file-001", msg4)
	
	return nil
}