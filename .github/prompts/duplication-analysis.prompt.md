# Code Duplication Analysis and Reduction

Analyze and reduce code duplication in this Go project:

## Analysis Phase
1. **Run Duplication Detection**:
   ```bash
   dupl -threshold 50 .
   ```

2. **Categorize Duplications**:
   - Validation logic patterns
   - Table-driven test boilerplate  
   - Agent creation/management patterns
   - Configuration handling
   - Error handling patterns

## Reduction Strategies

### For Validation Duplication
- Create centralized validation framework in `internal/common/validation.go`
- Extract common rules: Required, Positive, Range, OneOf
- Apply across config, request, and entity validation

### For Test Duplication  
- Create test helpers in `internal/testutil/helpers.go`
- Implement generic `RunValidationTests()` function
- Convert repetitive table-driven test patterns

### For Business Logic Duplication
- Extract common utilities to `internal/common/`
- Create reusable patterns (builders, registries, collectors)
- Apply DRY principle while maintaining readability

## Implementation Steps
1. Create common packages and utilities
2. Refactor existing code to use common patterns
3. Update tests to use helpers
4. Verify functionality with `go test ./...`
5. Re-run dupl to measure improvement

## Success Metrics
- Reduce clone groups from baseline measurement
- Maintain 100% test coverage
- Zero compilation errors
- Improved code maintainability

Use this prompt when `dupl` shows significant code duplication that needs systematic reduction.
