# Geoffrey Architecture

This document provides an overview of Geoffrey's architecture and design decisions.

## System Overview

Geoffrey follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Presentation Layer                           │
│                  (CLI + Terminal UI)                            │
├─────────────────────────────────────────────────────────────────┤
│                     Pipeline Layer                              │
│     (Interview → Design → DevPlan → Review)                     │
├─────────────────────────────────────────────────────────────────┤
│                    Execution Layer                              │
│              (Task Executor + Live Monitor)                     │
├─────────────────────────────────────────────────────────────────┤
│                   Integration Layer                             │
│         (API Bridge + Multi-Provider Support)                   │
├─────────────────────────────────────────────────────────────────┤
│                   Persistence Layer                             │
│              (State Store + Git Manager)                        │
├─────────────────────────────────────────────────────────────────┤
│                  Infrastructure Layer                           │
│      (Configuration + Logging + Error Handling)                 │
└─────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. CLI (Cobra)
- Command-line interface for all operations
- Argument parsing and validation
- Help text and documentation
- Exit code management

### 2. Terminal UI (Bubbletea)
- Interactive terminal interface
- Real-time streaming output
- Progress tracking
- Keyboard navigation

### 3. Interview Engine
- Five-phase interview workflow
- LLM-powered follow-up questions
- State persistence for pause/resume
- JSON export of gathered data

### 4. Design Generator
- Architecture document generation
- System diagrams and data flows
- Technology rationale
- Risk assessment

### 5. DevPlan Generator
- Phase and task generation
- Token and cost estimation
- Phase manipulation (merge, split, reorder)
- Master plan export

### 6. Phase Reviewer
- Automated phase analysis
- Issue detection and categorization
- Improvement suggestions
- Selective improvement application

### 7. API Bridge
- Multi-provider orchestration
- Response normalization
- Retry with exponential backoff
- Rate limit and quota tracking

### 8. Task Executor
- Phase and task execution
- Real-time output streaming
- Pause/resume/skip support
- Blocker detection

### 9. Git Manager
- Repository initialization
- Commit creation with metadata
- Tag management for checkpoints
- Conflict detection

### 10. State Store (SQLite)
- Embedded database
- Project state persistence
- Token usage tracking
- Checkpoint management

### 11. Configuration Manager
- Multi-source configuration
- API key management
- Model preferences
- Budget limits

### 12. Token Counter & Cost Estimator
- Token counting per model
- Cost calculation
- Statistics aggregation
- Budget monitoring

## Data Flow

### Interview → Design → DevPlan → Review → Develop

1. **Interview Phase**
   - User answers questions
   - System generates follow-ups
   - Data saved to SQLite
   - JSON exported and committed to Git

2. **Design Phase**
   - Interview data loaded
   - Architecture generated via LLM
   - Document saved to SQLite
   - Markdown committed to Git

3. **DevPlan Phase**
   - Architecture and interview loaded
   - Phases generated via LLM
   - Token/cost estimates calculated
   - Phase files committed to Git

4. **Review Phase**
   - DevPlan loaded
   - Each phase analyzed
   - Issues categorized
   - Improvements suggested and applied

5. **Development Phase**
   - Tasks executed sequentially
   - Output streamed in real-time
   - Progress tracked in SQLite
   - Changes committed to Git

## State Management

### SQLite Database Schema

```sql
-- Core entities
projects
interview_data
architectures
phases
tasks

-- Tracking
checkpoints
token_usage
rate_limits
quotas
blockers

-- Configuration
config
token_stats_cache
```

### Git Integration

All artifacts are committed to Git:
- Interview JSON: `interview.json`
- Architecture: `architecture.md`
- DevPlan: `devplan/phase-NNN.md`
- Master plan: `devplan/devplan.md`
- Detours: `devplan/detours/`
- Decisions: `devplan/decisions.md`

## Error Handling

### Error Categories

1. **User Errors**: Invalid input, missing config
   - Display helpful message
   - Suggest correction
   - Exit with code 1

2. **API Errors**: Rate limits, auth failures
   - Retry with exponential backoff
   - Fall back to alternative model
   - Track in rate limit store

