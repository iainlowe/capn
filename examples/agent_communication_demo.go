package main

import (
	"fmt"
	"log"
	"time"

	"github.com/iainlowe/capn/internal/agents"
	"github.com/iainlowe/capn/internal/agents/crew"
)

// setupCommunicationInfrastructure initializes the message router and logger
func setupCommunicationInfrastructure() (*agents.MessageRouter, *agents.MemoryCommunicationLogger) {
	router := agents.NewMessageRouter()
	logger := agents.NewMemoryCommunicationLogger()
	router.SetLogger(logger)
	return router, logger
}

// createAndRegisterAgents creates crew agents and registers them with the router
func createAndRegisterAgents(router *agents.MessageRouter) (agents.Agent, agents.Agent, agents.Agent) {
	// Create crew agents
	fileAgent := crew.NewFileAgent("file-1", "FileAgent-1")
	networkAgent := crew.NewNetworkAgent("net-1", "NetworkAgent-1")
	researchAgent := crew.NewResearchAgent("research-1", "ResearchAgent-1")

	// Configure agents with router
	fileAgent.SetRouter(router)
	networkAgent.SetRouter(router)
	researchAgent.SetRouter(router)

	// Register agents with router
	if err := router.RegisterAgent(fileAgent); err != nil {
		log.Fatalf("Failed to register file agent: %v", err)
	}
	if err := router.RegisterAgent(networkAgent); err != nil {
		log.Fatalf("Failed to register network agent: %v", err)
	}
	if err := router.RegisterAgent(researchAgent); err != nil {
		log.Fatalf("Failed to register research agent: %v", err)
	}

	fmt.Printf("✓ Registered %d agents with the communication system\n", len(router.GetRegisteredAgents()))
	return fileAgent, networkAgent, researchAgent
}

// simulateCommunicationScenario demonstrates the IRC/Slack-style messaging between agents
func simulateCommunicationScenario(logger *agents.MemoryCommunicationLogger, fileAgent, researchAgent agents.Agent) {
	fmt.Println("\n=== Communication Scenario ===")

	// 1. Captain asks FileAgent to analyze files
	captainToFile := agents.Message{
		ID:        "msg-1",
		From:      "Captain",
		To:        "file-1",
		Content:   "Please analyze the Go files in ./src directory",
		Type:      agents.MessageTypeText,
		Timestamp: time.Date(2025, 8, 3, 10, 15, 32, 0, time.UTC),
	}

	// Manually log captain message (captain would be registered in full implementation)
	logger.LogMessage("Captain", "file-1", captainToFile)
	fmt.Println("Captain -> FileAgent-1: Analysis request sent")

	// 2. FileAgent responds to Captain
	time.Sleep(100 * time.Millisecond) // Simulate processing time
	fileToCaptain := agents.Message{
		ID:        "msg-2",
		From:      "file-1",
		To:        "Captain",
		Content:   "Found 23 Go files, analyzing structure...",
		Type:      agents.MessageTypeText,
		Timestamp: time.Date(2025, 8, 3, 10, 15, 45, 0, time.UTC),
	}
	logger.LogMessage("file-1", "Captain", fileToCaptin)
	fmt.Println("FileAgent-1 -> Captain: Status update sent")

	// 3. FileAgent asks ResearchAgent for help
	time.Sleep(100 * time.Millisecond)
	fileToResearch := agents.Message{
		ID:        "msg-3",
		From:      "file-1",
		To:        "research-1",
		Content:   "Hey Research, can you look up best practices for this pattern?",
		Type:      agents.MessageTypeText,
		Timestamp: time.Date(2025, 8, 3, 10, 16, 12, 0, time.UTC),
	}

	if err := fileAgent.SendMessage("research-1", fileToResearch); err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}
	fmt.Println("FileAgent-1 -> ResearchAgent-1: Research request sent")

	// 4. ResearchAgent responds to FileAgent
	time.Sleep(100 * time.Millisecond)
	researchToFile := agents.Message{
		ID:        "msg-4",
		From:      "research-1",
		To:        "file-1",
		Content:   "Sure! Found 5 relevant patterns in Go documentation",
		Type:      agents.MessageTypeText,
		Timestamp: time.Date(2025, 8, 3, 10, 16, 28, 0, time.UTC),
	}

	if err := researchAgent.SendMessage("file-1", researchToFile); err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}
	fmt.Println("ResearchAgent-1 -> FileAgent-1: Research results sent")
}

