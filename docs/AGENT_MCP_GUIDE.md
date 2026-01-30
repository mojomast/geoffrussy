# Geoffrey MCP Server - Agent Guide

This guide is written for AI agents to help them effectively use Geoffrey's MCP server to manage software development projects.

## Quick Reference

**Server Command:** `geoffrussy mcp-server --project-path /path/to/project`

**Available Tools:**
- `get_status` - Get project status
- `get_stats` - Get token usage statistics
- `list_phases` - List all development phases
- `create_checkpoint` - Create a checkpoint
- `list_checkpoints` - List all checkpoints

**Available Resources:**
- `project://status` - Project status (JSON)
- `project://architecture` - Architecture documentation (Markdown)
- `project://devplan` - Development plan (JSON)
- `project://phases` - All phases (JSON)
- `project://interview` - Interview requirements (JSON)
- `project://checkpoints` - All checkpoints (JSON)
- `project://stats` - Token usage statistics (JSON)

## Tool Usage Guide

### get_status

Get the current project status, including stage, progress, and active tasks.

**When to use:**
- Before taking any action to understand project state
- After completing tasks to verify progress
- To check if tasks are blocked or in progress

**Parameters:**
```json
{
  "projectPath": "/absolute/path/to/project"
}
```

**Returns:**
- `projectId` - Project identifier
- `projectName` - Human-readable name
- `currentStage` - One of: init, interview, design, plan, develop, complete
- `currentPhase` - Current phase ID if in development
- `completionPercentage` - Overall progress (0-100)
- Task counts (total, completed, inProgress, blocked)
- Phase counts (total, completed, inProgress, blocked)

**Example:**
```json
{
  "projectPath": "/home/user/myproject"
}
```

---

### get_stats

Get token usage and cost statistics for the project.

**When to use:**
- To monitor API costs
- To check which providers are being used
- To understand token consumption patterns

**Parameters:**
```json
{
  "projectPath": "/absolute/path/to/project"
}
```

**Returns:**
- Total cost
- Total input/output tokens
- Breakdown by provider (OpenAI, Anthropic, etc.)
- Breakdown by phase
- Last updated timestamp

---

### list_phases

List all development phases with their status and tasks.

**When to use:**
- To see what work is planned
- To identify the next phase to work on
- To understand dependencies between phases

**Parameters:**
```json
{
  "projectPath": "/absolute/path/to/project"
}
```

**Returns:**
Array of phases, each with:
- `phaseId` - Unique identifier
- `phaseNumber` - Order in development plan
- `title` - Human-readable title
- `status` - not_started, in_progress, completed, blocked
- `taskCount` - Number of tasks
- `createdAt`, `startedAt`, `completedAt` - Timestamps
- `content` - Full phase description

**Phase Status Icons:**
- ⏭️ Not started
- 🔄 In progress
- ✅ Completed
- 🚫 Blocked

---

### create_checkpoint

Create a checkpoint to save the current project state.

**When to use:**
- **Before** making significant changes (refactor, major feature)
- After completing a milestone
- Before risky operations
- To create restore points for experimentation

**Parameters:**
```json
{
  "projectPath": "/absolute/path/to/project",
  "name": "descriptive-checkpoint-name"
}
```

**Checkpoint Naming Best Practices:**
- Use kebab-case: `before-refactor`, `after-auth-fix`, `milestone-1-complete`
- Be descriptive: `pre-user-auth-implementation`, `post-api-redesign`
- Include context: `before-breaking-change-db-schema`, `after-adding-webhooks`

**Returns:**
- `checkpointId` - Unique identifier
- Confirmation message

**Example:**
```json
{
  "projectPath": "/home/user/myproject",
  "name": "before-refactoring-user-module"
}
```

---

### list_checkpoints

List all checkpoints for the project.

**When to use:**
- To see available restore points
- To find the right checkpoint to restore
- To verify a checkpoint was created

**Parameters:**
```json
{
  "projectPath": "/absolute/path/to/project"
}
```

**Returns:**
Array of checkpoints, each with:
- `checkpointId` - Unique identifier
- `name` - Human-readable name
- `gitTag` - Git tag created
- `createdAt` - Timestamp
- `metadata` - Additional info

---

## Resource Usage Guide

Resources are read-only data sources accessed via URIs. Use them to get detailed project information.

### project://status

Get complete project status in JSON format.

**When to use:**
- When you need structured status data
- For programmatic analysis of project state
- To get complete status without making a tool call

**Access:** Read the resource `project://status`