3. **System Errors**: Database corruption, disk full
   - Save state before exit
   - Display error with context
   - Suggest recovery steps

4. **Git Errors**: Conflicts, uncommitted changes
   - Pause execution
   - Display conflict details
   - Request user resolution

### Retry Strategy

```go
func RetryWithBackoff(operation func() error, maxRetries int) error {
    for attempt := 0; attempt < maxRetries; attempt++ {
        err := operation()
        if err == nil {
            return nil
        }
        
        if !isRetryable(err) {
            return err
        }
        
        delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
        time.Sleep(delay)
    }
    return fmt.Errorf("max retries exceeded")
}
```

## Testing Strategy

### Dual Testing Approach

1. **Unit Tests**: Specific examples and edge cases
   - Test individual functions
   - Test error conditions
   - Test integration points

2. **Property-Based Tests**: Universal properties
   - State preservation round-trip
   - Phase dependency ordering
   - Cost calculation accuracy
   - Configuration precedence

### Test Organization

```
test/
├── integration/
│   ├── pipeline_test.go      # End-to-end pipeline
│   ├── git_test.go            # Git operations
│   └── resume_test.go         # Resume capability
└── properties/
    ├── state_test.go          # State preservation
    ├── devplan_test.go        # DevPlan properties
    ├── cost_test.go           # Cost calculation
    └── config_test.go         # Configuration
```

## Security Considerations

### API Key Storage

- Stored in `~/.geoffrussy/config.yaml`
- File permissions: 0600 (owner read/write only)
- Never logged or displayed
- Validated before use

### Error Messages

- No sensitive data in error messages
- No API keys in logs
- No PII in telemetry

### Git Commits

- No API keys in commit messages
- No sensitive data in committed files
- Clear metadata for audit trail

## Performance Considerations

### Database

- SQLite with WAL mode for concurrency
- Indexes on frequently queried columns
- Prepared statements for queries
- Connection pooling

### API Calls

- Rate limit tracking to avoid throttling
- Quota monitoring to prevent overages
- Retry with exponential backoff
- Request batching where possible

### Memory

- Stream large responses
- Limit in-memory state
- Use database for persistence
- Clean up resources promptly

## Scalability

### Current Limitations

- Single-user, single-project focus
- Local SQLite database
- Synchronous execution

### Future Enhancements

- Multi-project support
- Parallel task execution
- Distributed state store
- Team collaboration features

## Dependencies

### Core Dependencies

- **Cobra**: CLI framework
- **Bubbletea**: Terminal UI framework
- **Viper**: Configuration management
- **SQLite**: Embedded database
- **gopter**: Property-based testing

### Provider SDKs

- OpenAI Go SDK
- Anthropic Go SDK
- Custom HTTP clients for other providers

## Build and Distribution

### Build Process

1. Go compilation with CGO for SQLite
2. Static linking of SQLite
3. Version injection via ldflags
4. Cross-compilation for all platforms

### Release Process

1. Tag version: `git tag v1.0.0`
2. Push tag: `git push origin v1.0.0`
3. GitHub Actions builds all platforms
4. Binaries uploaded to GitHub Releases
5. Checksums generated and uploaded

### Supported Platforms

- Linux: AMD64, ARM64
- macOS: AMD64, ARM64
- Windows: AMD64, ARM64

## Monitoring and Observability

### Logging

- Structured logging with levels
- Log to file and stderr
- Rotation and retention
- No sensitive data

### Metrics

- Token usage by provider
- Cost by phase
- API call latency
- Error rates

### Tracing

- Pipeline stage transitions
- Task execution timeline
- API call traces
- Git operations

## Future Architecture

### Planned Enhancements

1. **Plugin System**: Extend with custom providers
2. **Web UI**: Browser-based interface
3. **Team Features**: Shared projects and collaboration
4. **Cloud Sync**: Optional cloud state backup
5. **Analytics**: Usage patterns and insights

### Migration Path

- Maintain backward compatibility
- Database migrations for schema changes
- Configuration version tracking
- Deprecation warnings for breaking changes
