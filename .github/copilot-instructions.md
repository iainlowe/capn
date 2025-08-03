# Copilot Instructions

This file contains instructions for GitHub Copilot to understand the project context and coding standards.

## Project Overview

Capn is an LLM-based CLI agent system built in Go that orchestrates intelligent task execution through a hierarchical multi-agent architecture. The system consists of a central planning agent (the "Captain") that creates execution plans using strategic thinking, potentially leveraging MCP (Model Context Protocol) servers, and spawns specialized sub-agents (the "Crew") to accomplish complex goals through distributed, parallel execution.

## Test-Driven Development (TDD) Requirements
**ALWAYS follow TDD practices when writing code:**

1. **Write Tests First**: Before implementing any feature or function, write the corresponding tests
2. **Red Phase**: Run tests and verify they fail (confirming the test is actually testing something)
3. **Green Phase**: Write the minimal code necessary to make the tests pass
4. **Refactor Phase**: Improve code quality while keeping tests green
5. **Code Coverage**: Maintain code coverage above 90% at all times
6. **Frequent Test Runs**: Run all tests frequently during development and fix any failures immediately
7. **Test Quality**: Write comprehensive, meaningful tests that cover edge cases and error conditions

### TDD Workflow

- Write a failing test
- Run tests to confirm failure
- Write minimal implementation to pass the test
- Run tests to confirm they pass
- Refactor if needed while keeping tests green
- Calculate and verify code coverage remains above 90%
- Repeat for next feature/requirement

## Coding Standards

- Follow established patterns in the codebase
- Use appropriate testing frameworks for the language/technology stack
- Maintain clean, readable code
- Write descriptive test names that explain what is being tested
- Organize tests logically with proper setup, execution, and assertion phases
- Mock external dependencies in unit tests
- Include integration tests for system boundaries

## Testing Guidelines

- Unit tests should be fast, isolated, and deterministic
- Test both happy path and error scenarios
- Use descriptive assertions that explain what went wrong when they fail
- Keep test code as clean and maintainable as production code
- Run the full test suite before committing code
- Never commit code that breaks existing tests

## Shell Command Guidelines

- **Always prefix shell commands with `timeout`** using reasonable durations:
  - Quick commands (ls, cat, echo): `timeout 10s`
  - Test runs: `timeout 300s` (5 minutes)
  - Build operations: `timeout 600s` (10 minutes)
  - Installation/setup: `timeout 1200s` (20 minutes)
  - Long-running processes: `timeout 3600s` (1 hour)
- Use appropriate timeout values based on expected command duration
- Include timeout to prevent hanging processes and improve reliability
- Example: `timeout 30s npm test` instead of just `npm test`

## Go Coding Standards

- **Follow Go conventions**: Use `gofmt`, `goimports`, and `golint` for code formatting and linting
- **Package naming**: Use short, lowercase package names without underscores
- **Error handling**: Always handle errors explicitly; use `if err != nil` pattern
- **Interface design**: Keep interfaces small and focused (interface segregation principle)
- **Naming conventions**:
  - Use camelCase for unexported functions/variables
  - Use PascalCase for exported functions/variables
  - Use descriptive names that explain purpose
- **Comments**: Write godoc-compatible comments for all exported functions, types, and packages
- **Testing**: Use table-driven tests with `t.Run()` for subtests
- **Dependency management**: Use Go modules (`go.mod`) for dependency management

## Go Directory Structure
Follow the standard Go project layout:
```
/
├── cmd/                    # Main applications
│   └── myapp/
│       └── main.go
├── internal/              # Private application code
│   ├── app/
│   ├── pkg/
│   └── handlers/
├── pkg/                   # Public library code
├── api/                   # OpenAPI/Swagger specs, protocol definitions
├── web/                   # Web application specific components
├── configs/               # Configuration file templates
├── init/                  # System init configs
├── scripts/               # Build, install, analysis scripts
├── build/                 # Packaging and CI
├── deployments/           # IaaS, PaaS, system configs
├── test/                  # Additional external test apps and data
├── docs/                  # Design and user documents
├── tools/                 # Supporting tools
├── examples/              # Examples for your applications/libraries
├── vendor/                # Application dependencies (managed by go mod)
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── Makefile              # Build automation
└── README.md
```

## Go Testing Guidelines

- **Test file naming**: Use `_test.go` suffix for test files
- **Test function naming**: Use `TestFunctionName` format
- **Benchmark tests**: Use `BenchmarkFunctionName` format
- **Example tests**: Use `ExampleFunctionName` format for godoc examples
- **Table-driven tests**: Structure tests with input/expected output tables
- **Test coverage**: Use `go test -cover` to measure coverage
- **Mock generation**: Use `go generate` with tools like `mockgen` for interface mocks
- **Integration tests**: Use build tags (`// +build integration`) to separate unit and integration tests

## Project Management Guidelines

- **Directory Creation**: Only create directories that are immediately needed and will be used
- **Cleanup**: Always clean up temporary files, unused directories, and resources before ending your turn
- **Minimal Structure**: Start with minimal directory structure and expand only as requirements emerge
- **No Speculative Directories**: Don't create directories "just in case" - create them when there's actual content to place in them

## Architecture
Refer to ARCHITECTURE.md for detailed architecture information.
