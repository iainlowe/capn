package main

import (
	"fmt"
	"time"

	"github.com/iainlowe/capn/internal/agent"
	"github.com/iainlowe/capn/internal/communication"
	"github.com/iainlowe/capn/internal/crew"
)

// ExampleUsage demonstrates the IRC/Slack-style agent communication system
func ExampleUsage() {
	fmt.Println("=== Capn Agent Communication System Demo ===")
	
	// Create message router and logger
	router := communication.NewMessageRouter()
	logger := communication.NewInMemoryLogger()
	router.SetLogger(logger)
	
	// Create agents
	captain := agent.NewBaseAgent("captain-001", "Captain", agent.Captain)
	fileAgent := crew.NewFileAgent("file-001", "FileAgent-1")
	researchAgent := crew.NewResearchAgent("research-001", "ResearchAgent-2")
	
	// Set up router for all agents
	captain.SetRouter(router)
	fileAgent.SetRouter(router)
	researchAgent.SetRouter(router)
	
	// Register agents with router
	router.RegisterAgent(captain)
	router.RegisterAgent(fileAgent)
	router.RegisterAgent(researchAgent)
	
	fmt.Println("\n=== Registered Agents ===")
	for _, agentID := range router.GetRegisteredAgents() {
		agent, _ := router.GetAgent(agentID)
		fmt.Printf("- %s (%s) - Status: %s\n", agent.Name(), agent.Type().String(), agent.Status().String())
	}
	
	// Simulate IRC/Slack-style conversation
	fmt.Println("\n=== Agent Conversation ===")
	
	// Step 1: Captain asks FileAgent to analyze files
	msg1 := agent.Message{
		ID:      "msg-001",
		Content: "Please analyze the Go files in ./src directory",
		Type:    agent.TaskMessage,
	}
	captain.SendMessage("file-001", msg1)
	fmt.Println(logger.FormatMessage(agent.Message{
		From:      "captain-001",
		To:        "file-001",
		Content:   msg1.Content,
		Timestamp: time.Now(),
	}))
	
	// Step 2: FileAgent responds to Captain
	msg2 := agent.Message{
		ID:      "msg-002",
		Content: "Found 23 Go files, analyzing structure...",
		Type:    agent.ResponseMessage,
	}
	fileAgent.SendMessage("captain-001", msg2)
	fmt.Println(logger.FormatMessage(agent.Message{
		From:      "file-001",
		To:        "captain-001",
		Content:   msg2.Content,
		Timestamp: time.Now(),
	}))
	
	// Step 3: FileAgent asks ResearchAgent for help
	msg3 := agent.Message{
		ID:      "msg-003",
		Content: "Hey Research, can you look up best practices for this pattern?",
		Type:    agent.TaskMessage,
	}
	fileAgent.SendMessage("research-001", msg3)
	fmt.Println(logger.FormatMessage(agent.Message{
		From:      "file-001",
		To:        "research-001",
		Content:   msg3.Content,
		Timestamp: time.Now(),
	}))
	
	// Step 4: ResearchAgent responds to FileAgent
	msg4 := agent.Message{
		ID:      "msg-004",
		Content: "Sure! Found 5 relevant patterns in Go documentation",
		Type:    agent.ResponseMessage,
	}
	researchAgent.SendMessage("file-001", msg4)
	fmt.Println(logger.FormatMessage(agent.Message{
		From:      "research-001",
		To:        "file-001",
		Content:   msg4.Content,
		Timestamp: time.Now(),
	}))
	
	// Show communication history
	fmt.Println("\n=== Communication History ===")
	allMessages := logger.GetAllMessages()
	fmt.Printf("Total messages exchanged: %d\n", len(allMessages))
	
	fmt.Println("\n=== Captain's Sent Messages ===")
	captainHistory := logger.GetHistory("captain-001")
	for _, msg := range captainHistory {
		fmt.Printf("- To %s: %s\n", msg.To, msg.Content)
	}
	
	fmt.Println("\n=== Search Messages containing 'Go files' ===")
	searchResults := logger.SearchMessages("Go files")
	for _, result := range searchResults {
		fmt.Printf("- %s -> %s: %s\n", result.From, result.To, result.Content)
	}
	
	fmt.Println("\n=== Agent Status After Communication ===")
	for _, agentID := range router.GetRegisteredAgents() {
		agent, _ := router.GetAgent(agentID)
		fmt.Printf("- %s: Status=%s, Health=%s\n", 
			agent.Name(), agent.Status().String(), agent.Health().String())
	}
}

func main() {
	ExampleUsage()
}