**Fields:**
```json
{
  "projectId": "myproject",
  "projectName": "My Project",
  "currentStage": "develop",
  "currentPhase": "phase-3",
  "completionPercentage": 45,
  "totalTasks": 50,
  "completedTasks": 22,
  "inProgressTasks": 3,
  "blockedTasks": 2,
  "totalPhases": 10,
  "completedPhases": 4,
  "inProgressPhases": 1,
  "blockedPhases": 0
}
```

---

### project://architecture

Get the complete system architecture documentation in Markdown format.

**When to use:**
- **Before** implementing features (understand the design)
- To understand component interactions
- To see data flows and API contracts
- To understand technical decisions and trade-offs

**When NOT available:**
- Before `geoffrussy design` has been run
- In early project stages (init, interview)

**Access:** Read the resource `project://architecture`

**Content includes:**
- System overview
- Component descriptions
- Data flows
- Technology rationale
- Scaling strategies
- API contracts
- Security approach

**Agent workflow:**
1. Read `project://status` to verify architecture exists
2. Read `project://architecture` to understand design
3. Implement features according to architecture
4. Read again after changes to ensure compliance

---

### project://devplan

Get the complete development plan with all phases and tasks in JSON format.

**When to use:**
- To see the complete roadmap
- To understand dependencies between phases
- To find specific tasks
- To verify what's been completed

**When NOT available:**
- Before `geoffrussy plan` has been run

**Access:** Read the resource `project://devplan`

**Structure:**
```json
{
  "projectId": "myproject",
  "totalPhases": 10,
  "phases": [
    {
      "phaseId": "phase-1",
      "phaseNumber": 1,
      "title": "Setup and Configuration",
      "status": "completed",
      "taskCount": 5,
      "tasks": [
        {
          "taskId": "task-1-1",
          "title": "Initialize project",
          "status": "completed",
          // ... more task fields
        }
      ]
    }
    // ... more phases
  ]
}
```

---

### project://phases

List of all development phases with status in JSON format.

**When to use:**
- To quickly see all phases and their status
- To find the next phase to work on
- To check which phases are blocked

**Access:** Read the resource `project://phases`

**Each phase includes:**
- Phase ID, number, and title
- Status (not_started, in_progress, completed, blocked)
- Created, started, and completed timestamps
- Full phase content

**Agent workflow:**
1. Read `project://phases`
2. Find first phase with status `in_progress`
3. If none, find first `not_started` phase
4. Read phase content for details

---

### project://interview

Get collected requirements from the interview process in JSON format.

**When to use:**
- To understand project requirements
- To see what features need to be built
- To understand constraints and assumptions
- To verify alignment with requirements

**Access:** Read the resource `project://interview`

**Includes:**
- Project name and problem statement
- Target users
- Success metrics
- Technical stack choices
- Integration requirements
- Scope definition
- Constraints and assumptions

---

### project://checkpoints

List of all saved checkpoints in JSON format.

**When to use:**
- To see available restore points
- To find the right checkpoint to restore
- To verify checkpoint creation

**Access:** Read the resource `project://checkpoints`

**Each checkpoint includes:**
- Checkpoint ID
- Name and Git tag
- Creation timestamp
- Metadata

---

### project://stats

Token usage and cost statistics in JSON format.

**When to use:**
- To monitor API costs
- To track token consumption
- To analyze usage patterns

**Access:** Read the resource `project://stats`

**Includes:**
- Total cost
- Total input/output tokens
- Breakdown by provider
- Breakdown by phase
- Last updated timestamp

---

## Workflows for AI Agents

### Workflow 1: Starting Work on a Project

**Goal:** Understand the project state before starting work.

**Steps:**
1. Call `get_status` to get current project state
2. Read `project://architecture` to understand the design (if available)
3. Read `project://devplan` to see the roadmap (if available)
4. Read `project://phases` to see phase status
5. Create a checkpoint with `create_checkpoint` before starting work

**Example:**
```json
// Step 1: Get status
{
  "name": "get_status",
  "arguments": {
    "projectPath": "/home/user/project"
  }
}

// Step 2: Read architecture
{
  "name": "resources/read",
  "arguments": {
    "uri": "project://architecture"
  }
}

// Step 3: Create checkpoint before work
{
  "name": "create_checkpoint",
  "arguments": {
    "projectPath": "/home/user/project",
    "name": "before-starting-new-feature"
  }
}
```

---

### Workflow 2: Monitoring Project Progress

**Goal:** Check progress and identify blockers.

**Steps:**
1. Call `get_status` to get current state
2. Call `list_phases` to see all phases
3. Read `project://stats` to check costs
4. If blocked tasks exist, investigate blockers

**Example:**
```json
// Check status
{
  "name": "get_status",
  "arguments": {
    "projectPath": "/home/user/project"
  }
}

// List phases to see what's blocked
{
  "name": "list_phases",
  "arguments": {
    "projectPath": "/home/user/project"
  }
}
```

