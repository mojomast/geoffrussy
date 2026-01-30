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

### get_status
Get current project status including stage, progress, and active tasks.

**Parameters:**
- `projectPath` (required): Absolute path to the project directory

**Returns:** Project status summary with completion percentage, tasks, and phases information.

**Example:**
```json
{
  "projectPath": "/path/to/project"
}
```

### get_stats
Get token usage and cost statistics for the project.

**Parameters:**
- `projectPath` (required): Absolute path to the project directory

**Returns:** Token usage statistics by provider and phase, with associated costs.

### list_phases
List all development phases with their status and tasks.

**Parameters:**
- `projectPath` (required): Absolute path to the project directory

**Returns:** List of all phases with status icons, task counts, and completion status.

### create_checkpoint
Create a checkpoint to save current project state.

**Parameters:**
- `projectPath` (required): Absolute path to the project directory
- `name` (required): Name for the checkpoint

**Returns:** Confirmation with checkpoint ID.

**Example:**
```json
{
  "projectPath": "/path/to/project",
  "name": "before-refactoring"
}
```

### list_checkpoints
List all checkpoints for the project.

**Parameters:**
- `projectPath` (required): Absolute path to the project directory

**Returns:** List of all checkpoints with creation timestamps and IDs.

## Available Resources

Resources provide read-only access to project data via URIs.

### project://status
Current project status, stage, and progress information in JSON format.

**Fields:**
- `projectId`: Project identifier
- `projectName`: Human-readable project name
- `currentStage`: Current pipeline stage (init, interview, design, plan, develop, complete)
- `currentPhase`: Current phase ID if in development
- `completionPercentage`: Overall completion percentage
- `totalTasks`, `completedTasks`, `inProgressTasks`, `blockedTasks`: Task counters
- `totalPhases`, `completedPhases`, `inProgressPhases`, `blockedPhases`: Phase counters

### project://architecture
Generated system architecture documentation in Markdown format.

Contains the complete architecture document including:
- System overview
- Component descriptions
- Data flows
- Technology rationale
- Scaling strategies
- API contracts
- Security approach

### project://devplan
Complete development plan with all phases and tasks in JSON format.

**Structure:**
```json
{
  "projectId": "project-name",
  "totalPhases": 10,
  "phases": [...]
}
```

### project://phases
List of all development phases with status in JSON format.

Each phase includes:
- Phase ID, number, and title
- Status (not_started, in_progress, completed, blocked)
- Created, started, and completed timestamps
- Full phase content

### project://interview
Collected requirements from the interview process in JSON format.

Includes:
- Project name and problem statement
- Target users
- Success metrics
- Technical stack choices
- Integration requirements
- Scope definition
- Constraints and assumptions

### project://checkpoints
List of all saved checkpoints in JSON format.

Each checkpoint includes:
- Checkpoint ID
- Name and Git tag
- Creation timestamp
- Metadata

### project://stats
Token usage and cost statistics in JSON format.

Includes:
- Total cost
- Total input/output tokens
- Breakdown by provider
- Breakdown by phase
- Last updated timestamp

## Use Cases for Autonomous Agents

### 1. Project Status Monitoring
An agent can periodically query `project://status` to monitor project progress and identify blockers.

### 2. Checkpoint Management
Before making significant changes, agents can use `create_checkpoint` to create a restore point, then use `list_checkpoints` to manage saved states.

### 3. Phase Progress Tracking
Use `list_phases` to see which phases are complete, in progress, or blocked, enabling the agent to focus on the current work.

### 4. Cost Optimization
Query `project://stats` to track token usage and costs, helping agents make informed decisions about model selection.

### 5. Architecture Review
Access `project://architecture` to understand the system design before implementing features.

## Configuration Options

Add MCP settings to `~/.geoffrussy/config.yaml`:

```yaml
mcp:
  enabled: true
  log_level: info
  server_mode: stdio  # Currently only stdio is supported
```

