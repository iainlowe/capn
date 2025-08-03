# Capn - Distributed CLI Agent System

Capn is a sophisticated CLI agent system built in Go that orchestrates intelligent task execution through a hierarchical multi-agent architecture. The system consists of a central planning agent (the "Captain") that creates execution plans using OpenAI-powered reasoning and spawns specialized sub-agents (the "Crew") to accomplish complex goals.

## Features

- **LLM-Powered Planning**: Uses OpenAI GPT models for intelligent goal analysis and task decomposition
- **Chain-of-Thought Reasoning**: Complex problem solving with structured reasoning
- **Plan Validation**: Automatic feasibility checking and plan optimization
- **Dry-Run Mode**: Visualize execution plans without running actual commands
- **Hierarchical Architecture**: Captain-Crew model for distributed task execution
- **Robust Error Handling**: Comprehensive error handling with retry mechanisms

## Installation

Build from source:

```bash
go build -o capn ./cmd/capn
```

## Configuration

Create a configuration file (see `example-config.yaml`):

```yaml
captain:
  openai_api_key: "your-openai-api-key-here"
  model: "gpt-4"
  max_tokens: 2000
  temperature: 0.7
```

Or set the `OPENAI_API_KEY` environment variable.

## Usage

### Analyze and Plan Goals

```bash
# Create an intelligent execution plan
capn plan "create a REST API in Go"

# Plan with configuration file
capn -c config.yaml plan "build a web application"
```

### Execute with Planning

```bash
# Dry-run execution (plan visualization)
capn execute --dry-run "deploy application to production"

# Full execution (when implemented)
capn execute "setup development environment"
```

### Global Options

```bash
capn --help
capn --verbose plan "complex task"
capn --parallel 10 execute "parallel task"
```

## Architecture

- **Captain Agent**: Central orchestrator with LLM-powered planning
- **Planning Engine**: Goal analysis, task decomposition, and optimization
- **OpenAI Integration**: GPT models for intelligent reasoning
- **Execution Engine**: Task execution and monitoring (planned)
- **Configuration System**: YAML-based configuration with validation

## Development

Run tests:

```bash
go test ./...
go test -cover ./...
```

Current test coverage: >90%

## License

See LICENSE file for details.