// displayCommunicationLog shows all messages in the communication log
func displayCommunicationLog(logger *agents.MemoryCommunicationLogger) {
	fmt.Println("\n=== Complete Communication Log ===")
	allMessages := logger.GetAllMessages()
	for _, msgLog := range allMessages {
		fmt.Println(msgLog.Formatted)
	}
}

// demonstrateSearchCapabilities shows the message search functionality
func demonstrateSearchCapabilities(logger *agents.MemoryCommunicationLogger) {
	fmt.Println("\n=== Search Capabilities ===")
	searchResults := logger.SearchMessages("Go files")
	fmt.Printf("Search for 'Go files' found %d messages:\n", len(searchResults))
	for _, result := range searchResults {
		fmt.Printf("  - %s\n", result.Formatted)
	}

	searchResults = logger.SearchMessages("research")
	fmt.Printf("\nSearch for 'research' found %d messages:\n", len(searchResults))
	for _, result := range searchResults {
		fmt.Printf("  - %s\n", result.Formatted)
	}
}

// showAgentHistory displays communication history for a specific agent
func showAgentHistory(logger *agents.MemoryCommunicationLogger) {
	fmt.Println("\n=== Agent Communication History ===")
	fileHistory := logger.GetHistory("file-1")
	fmt.Printf("FileAgent-1 communication history (%d messages):\n", len(fileHistory))
	for _, msgLog := range fileHistory {
		fmt.Printf("  - %s\n", msgLog.Formatted)
	}
}

// demonstrateHealthStatus shows agent status and health information
func demonstrateHealthStatus(router *agents.MessageRouter) {
	fmt.Println("\n=== Agent Status and Health ===")
	registeredAgents := router.GetRegisteredAgents()
	for _, agent := range registeredAgents {
		status := agent.Status()
		health := agent.Health()
		fmt.Printf("%s (%s): Status=%s, Health=%s\n",
			agent.Name(), agent.ID(), status, health.Status)
	}
}

// demonstrateBroadcastMessaging shows broadcast functionality to all agents
func demonstrateBroadcastMessaging(router *agents.MessageRouter, logger *agents.MemoryCommunicationLogger) {
	fmt.Println("\n=== Broadcast Messaging ===")
	broadcastMsg := agents.Message{
		ID:        "msg-broadcast",
		From:      "system",
		To:        "all",
		Content:   "System maintenance will begin in 10 minutes",
		Type:      agents.MessageTypeText,
		Timestamp: time.Now(),
	}

	// Manually broadcast since we don't have a system agent registered
	registeredAgents := router.GetRegisteredAgents()
	for _, agent := range registeredAgents {
		if agent.ID() != "system" { // Don't send to self
			copyMsg := broadcastMsg
			copyMsg.To = agent.ID()
			if err := agent.ReceiveMessage(copyMsg); err != nil {
				log.Printf("Failed to deliver broadcast to %s: %v", agent.ID(), err)
			} else {
				logger.LogMessage("system", agent.ID(), copyMsg)
				fmt.Printf("  ✓ Delivered to %s\n", agent.Name())
			}
		}
	}
}

// printFinalSummary displays the completion summary
func printFinalSummary(logger *agents.MemoryCommunicationLogger) {
	fmt.Println("\n=== Demo Complete ===")
	fmt.Printf("Total messages logged: %d\n", len(logger.GetAllMessages()))
	fmt.Println("IRC/Slack-style agent communication system is fully operational!")
}

func main() {
	fmt.Println("=== Capn Agent Communication System Demo ===")
	fmt.Println("Demonstrating IRC/Slack-style agent communication")

	// Initialize communication infrastructure
	router, logger := setupCommunicationInfrastructure()

	// Create and register agents
	fileAgent, _, researchAgent := createAndRegisterAgents(router)

	// Simulate the communication scenario from the issue description
	simulateCommunicationScenario(logger, fileAgent, researchAgent)

	// Display the complete communication log
	displayCommunicationLog(logger)

	// Demonstrate search capabilities
	demonstrateSearchCapabilities(logger)

	// Show agent history
	showAgentHistory(logger)

	// Demonstrate agent health and status
	demonstrateHealthStatus(router)

	// Demonstrate broadcast messaging
	demonstrateBroadcastMessaging(router, logger)

	// Print final summary
	printFinalSummary(logger)
}
