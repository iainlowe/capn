package communication

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iainlowe/capn/internal/agent"
	"github.com/iainlowe/capn/internal/crew"
)

func TestIntegration_IRCStyleCommunication(t *testing.T) {
	// Create message router and logger
	router := NewMessageRouter()
	logger := NewInMemoryLogger()
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
	require.NoError(t, router.RegisterAgent(captain))
	require.NoError(t, router.RegisterAgent(fileAgent))
	require.NoError(t, router.RegisterAgent(researchAgent))
	
	// Simulate IRC/Slack-style conversation
	
	// Step 1: Captain asks FileAgent to analyze files
	msg1 := agent.Message{
		ID:      "msg-001",
		Content: "Please analyze the Go files in ./src directory",
		Type:    agent.TaskMessage,
	}
	err := captain.SendMessage("file-001", msg1)
	require.NoError(t, err)
	
	// Step 2: FileAgent responds to Captain
	msg2 := agent.Message{
		ID:      "msg-002",
		Content: "Found 23 Go files, analyzing structure...",
		Type:    agent.ResponseMessage,
	}
	err = fileAgent.SendMessage("captain-001", msg2)
	require.NoError(t, err)
	
	// Step 3: FileAgent asks ResearchAgent for help
	msg3 := agent.Message{
		ID:      "msg-003",
		Content: "Hey Research, can you look up best practices for this pattern?",
		Type:    agent.TaskMessage,
	}
	err = fileAgent.SendMessage("research-001", msg3)
	require.NoError(t, err)
	
	// Step 4: ResearchAgent responds to FileAgent
	msg4 := agent.Message{
		ID:      "msg-004",
		Content: "Sure! Found 5 relevant patterns in Go documentation",
		Type:    agent.ResponseMessage,
	}
	err = researchAgent.SendMessage("file-001", msg4)
	require.NoError(t, err)
	
	// Verify all messages were logged
	allMessages := logger.GetAllMessages()
	assert.Len(t, allMessages, 4)
	
	// Verify message content
	assert.Equal(t, "captain-001", allMessages[0].From)
	assert.Equal(t, "file-001", allMessages[0].To)
	assert.Contains(t, allMessages[0].Content, "analyze the Go files")
	
	assert.Equal(t, "file-001", allMessages[1].From)
	assert.Equal(t, "captain-001", allMessages[1].To)
	assert.Contains(t, allMessages[1].Content, "Found 23 Go files")
	
	assert.Equal(t, "file-001", allMessages[2].From)
	assert.Equal(t, "research-001", allMessages[2].To)
	assert.Contains(t, allMessages[2].Content, "Hey Research")
	
	assert.Equal(t, "research-001", allMessages[3].From)
	assert.Equal(t, "file-001", allMessages[3].To)
	assert.Contains(t, allMessages[3].Content, "Found 5 relevant patterns")
	
	// Test IRC/Slack-style formatting
	formatted := logger.FormatMessage(agent.Message{
		From:      allMessages[0].From,
		To:        allMessages[0].To,
		Content:   allMessages[0].Content,
		Timestamp: allMessages[0].Timestamp,
	})
	assert.Regexp(t, `\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] captain-001 -> file-001: "Please analyze the Go files in ./src directory"`, formatted)
	
	// Test message history for individual agents
	captainHistory := logger.GetHistory("captain-001")
	assert.Len(t, captainHistory, 1) // Captain sent 1 message
	
	fileAgentHistory := logger.GetHistory("file-001")
	assert.Len(t, fileAgentHistory, 2) // FileAgent sent 2 messages
	
	researchHistory := logger.GetHistory("research-001")
	assert.Len(t, researchHistory, 1) // ResearchAgent sent 1 message
	
	// Test message search
	searchResults := logger.SearchMessages("Go files")
	assert.Len(t, searchResults, 2) // Both msg-001 and msg-002 contain references to files
	
	// More specific search
	searchResults = logger.SearchMessages("analyze the Go files")
	assert.Len(t, searchResults, 1)
	assert.Equal(t, "msg-001", searchResults[0].MessageID)
	
	searchResults = logger.SearchMessages("best practices")
	assert.Len(t, searchResults, 1)
	assert.Equal(t, "msg-003", searchResults[0].MessageID)
}

