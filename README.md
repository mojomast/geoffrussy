# Geoffrey AI Coding Agent

Geoffrey is a next-generation AI-powered development orchestration platform that reimagines human-AI collaboration on software projects. The system prioritizes deep project understanding through a multi-stage iterative pipeline: **Interview → Architecture Design → DevPlan Generation → Phase Review**.

## Features

- 🎯 **Deep Project Understanding**: Five-phase interactive interview to gather comprehensive requirements
- 🏗️ **Architecture-First Approach**: Generate complete system architecture before writing code
- 📋 **Executable DevPlans**: Break down projects into 7-10 phases with 3-5 tasks each
- 🔍 **Automated Review**: AI-powered phase review to catch issues before development
- 🤖 **Multi-Model Support**: Use OpenAI, Anthropic, ZAI (GLM), Ollama, and more
- 💰 **Cost Tracking**: Monitor token usage and costs across all API calls
- 📊 **Rate Limit Monitoring**: Track and respect API rate limits and quotas
- 🔄 **Checkpoint System**: Save progress and rollback when needed
- 📈 **Real-Time Progress Monitor**: Track tasks, phases, completion percentage, and token usage
- ⏸️ **Phase Control**: Stop after current phase or continue through all phases automatically
- 🎨 **Interactive Terminal UI**: Beautiful terminal interface with ASCII art banner
- 📦 **Single Binary**: No dependencies, works on Linux, macOS, and Windows
- 🔌 **MCP Integration**: Model Context Protocol support for autonomous AI agents

## Installation

### Download Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/mojomast/geoffrussy/releases):

```bash
# Linux (AMD64)
wget https://github.com/mojomast/geoffrussy/releases/latest/download/geoffrussy-linux-amd64
chmod +x geoffrussy-linux-amd64
sudo mv geoffrussy-linux-amd64 /usr/local/bin/geoffrussy

# macOS (ARM64)
wget https://github.com/mojomast/geoffrussy/releases/latest/download/geoffrussy-darwin-arm64
chmod +x geoffrussy-darwin-arm64
sudo mv geoffrussy-darwin-arm64 /usr/local/bin/geoffrussy

# Windows (AMD64)
# Download geoffrussy-windows-amd64.exe and add to PATH
```

### Build from Source

Requirements:
- Go 1.21 or later
- GCC (for SQLite)
- Git

```bash
git clone https://github.com/mojomast/geoffrussy.git
cd geoffrussy
make build
sudo make install
```

### Using Docker

```bash
# Build the image
docker-compose build

# Run Geoffrey
docker-compose run geoffrussy version

# Development environment
docker-compose up -d geoffrussy-dev
docker-compose exec geoffrussy-dev sh
```

## Quick Start

### 1. Initialize Geoffrey

```bash
cd your-project
geoffrussy init
```

This will:
- Create configuration directory (`~/.geoffrussy/`)
- Prompt for API keys (OpenAI, Anthropic, etc.)
- Initialize SQLite database
- Set up Git repository if needed

### 2. Start the Interview

```bash
geoffrussy interview
```

Geoffrey will guide you through five phases:
1. **Project Essence**: Problem statement, target users, success metrics
2. **Technical Constraints**: Language, performance, scale, compliance
3. **Integration Points**: APIs, databases, authentication
4. **Scope Definition**: MVP features, timeline, resources
5. **Refinement & Validation**: Review and confirm all information

### 3. Generate Architecture

```bash
geoffrussy design
```

Geoffrey will create a comprehensive architecture document including:
- System diagrams
- Component breakdown
- Data flow diagrams
- Technology rationale
- Scaling strategy
- API contracts
- Database schema
- Security approach
- Observability strategy
- Deployment architecture
- Risk assessment

### 4. Generate DevPlan

```bash
geoffrussy plan
```

