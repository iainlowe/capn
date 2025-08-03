# capn

Capn is a distributed CLI agent system that uses a Captain agent with LLM-powered planning to break down goals into executable tasks and orchestrate their execution.

## Usage

```bash
capn execute [--plan-only | --dry-run] <goal>
```

- `--plan-only` generates and displays an execution plan without running it.
- `--dry-run` creates a plan and simulates execution, showing the expected results.

To enable LLM-backed planning, provide an OpenAI API key via configuration or the `OPENAI_API_KEY` environment variable.