### Configuration Fields

- `enabled`: Enable/disable MCP server functionality (default: true)
- `log_level`: Logging level for MCP server (debug, info, warn, error)
- `server_mode`: Transport mode, currently only "stdio" is supported

## Protocol Details

### Transport
Geoffrey's MCP server uses stdio transport, communicating via standard input/output using JSON-RPC 2.0 messages.

### Message Format
All messages follow JSON-RPC 2.0 specification:

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "get_status",
    "arguments": {
      "projectPath": "/path/to/project"
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Project status..."
      }
    ]
  }
}
```

### Supported Methods

- `initialize`: Initialize MCP connection
- `tools/list`: List available tools
- `tools/call`: Execute a tool
- `resources/list`: List available resources
- `resources/read`: Read a resource
- `ping`: Health check

## Logging

The MCP server writes logs to stderr to avoid corrupting JSON-RPC messages on stdout. This includes:
- Server startup messages
- Error messages
- Warning messages

**Important:** Never write to stdout in the MCP server, as this will break the JSON-RPC protocol.

## Troubleshooting

### Server Not Starting

1. Check that the `geoffrussy` binary is accessible
2. Verify the project path exists
3. Ensure the project has been initialized with `geoffrussy init`

### Tools Not Appearing in Claude

1. Verify the configuration path is correct
2. Check that the `command` path is absolute
3. Restart Claude for Desktop after configuration changes
4. Check stderr logs for error messages

### Resource Read Failures

1. Ensure the project has completed the relevant stages:
   - `project://architecture` requires `geoffrussy design` to have run
   - `project://devplan` requires `geoffrussy plan` to have run
   - `project://interview` requires `geoffrussy interview` to have been completed

2. Verify the database file exists at `<project>/.geoffrussy/state.db`

### Permission Issues

Ensure the MCP server has read/write access to:
- Project directory
- `.geoffrussy` subdirectory
- Git repository (for checkpoint creation)

## Advanced Usage

### Custom MCP Clients

You can build custom MCP clients that connect to Geoffrey's MCP server. Use any MCP client library that supports stdio transport.

Example in Python using a hypothetical MCP library:
```python
from mcp_client import MCPClient

client = MCPClient.from_stdio(
    command=["geoffrussy", "mcp-server", "--project-path", "/path/to/project"]
)

# Initialize connection
client.initialize()

# Call a tool
result = client.call_tool("get_status", {
    "projectPath": "/path/to/project"
})

print(result.content[0].text)
```

### Integration with CI/CD

Geoffrey's MCP server can be integrated into CI/CD pipelines to automate project status reporting:

```bash
#!/bin/bash
# Start MCP server and query status
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_status","arguments":{"projectPath":"'$(pwd)'"}}}' | \
  geoffrussy mcp-server --project-path $(pwd)
```

## Security Considerations

1. **Path Validation**: The MCP server validates all file paths to prevent directory traversal attacks
2. **Resource Access**: Only project-specific data is exposed; no system-wide access is provided
3. **Authentication**: Currently, the MCP server does not implement authentication. It should only be run in trusted environments.
4. **Stdio Transport**: The stdio transport is secure as it doesn't expose network ports

## Future Enhancements

Planned additions to Geoffrey's MCP support:

- [ ] Additional tools for interview submission and architecture generation
- [ ] WebSocket transport for remote access
- [ ] Authentication and authorization
- [ ] Rate limiting and quota management
- [ ] Streaming tool responses for long-running operations
- [ ] Prompt templates for common workflows
- [ ] Resource subscriptions for real-time updates

## References

- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
- [Geoffrey Documentation](../README.md)

## Support

For issues or questions about MCP integration:
1. Check the troubleshooting section above
2. Review Geoffrey logs in stderr
3. Submit an issue on the Geoffrey repository

## License

Geoffrey's MCP integration is licensed under the same terms as Geoffrey itself.