Geoffrey will generate 7-10 executable phases with 3-5 tasks each, following this structure:
- Phase 000: Setup & Infrastructure
- Phase 001: Database & Models
- Phase 002: Core API
- Phase 003: Authentication & Authorization
- Phase 004: Frontend Foundation
- Phase 005: Real-time Sync
- Phase 006: Integrations
- Phase 007: Testing & Validation
- Phase 008: Performance & Observability
- Phase 009: Deployment & Hardening

### 5. Review the Plan

```bash
geoffrussy review
```

Geoffrey will analyze the DevPlan for:
- Clarity and completeness
- Dependencies and ordering
- Scope and feasibility
- Risks and testing gaps
- Integration issues

### 6. Execute Development

```bash
# Execute all phases until complete
geoffrussy develop

# Execute specific phase and stop
geoffrussy develop --phase phase-5 --stop-after-phase

# Use a specific model
geoffrussy develop --model glm-4.7
```

Geoffrey will execute each phase, streaming real-time output and allowing you to:
- Pause and resume execution
- Skip tasks
- Request detours (mid-execution changes)
- Handle blockers

The execution monitor displays:
- **Project progress**: Tasks completed/total, phases completed/total, completion percentage
- **Phase and task tracking**: Current phase ID and task ID
- **Elapsed time**: Time since execution started
- **Token usage**: Input and output tokens consumed
- **Real-time updates**: Live stream of task execution output

## Commands

```bash
geoffrussy init              # Initialize project configuration
geoffrussy interview         # Start or resume interview phase
geoffrussy design            # Generate or review architecture
geoffrussy plan              # Generate or review DevPlan
geoffrussy review            # Run phase review and validation
geoffrussy develop           # Execute development phases
geoffrussy develop --model <model>        # Use specific model (e.g., glm-4.7, gpt-4)
geoffrussy develop --phase <id>          # Execute specific phase
geoffrussy develop --stop-after-phase     # Stop after completing current phase
geoffrussy status            # Show current progress
geoffrussy stats             # Show token usage and cost statistics
geoffrussy quota             # Check rate limits and quotas
geoffrussy checkpoint        # Create or list checkpoints
geoffrussy rollback          # Rollback to a checkpoint
geoffrussy mcp-server        # Start MCP server for AI agents
geoffrussy version           # Print version number
```

## Configuration

Geoffrey supports configuration via:
1. Command-line flags (highest precedence)
2. Environment variables
3. Config file (`~/.geoffrussy/config.yaml`)

### Example Configuration

```yaml
# ~/.geoffrussy/config.yaml
api_keys:
  openai: sk-...
  anthropic: sk-ant-...
  zai: <your-zai-api-key>  # For GLM models (glm-4.7, etc.)
  ollama: http://localhost:11434

default_models:
  interview: gpt-4
  design: claude-3-5-sonnet
  devplan: gpt-4
  review: claude-3-5-sonnet
  develop: glm-4.7  # Supports: glm-4.7, gpt-4, claude-3-5-sonnet, etc.

budget_limit: 100.0  # USD
verbose_logging: false

# MCP Server Configuration (optional)
mcp:
  enabled: true
  log_level: info
  server_mode: stdio
```

### Environment Variables

```bash
export GEOFFRUSSY_OPENAI_API_KEY=sk-...
export GEOFFRUSSY_ANTHROPIC_API_KEY=sk-ant-...
export GEOFFRUSSY_BUDGET_LIMIT=100.0
```

## MCP (Model Context Protocol) Integration

Geoffrey supports the Model Context Protocol, enabling AI agents to autonomously use Geoffrey for building software.

### Quick Start

Start the MCP server:

```bash
geoffrussy mcp-server --project-path /path/to/your/project
```

**Optional:** Enable debug logging for troubleshooting:

```bash
geoffrussy mcp-server --project-path /path/to/your/project --debug
```

### Claude for Desktop Configuration

Add Geoffrey to your `claude_desktop_config.json`:

