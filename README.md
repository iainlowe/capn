# Capn

Capn is an LLM-based CLI agent system built in Go that orchestrates intelligent task execution through a hierarchical multi-agent architecture.

## Quick Start

### Installation

```bash
go build -o capn cmd/capn/main.go
```

### Basic Usage

```bash
# Show help
capn --help

# Execute a goal
capn execute "build a web server"

# Plan without executing (dry-run)
capn --dry-run execute "analyze this codebase"

# Use verbose logging
capn --verbose execute "research best practices"

# Use configuration file
capn --config config.yaml execute "deploy application"
```

### Configuration

Create a `config.yaml` file to customize Capn's behavior:

```yaml
global:
  verbose: true
  parallel: 8
  timeout: 10m

captain:
  max_concurrent_agents: 10
  planning_timeout: 60s

crew:
  timeouts:
    research: 300s
    coding: 600s

mcp:
  timeout: 15s
  retry_count: 5
```

See `config.example.yaml` for a complete configuration example.

### Commands

- `execute <goal>` - Plan and execute goals
- `status` - Show current operation status  
- `agents` - Manage agent configurations
- `mcp` - Manage MCP server connections

### Global Options

- `--config, -c` - Configuration file path
- `--verbose, -v` - Enable verbose logging
- `--dry-run` - Plan without execution
- `--timeout` - Global timeout duration
- `--parallel, -p` - Maximum parallel agents