---

### Workflow 3: Making Significant Changes

**Goal:** Safely make significant changes with rollback capability.

**Steps:**
1. Call `list_checkpoints` to see existing restore points
2. Create a checkpoint with descriptive name
3. Make changes
4. Call `get_status` to verify success
5. If failed, restore from checkpoint (manual operation via git)

**Example:**
```json
// List existing checkpoints
{
  "name": "list_checkpoints",
  "arguments": {
    "projectPath": "/home/user/project"
  }
}

// Create checkpoint before changes
{
  "name": "create_checkpoint",
  "arguments": {
    "projectPath": "/home/user/project",
    "name": "before-major-refactor-api-layer"
  }
}
```

---

### Workflow 4: Finding the Next Task

**Goal:** Find the next task to work on.

**Steps:**
1. Call `get_status` to see if there's a current phase
2. Call `list_phases` to find in-progress or next not-started phase
3. Read `project://devplan` for detailed task list
4. Start working on identified task

**Example:**
```json
// Get current status
{
  "name": "get_status",
  "arguments": {
    "projectPath": "/home/user/project"
  }
}

// List phases to find work
{
  "name": "list_phases",
  "arguments": {
    "projectPath": "/home/user/project"
  }
}

// Read dev plan for task details
{
  "name": "resources/read",
  "arguments": {
    "uri": "project://devplan"
  }
}
```

---

### Workflow 5: Cost Monitoring

**Goal:** Keep track of API costs and token usage.

**Steps:**
1. Call `get_stats` to get current statistics
2. Read `project://stats` for detailed breakdown
3. Analyze usage by provider and phase
4. Recommend cost optimizations if needed

**Example:**
```json
// Get statistics
{
  "name": "get_stats",
  "arguments": {
    "projectPath": "/home/user/project"
  }
}

// Read detailed stats
{
  "name": "resources/read",
  "arguments": {
    "uri": "project://stats"
  }
}
```

---

## Best Practices for AI Agents

