# Geoffrey Quick Start Guide

Welcome to Geoffrey! This guide will get you up and running in minutes.

## Prerequisites

- Go 1.21+ (for building from source)
- Git
- GCC (for SQLite support)

## Setup (5 minutes)

### 1. Get the Code

```bash
git clone https://github.com/mojomast/geoffrussy.git
cd geoffrussy
```

### 2. Install Dependencies

```bash
go mod download
go mod tidy
```

### 3. Build

```bash
make build
```

### 4. Verify Installation

```bash
./bin/geoffrussy version
```

You should see: `Geoffrey version dev`

## Using Docker (Alternative)

If you prefer Docker:

```bash
# Build
docker-compose build

# Run
docker-compose run geoffrussy version
```

## Next Steps

### For Users

1. **Initialize a project**:
   ```bash
   cd your-project
   geoffrussy init
   ```

2. **Start the interview**:
   ```bash
   geoffrussy interview
   ```

3. **Use MCP (optional)**:
   ```bash
   geoffrussy mcp-server --project-path /path/to/project
   ```

4. **Read the full documentation**:
    - [README.md](README.md) - Complete overview
    - [docs/SETUP.md](docs/SETUP.md) - Detailed setup guide
    - [docs/mcp-integration.md](docs/mcp-integration.md) - MCP integration guide
    - [docs/AGENT_MCP_GUIDE.md](docs/AGENT_MCP_GUIDE.md) - Agent guide for MCP

### For Developers

1. **Run tests**:
   ```bash
   make test
   ```

2. **Read the architecture**:
   - [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
   - [CONTRIBUTING.md](CONTRIBUTING.md)

3. **Start developing**:
   - Check [docs/PROJECT_STATUS.md](docs/PROJECT_STATUS.md) for current status
   - See [.kiro/specs/geoffrey-ai-agent/tasks.md](.kiro/specs/geoffrey-ai-agent/tasks.md) for task list

## Common Commands

```bash
# Build
make build

# Test
make test

# Format code
make fmt

# Run linters
make lint

# Clean
make clean

# Build for all platforms
make build-all
```

## Troubleshooting

### "go: command not found"
Install Go from [golang.org](https://golang.org/dl/)

### "gcc: command not found"
- Linux: `sudo apt-get install gcc libsqlite3-dev`
- macOS: `xcode-select --install`
- Windows: Install MinGW-w64

### Build errors
```bash
go mod download
go mod tidy
make clean
make build
```

## Getting Help

- 📖 [Full Documentation](docs/)
- 🐛 [Report Issues](https://github.com/mojomast/geoffrussy/issues)
- 💬 [Discussions](https://github.com/mojomast/geoffrussy/discussions)

## What's Next?

Geoffrey is currently in **Phase 1** (Infrastructure Setup). The next phases will implement:

- Phase 2: State Store (SQLite)
- Phase 3: Configuration Manager
- Phase 4: API Bridge
- Phase 5+: Core features (Interview, Design, DevPlan, etc.)

See [docs/PROJECT_STATUS.md](docs/PROJECT_STATUS.md) for detailed progress.

---

**Note**: Many features are not yet implemented. This is the foundation for the Geoffrey AI Coding Agent. Check the [tasks.md](.kiro/specs/geoffrey-ai-agent/tasks.md) for the complete implementation roadmap.