func TestIntegration_AgentTaskExecution(t *testing.T) {
	// Create agents
	fileAgent := crew.NewFileAgent("file-001", "FileAgent-1")
	networkAgent := crew.NewNetworkAgent("network-001", "NetworkAgent-1")
	researchAgent := crew.NewResearchAgent("research-001", "ResearchAgent-1")
	
	// Test FileAgent task execution
	fileTask := agent.Task{
		ID:          "task-001",
		Description: "List files in project directory",
		Parameters:  map[string]interface{}{"path": "./"},
	}
	
	result := fileAgent.Execute(context.Background(), fileTask)
	assert.True(t, result.IsSuccess())
	assert.Contains(t, result.Data.(string), "FileAgent-1")
	assert.Contains(t, result.Data.(string), "file system operations")
	
	// Test NetworkAgent task execution
	networkTask := agent.Task{
		ID:          "task-002",
		Description: "Fetch API documentation",
		Parameters:  map[string]interface{}{"url": "https://api.github.com"},
	}
	
	result = networkAgent.Execute(context.Background(), networkTask)
	assert.True(t, result.IsSuccess())
	assert.Contains(t, result.Data.(string), "NetworkAgent-1")
	assert.Contains(t, result.Data.(string), "API interactions")
	
	// Test ResearchAgent task execution
	researchTask := agent.Task{
		ID:          "task-003",
		Description: "Research Go design patterns",
		Parameters:  map[string]interface{}{"topic": "concurrency patterns"},
	}
	
	result = researchAgent.Execute(context.Background(), researchTask)
	assert.True(t, result.IsSuccess())
	assert.Contains(t, result.Data.(string), "ResearchAgent-1")
	assert.Contains(t, result.Data.(string), "information gathering")
}

func TestIntegration_BroadcastCommunication(t *testing.T) {
	// Create message router and logger
	router := NewMessageRouter()
	logger := NewInMemoryLogger()
	router.SetLogger(logger)
	
	// Create agents
	captain := agent.NewBaseAgent("captain-001", "Captain", agent.Captain)
	fileAgent := crew.NewFileAgent("file-001", "FileAgent-1")
	networkAgent := crew.NewNetworkAgent("network-001", "NetworkAgent-1")
	researchAgent := crew.NewResearchAgent("research-001", "ResearchAgent-1")
	
	// Register agents
	require.NoError(t, router.RegisterAgent(captain))
	require.NoError(t, router.RegisterAgent(fileAgent))
	require.NoError(t, router.RegisterAgent(networkAgent))
	require.NoError(t, router.RegisterAgent(researchAgent))
	
	// Captain broadcasts status message to all crew members
	broadcastMsg := agent.Message{
		ID:      "broadcast-001",
		From:    "captain-001",
		Content: "All agents, please report your current status",
		Type:    agent.StatusMessage,
	}
	
	// Exclude captain from receiving the broadcast
	err := router.BroadcastMessage(broadcastMsg, []string{"captain-001"})
	require.NoError(t, err)
	
	// Verify all crew agents received the message
	fileMessages := fileAgent.GetReceivedMessages()
	assert.Len(t, fileMessages, 1)
	assert.Equal(t, "captain-001", fileMessages[0].From)
	assert.Contains(t, fileMessages[0].Content, "report your current status")
	
	networkMessages := networkAgent.GetReceivedMessages()
	assert.Len(t, networkMessages, 1)
	
	researchMessages := researchAgent.GetReceivedMessages()
	assert.Len(t, researchMessages, 1)
	
	// Verify captain didn't receive the message
	captainMessages := captain.GetReceivedMessages()
	assert.Len(t, captainMessages, 0)
	
	// Verify all messages were logged
	allMessages := logger.GetAllMessages()
	assert.Len(t, allMessages, 3) // One message to each crew member
}