**macOS/Linux:**
```bash
code ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

**Windows:**
```bash
code %APPDATA%\Claude\claude_desktop_config.json
```

**Configuration:**
```json
{
  "mcpServers": {
    "geoffrey": {
      "command": "/absolute/path/to/geoffrussy",
      "args": ["mcp-server", "--project-path", "/absolute/path/to/project"]
    }
  }
}
```

**Note:** Use absolute paths for both the `command` and `--project-path` arguments.

### Available Tools

- `get_status` - Get project status
- `get_stats` - Get token usage statistics
- `list_phases` - List all development phases
- `create_checkpoint` - Create a checkpoint
- `list_checkpoints` - List all checkpoints

### Available Resources

- `project://status` - Project status (JSON)
- `project://architecture` - Architecture documentation (Markdown)
- `project://devplan` - Development plan (JSON)
- `project://phases` - All phases (JSON)
- `project://interview` - Interview requirements (JSON)
- `project://checkpoints` - All checkpoints (JSON)
- `project://stats` - Token usage statistics (JSON)

### New Features

- ✅ **Debug Mode:** `--debug` flag enables verbose logging for troubleshooting
- ✅ **Relative Path Support:** Works with both absolute and relative project paths
- ✅ **Proper Error Handling:** All errors returned as JSON-RPC errors with helpful messages
- ✅ **Parameter Validation:** Type-safe validation for all tool parameters

### Documentation

See [docs/mcp-integration.md](docs/mcp-integration.md) for complete MCP documentation including:
- Detailed tool and resource schemas
- Protocol specifications
- Use cases for autonomous agents
- Troubleshooting guide
- Advanced usage examples

See [mcphandoff.md](mcphandoff.md) for comprehensive testing report and handoff information.

## Development

### Prerequisites

- Go 1.21+
- GCC (for SQLite)
- Make
- Docker (optional)

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run unit tests only
make test-unit

# Run property tests
make test-property

# Run integration tests
make test-integration

# Format code
make fmt

# Run linters
make lint

# Clean build artifacts
make clean
```

### Project Structure

```
geoffrussy/
├── cmd/
│   └── geoffrussy/          # Main entry point
├── internal/
│   ├── cli/                 # CLI commands (Cobra)
│   ├── tui/                 # Terminal UI (Bubbletea)
│   ├── interview/           # Interview engine
│   ├── design/              # Design generator
│   ├── devplan/             # DevPlan generator
│   ├── review/              # Phase reviewer
│   ├── api/                 # API bridge and providers
│   ├── executor/            # Task executor
│   ├── git/                 # Git manager
│   ├── state/               # State store (SQLite)
│   ├── config/              # Configuration manager
│   ├── token/               # Token counter
│   └── cost/                # Cost estimator
├── test/
│   ├── integration/         # Integration tests
│   └── properties/          # Property-based tests
├── docs/                    # Documentation
├── .github/
│   └── workflows/           # CI/CD pipelines
├── Dockerfile               # Production container
├── docker-compose.yml       # Development environment
├── Makefile                 # Build automation
└── go.mod                   # Go module definition
```

## Testing

Geoffrey uses a dual testing approach:

### Unit Tests
Verify specific examples, edge cases, and error conditions:
```bash
make test-unit
```

### Property-Based Tests
Verify universal properties across all inputs using [gopter](https://github.com/leanovate/gopter):
```bash
make test-property
```

### Integration Tests
Verify end-to-end workflows:
```bash
make test-integration
```

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on:
- Code of conduct
- Development workflow
- Testing requirements
- Pull request process

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) for Terminal UI
- Uses [gopter](https://github.com/leanovate/gopter) for property-based testing
- Uses [SQLite](https://www.sqlite.org/) for state persistence

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/mojomast/geoffrussy/issues)
- 💬 [Discussions](https://github.com/mojomast/geoffrussy/discussions)

## Roadmap

See the [DevPlan](.kiro/specs/geoffrey-ai-agent/) for the complete implementation roadmap.

Current status: **Phase 1 - Project Setup and Infrastructure** ✅