### 1. Always Check Project State First
Before taking any action, call `get_status` to understand:
- Current stage (you can't use architecture if design hasn't run)
- Current progress (are you working on the right phase?)
- Blocked tasks (don't start work on blocked items)

### 2. Use Descriptive Checkpoint Names
Bad: `checkpoint-1`, `test`, `save`
Good: `before-refactor-auth`, `after-api-redesign`, `milestone-user-auth-complete`

### 3. Verify Resource Availability
Not all resources are available in all stages:
- `project://interview` - Available after interview stage
- `project://architecture` - Available after design stage
- `project://devplan` - Available after plan stage

Always check `project://status` first to verify stage.

### 4. Handle Errors Gracefully
If a tool or resource call fails:
- Read the error message carefully
- Check if required data exists (e.g., architecture not generated yet)
- Report the issue to the user with context
- Suggest next steps

### 5. Use Resources for Read Operations
For read-only data, prefer resources over tools:
- Use `project://status` instead of `get_status` when you don't need the tool's side effects
- Use `project://devplan` instead of parsing `list_phases` output

### 6. Create Checkpoints Before Risky Operations
Create checkpoints before:
- Major refactors
- Database schema changes
- API contract changes
- Removing or modifying critical components

### 7. Monitor Costs Regularly
Periodically check `project://stats` to:
- Track token usage
- Monitor API costs
- Identify expensive operations
- Optimize model selection

### 8. Understand Phase Dependencies
From `project://devplan`, understand that:
- Phases must be completed in order
- Some phases depend on previous phases
- Blocked phases cannot be started until dependencies are resolved

### 9. Report Progress Clearly
When working on tasks:
- Call `get_status` before and after to show progress
- Report completion percentage changes
- Highlight any blocked tasks discovered

### 10. Use Absolute Paths
All tool calls require absolute paths:
- Bad: `./project`, `~/project`, `project`
- Good: `/home/user/project`, `/absolute/path/to/project`

---

## Error Handling Guide

### Common Errors and Solutions

**Error:** "Project not initialized"
- **Cause:** Project hasn't been initialized with `geoffrussy init`
- **Solution:** Run `geoffrussy init` in the project directory

**Error:** "Architecture not found"
- **Cause:** `project://architecture` accessed before design stage
- **Solution:** Run `geoffrussy design` first

**Error:** "Dev plan not found"
- **Cause:** `project://devplan` accessed before plan stage
- **Solution:** Run `geoffrussy plan` first

**Error:** "Database not found"
- **Cause:** Project state database doesn't exist
- **Solution:** Ensure project has been initialized

**Error:** "Permission denied"
- **Cause:** Insufficient permissions on project directory
- **Solution:** Ensure read/write access to project and `.geoffrussy` directory

**Error:** "Invalid project path"
- **Cause:** Path doesn't exist or is not absolute
- **Solution:** Use absolute path and verify directory exists

---

## Performance Tips

1. **Batch operations:** Read multiple resources in sequence rather than alternating between tools
2. **Cache status:** Cache `get_status` results when possible to avoid redundant calls
3. **Use resources wisely:** Prefer resources for repeated reads of the same data
4. **Avoid polling:** Don't repeatedly call tools waiting for changes; instead, ask the user to notify when state changes

---

## Security Considerations

1. **Path validation:** All paths are validated to prevent directory traversal
2. **No authentication:** Currently, no authentication is implemented; run only in trusted environments
3. **Project isolation:** The MCP server only accesses project-specific data
4. **Stdio transport:** Using stdio is secure as it doesn't expose network ports

---

## Integration Examples

### Python Client Example

```python
import json
import subprocess
from typing import Dict, Any

class GeoffreyMCPClient:
    def __init__(self, project_path: str):
        self.project_path = project_path
        self.geoffrey_path = "/absolute/path/to/geoffrussy"
    
    def call_tool(self, tool_name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        """Call a Geoffrey MCP tool"""
        request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": tool_name,
                "arguments": arguments
            }
        }
        
        cmd = [self.geoffrey_path, "mcp-server", "--project-path", self.project_path]
        result = subprocess.run(
            cmd,
            input=json.dumps(request),
            capture_output=True,
            text=True
        )
        
        response = json.loads(result.stdout)
        return response.get("result", {})
    
    def read_resource(self, uri: str) -> Any:
        """Read a Geoffrey MCP resource"""
        request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "resources/read",
            "params": {
                "uri": uri
            }
        }
        
        cmd = [self.geoffrey_path, "mcp-server", "--project-path", self.project_path]
        result = subprocess.run(
            cmd,
            input=json.dumps(request),
            capture_output=True,
            text=True
        )
        
        response = json.loads(result.stdout)
        return response.get("result", {})
    
    def get_status(self) -> Dict[str, Any]:
        """Get project status"""
        return self.call_tool("get_status", {
            "projectPath": self.project_path
        })
    
    def create_checkpoint(self, name: str) -> Dict[str, Any]:
        """Create a checkpoint"""
        return self.call_tool("create_checkpoint", {
            "projectPath": self.project_path,
            "name": name
        })

# Usage
client = GeoffreyMCPClient("/home/user/myproject")
status = client.get_status()
print(f"Project: {status['projectName']}, Progress: {status['completionPercentage']}%")

client.create_checkpoint("before-making-changes")
```

---

## Quick Reference Card

```
GET PROJECT STATUS:
  Tool: get_status → projectPath
  Resource: project://status

UNDERSTAND DESIGN:
  Resource: project://architecture
  Resource: project://interview

SEE ROADMAP:
  Resource: project://devplan
  Resource: project://phases
  Tool: list_phases → projectPath

MONITOR COSTS:
  Tool: get_stats → projectPath
  Resource: project://stats

CHECKPOINTS:
  Tool: list_checkpoints → projectPath
  Tool: create_checkpoint → projectPath, name
  Resource: project://checkpoints

WORKFLOW:
  1. get_status → understand state
  2. create_checkpoint → save before work
  3. Read resources → understand context
  4. Do work → implement changes
  5. get_status → verify progress
```

---

## Support

For issues or questions:
1. Check the error message carefully
2. Verify project stage and available resources
3. Review this guide for common workflows
4. Check Geoffrey logs (stderr output)

---

## Appendix: Complete Tool and Resource List

### Tools (5 total)

| Tool | Parameters | Returns | Use Case |
|------|------------|---------|----------|
| get_status | projectPath | Status object | Check project state |
| get_stats | projectPath | Statistics object | Check costs/tokens |
| list_phases | projectPath | Array of phases | See all phases |
| create_checkpoint | projectPath, name | Confirmation | Save project state |
| list_checkpoints | projectPath | Array of checkpoints | List restore points |

### Resources (7 total)

| URI | Format | When Available | Use Case |
|-----|--------|----------------|----------|
| project://status | JSON | Always | Get project status |
| project://architecture | Markdown | After design | Understand system design |
| project://devplan | JSON | After plan | See development roadmap |
| project://phases | JSON | After plan | List all phases |
| project://interview | JSON | After interview | See requirements |
| project://checkpoints | JSON | Always | List checkpoints |
| project://stats | JSON | Always | Check costs/tokens |

---

**Version:** 1.0  
**Last Updated:** 2026-01-30  
**For Geoffrey MCP Server:** 2024-11-05 protocol version