func TestIntegration_ConversationExample(t *testing.T) {
	// Recreate the example conversation from the issue description
	router := NewMessageRouter()
	logger := NewInMemoryLogger()
	router.SetLogger(logger)
	
	captain := agent.NewBaseAgent("captain", "Captain", agent.Captain)
	fileAgent := crew.NewFileAgent("file-1", "FileAgent-1")
	researchAgent := crew.NewResearchAgent("research-2", "ResearchAgent-2")
	
	captain.SetRouter(router)
	fileAgent.SetRouter(router)
	researchAgent.SetRouter(router)
	
	require.NoError(t, router.RegisterAgent(captain))
	require.NoError(t, router.RegisterAgent(fileAgent))
	require.NoError(t, router.RegisterAgent(researchAgent))
	
	// Simulate the conversation
	baseTime := time.Date(2025, 8, 3, 10, 15, 32, 0, time.UTC)
	
	// [2025-08-03 10:15:32] Captain -> FileAgent-1: "Please analyze the Go files in ./src directory"
	msg1 := agent.Message{
		ID:        "msg-001",
		Content:   "Please analyze the Go files in ./src directory",
		Type:      agent.TaskMessage,
		Timestamp: baseTime,
	}
	err := captain.SendMessage("file-1", msg1)
	require.NoError(t, err)
	
	// [2025-08-03 10:15:45] FileAgent-1 -> Captain: "Found 23 Go files, analyzing structure..."
	msg2 := agent.Message{
		ID:        "msg-002",
		Content:   "Found 23 Go files, analyzing structure...",
		Type:      agent.ResponseMessage,
		Timestamp: baseTime.Add(13 * time.Second),
	}
	err = fileAgent.SendMessage("captain", msg2)
	require.NoError(t, err)
	
	// [2025-08-03 10:16:12] FileAgent-1 -> ResearchAgent-2: "Hey Research, can you look up best practices for this pattern?"
	msg3 := agent.Message{
		ID:        "msg-003",
		Content:   "Hey Research, can you look up best practices for this pattern?",
		Type:      agent.TaskMessage,
		Timestamp: baseTime.Add(40 * time.Second),
	}
	err = fileAgent.SendMessage("research-2", msg3)
	require.NoError(t, err)
	
	// [2025-08-03 10:16:28] ResearchAgent-2 -> FileAgent-1: "Sure! Found 5 relevant patterns in Go documentation"
	msg4 := agent.Message{
		ID:        "msg-004",
		Content:   "Sure! Found 5 relevant patterns in Go documentation",
		Type:      agent.ResponseMessage,
		Timestamp: baseTime.Add(56 * time.Second),
	}
	err = researchAgent.SendMessage("file-1", msg4)
	require.NoError(t, err)
	
	// Verify the conversation was logged correctly
	allMessages := logger.GetAllMessages()
	require.Len(t, allMessages, 4)
	
	// Check the IRC/Slack-style formatting
	expectedFormats := []string{
		"[2025-08-03 10:15:32] captain -> file-1: \"Please analyze the Go files in ./src directory\"",
		"[2025-08-03 10:15:45] file-1 -> captain: \"Found 23 Go files, analyzing structure...\"",
		"[2025-08-03 10:16:12] file-1 -> research-2: \"Hey Research, can you look up best practices for this pattern?\"",
		"[2025-08-03 10:16:28] research-2 -> file-1: \"Sure! Found 5 relevant patterns in Go documentation\"",
	}
	
	for i, expectedFormat := range expectedFormats {
		formatted := logger.FormatMessage(agent.Message{
			From:      allMessages[i].From,
			To:        allMessages[i].To,
			Content:   allMessages[i].Content,
			Timestamp: allMessages[i].Timestamp,
		})
		assert.Equal(t, expectedFormat, formatted)
	}
}