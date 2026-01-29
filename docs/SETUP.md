# Geoffrey Setup Guide

This guide will help you set up Geoffrey for development or usage.

## Prerequisites

### For Users

- **Operating System**: Linux, macOS, or Windows
- **Git**: Version 2.0 or later
- **API Keys**: At least one of:
  - OpenAI API key
  - Anthropic API key
  - Ollama (local, no key needed)
  - Other supported providers

### For Developers

- **Go**: Version 1.21 or later
- **GCC**: For compiling SQLite
  - Linux: `sudo apt-get install gcc libsqlite3-dev`
  - macOS: Xcode Command Line Tools (`xcode-select --install`)
  - Windows: MinGW-w64 or TDM-GCC
- **Make**: Build automation
- **Docker**: Optional, for containerized development

## Installation

### Option 1: Download Pre-built Binary (Recommended for Users)

1. Visit the [releases page](https://github.com/mojomast/geoffrussy/releases)
2. Download the binary for your platform:
   - Linux AMD64: `geoffrussy-linux-amd64`
   - Linux ARM64: `geoffrussy-linux-arm64`
   - macOS AMD64: `geoffrussy-darwin-amd64`
   - macOS ARM64: `geoffrussy-darwin-arm64`
   - Windows AMD64: `geoffrussy-windows-amd64.exe`
   - Windows ARM64: `geoffrussy-windows-arm64.exe`

3. Make it executable (Linux/macOS):
   ```bash
   chmod +x geoffrussy-*
   ```

4. Move to PATH:
   ```bash
   # Linux/macOS
   sudo mv geoffrussy-* /usr/local/bin/geoffrussy
   
   # Windows: Add to PATH or move to a directory in PATH
   ```

5. Verify installation:
   ```bash
   geoffrussy version
   ```

### Option 2: Build from Source (Recommended for Developers)

1. **Clone the repository**:
   ```bash
   git clone https://github.com/mojomast/geoffrussy.git
   cd geoffrussy
   ```

2. **Install dependencies**:
   ```bash
   # Download Go modules
   go mod download
   
   # Verify modules
   go mod verify
   ```

3. **Build the binary**:
   ```bash
   make build
   ```

4. **Install to GOPATH/bin** (optional):
   ```bash
   make install
   ```

5. **Verify installation**:
   ```bash
   geoffrussy version
   # or if not installed to PATH:
   ./bin/geoffrussy version
   ```

### Option 3: Using Docker

1. **Clone the repository**:
   ```bash
   git clone https://github.com/mojomast/geoffrussy.git
   cd geoffrussy
   ```

2. **Build the Docker image**:
   ```bash
   docker-compose build
   ```

3. **Run Geoffrey**:
   ```bash
   docker-compose run geoffrussy version
   ```

4. **For development**:
   ```bash
   # Start development container
   docker-compose up -d geoffrussy-dev
   
   # Enter the container
   docker-compose exec geoffrussy-dev sh
   
   # Inside container
   make build
   make test
   ```

## Configuration

### First-Time Setup

1. **Initialize Geoffrey**:
   ```bash
   cd your-project
   geoffrussy init
   ```

2. **Enter API keys** when prompted:
   - OpenAI API key (starts with `sk-`)
   - Anthropic API key (starts with `sk-ant-`)
   - Other providers as needed

3. **Configuration location**:
   - Linux/macOS: `~/.geoffrussy/config.yaml`
   - Windows: `%USERPROFILE%\.geoffrussy\config.yaml`

### Manual Configuration

Create or edit `~/.geoffrussy/config.yaml`:

```yaml
# API Keys
api_keys:
  openai: sk-your-openai-key
  anthropic: sk-ant-your-anthropic-key
  ollama: http://localhost:11434
  firmware: your-firmware-key
  requesty: your-requesty-key
  z: your-z-key
  kimi: your-kimi-key

# Default models for each stage
default_models:
  interview: gpt-4
  design: claude-3-5-sonnet
  devplan: gpt-4
  review: claude-3-5-sonnet
  develop: gpt-4

# Budget limit in USD
budget_limit: 100.0

# Logging
verbose_logging: false
log_file: ~/.geoffrussy/geoffrussy.log

# Database
database_path: ~/.geoffrussy/geoffrussy.db
```

### Environment Variables

You can also configure via environment variables:

```bash
# API Keys
export GEOFFRUSSY_OPENAI_API_KEY=sk-your-key
export GEOFFRUSSY_ANTHROPIC_API_KEY=sk-ant-your-key
export GEOFFRUSSY_OLLAMA_URL=http://localhost:11434

# Settings
export GEOFFRUSSY_BUDGET_LIMIT=100.0
export GEOFFRUSSY_VERBOSE=true
export GEOFFRUSSY_CONFIG_PATH=~/.geoffrussy
```

### Configuration Precedence

Geoffrey uses the following precedence (highest to lowest):

1. **Command-line flags**: `--config`, `--verbose`, etc.
2. **Environment variables**: `GEOFFRUSSY_*`
3. **Config file**: `~/.geoffrussy/config.yaml`

## Setting Up API Keys

### OpenAI

1. Visit [OpenAI Platform](https://platform.openai.com/)
2. Sign up or log in
3. Navigate to API Keys
4. Create a new API key
5. Copy the key (starts with `sk-`)
6. Add to Geoffrey config

### Anthropic

1. Visit [Anthropic Console](https://console.anthropic.com/)
2. Sign up or log in
3. Navigate to API Keys
4. Create a new API key
5. Copy the key (starts with `sk-ant-`)
6. Add to Geoffrey config

### Ollama (Local)

1. Install Ollama from [ollama.ai](https://ollama.ai/)
2. Start Ollama: `ollama serve`
3. Pull a model: `ollama pull llama2`
4. Geoffrey will auto-detect at `http://localhost:11434`

### Other Providers

Refer to each provider's documentation for API key setup:
- Firmware.ai
- Requesty.ai
- Z.ai
- Kimi

## Verifying Setup

### Check Configuration

```bash
# View current configuration
cat ~/.geoffrussy/config.yaml

# Test API keys
geoffrussy init --validate
```

### Check Database

```bash
# Database location
ls -lh ~/.geoffrussy/geoffrussy.db

# View database schema (requires sqlite3)
sqlite3 ~/.geoffrussy/geoffrussy.db ".schema"
```

### Check Git Integration

```bash
# Initialize a test project
mkdir test-project
cd test-project
git init
geoffrussy init
```

## Troubleshooting

### "go: command not found"

**Solution**: Install Go from [golang.org](https://golang.org/dl/)

### "gcc: command not found"

**Solution**: Install GCC:
- Linux: `sudo apt-get install gcc`
- macOS: `xcode-select --install`
- Windows: Install MinGW-w64

### "cannot find package"

**Solution**: Download dependencies:
```bash
go mod download
go mod tidy
```

### "permission denied"

**Solution**: Make binary executable:
```bash
chmod +x geoffrussy
```

### "API key invalid"

**Solution**: 
1. Verify key format (OpenAI: `sk-`, Anthropic: `sk-ant-`)
2. Check key hasn't expired
3. Verify key has correct permissions
4. Test key with provider's API directly

### "database locked"

**Solution**:
1. Close other Geoffrey instances
2. Check file permissions: `chmod 644 ~/.geoffrussy/geoffrussy.db`
3. Delete and reinitialize if corrupted

### "rate limit exceeded"

**Solution**:
1. Check quota: `geoffrussy quota`
2. Wait for rate limit reset
3. Use different model/provider
4. Upgrade API plan

## Development Setup

### Install Development Tools

```bash
# golangci-lint for linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# gopter for property-based testing
go get github.com/leanovate/gopter

# Air for hot reload (optional)
go install github.com/cosmtrek/air@latest
```

### Run Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Property tests
make test-property

# Integration tests
make test-integration

# With coverage
go test -v -race -coverprofile=coverage.txt ./...
```

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint

# Run go vet
make vet
```

### Hot Reload (Optional)

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

## Next Steps

1. **Read the documentation**:
   - [README.md](../README.md) - Overview and quick start
   - [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture
   - [CONTRIBUTING.md](../CONTRIBUTING.md) - Contributing guidelines

2. **Try the tutorial**:
   ```bash
   geoffrussy init
   geoffrussy interview
   ```

3. **Join the community**:
   - GitHub Discussions
   - Issue Tracker
   - Community Chat (TBD)

## Getting Help

- 📖 [Documentation](.)
- 🐛 [Issue Tracker](https://github.com/yourusername/geoffrussy/issues)
- 💬 [Discussions](https://github.com/yourusername/geoffrussy/discussions)
- 📧 Email: support@geoffrussy.dev (TBD)
