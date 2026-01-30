# MCP (Model Context Protocol) Integration

Geoffrey supports the Model Context Protocol (MCP), enabling AI agents to autonomously use Geoffrey's capabilities for building software projects.

## Overview

The MCP server exposes Geoffrey's functionality as:
- **Tools**: Actionable functions AI agents can call
- **Resources**: Project data and status information agents can query

This allows AI applications like Claude for Desktop, or custom MCP clients, to programmatically manage Geoffrey projects.

## Quick Start

### 1. Start the MCP Server

```bash
geoffrussy mcp-server --project-path /path/to/your/project
```

The server runs over stdio transport and follows the JSON-RPC 2.0 protocol specification.

### 2. Configure Claude for Desktop

Add Geoffrey to your Claude for Desktop configuration (`claude_desktop_config.json`):

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

**Note:** Use absolute paths for both the `command` and `--project-path`.

### 3. Restart Claude for Desktop

After saving the configuration, restart Claude for Desktop to load the MCP server.

## Available Tools

### 1. Project Management

- **`get_status`**: Get current project status including stage, progress, and active tasks.
- **`get_stats`**: Get token usage and cost statistics for the project.
- **`list_phases`**: List all development phases with their status and tasks.
- **`create_checkpoint`**: Create a checkpoint to save current project state.
- **`list_checkpoints`**: List all checkpoints for the project.

### 2. Interview & Requirements

- **`run_interview`**: Start or resume the project interview to gather requirements.
- **`submit_interview_answer`**: Submit answer to current interview question.

### 3. Design & Architecture

- **`generate_design`**: Generate system architecture from interview requirements.
- **`regenerate_design`**: Regenerate architecture with guidance.

### 4. Planning

- **`create_devplan`**: Generate development plan with phases and tasks.

### 5. Development Execution

- **`execute_phase`**: Execute all tasks in a development phase.
- **`execute_task`**: Execute a single task.
- **`get_task_output`**: Get detailed execution output for a task.
- **`handle_blocker`**: Attempt to resolve a blocker or get guidance.

## Available Resources

Resources provide read-only access to project data via URIs.

- **`project://status`**: Project status summary (JSON).
- **`project://current_question`**: Currently active interview question (JSON).
- **`project://interview`**: Full interview requirements data (JSON).
- **`project://architecture`**: Generated system architecture (Markdown).
- **`project://devplan`**: Complete development plan with phases (JSON).
- **`project://phases`**: List of all phases with status (JSON).
- **`project://task_details`**: Detailed task list with logs/output snippets (JSON).
- **`project://checkpoints`**: List of saved checkpoints (JSON).
- **`project://stats`**: Token usage and cost statistics (JSON).

## Use Cases for Autonomous Agents

### 1. Full Autonomous Development
An agent can:
1. Call `run_interview` and loop `submit_interview_answer` to gather requirements.
2. Call `generate_design` to create architecture.
3. Call `create_devplan` to plan tasks.
4. Loop through `execute_phase` to build the project.

### 2. Project Status Monitoring
An agent can periodically query `project://status` to monitor project progress and identify blockers.

### 3. Checkpoint Management
Before making significant changes, agents can use `create_checkpoint` to create a restore point, then use `list_checkpoints` to manage saved states.

### 4. Cost Optimization
Query `project://stats` to track token usage and costs, helping agents make informed decisions about model selection.

## Configuration Options

Add MCP settings to `~/.geoffrussy/config.yaml`:

```yaml
mcp:
  enabled: true
  log_level: info  # Options: debug, info, warn, error
  server_mode: stdio  # Currently only "stdio" is supported
```

### Debug Mode

Enable debug mode for troubleshooting:

```bash
geoffrussy mcp-server --project-path /path/to/project --debug
```

## Protocol Details

### Transport
Geoffrey's MCP server uses stdio transport, communicating via standard input/output using JSON-RPC 2.0 messages.

### Message Format
All messages follow JSON-RPC 2.0 specification.

## Logging

The MCP server writes logs to stderr to avoid corrupting JSON-RPC messages on stdout.

**Important:** Never write to stdout in the MCP server, as this will break the JSON-RPC protocol.

## Troubleshooting

### Resource Read Failures

1. Ensure the project has completed the relevant stages:
   - `project://architecture` requires `geoffrussy design` (or `generate_design`) to have run
   - `project://devplan` requires `geoffrussy plan` (or `create_devplan`) to have run
   - `project://interview` requires interview completion

### Permission Issues

Ensure the MCP server has read/write access to:
- Project directory
- `.geoffrussy` subdirectory
- Git repository (for checkpoint creation)

## Security Considerations

1. **Path Validation**: The MCP server validates all file paths to prevent directory traversal attacks
2. **Resource Access**: Only project-specific data is exposed; no system-wide access is provided
3. **Authentication**: Currently, the MCP server does not implement authentication. It should only be run in trusted environments.
4. **Stdio Transport**: The stdio transport is secure as it doesn't expose network ports

## Support

For issues or questions about MCP integration:
1. Check the troubleshooting section above
2. Review Geoffrey logs in stderr
3. Submit an issue on the Geoffrey repository

## License

Geoffrey's MCP integration is licensed under the same terms as Geoffrey itself.