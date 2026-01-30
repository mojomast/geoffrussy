package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mojomast/geoffrussy/internal/config"
	"github.com/mojomast/geoffrussy/internal/mcp"
	"github.com/mojomast/geoffrussy/internal/state"
	"github.com/spf13/cobra"
)

var (
	mcpProjectPath string
)

var mcpCmd = &cobra.Command{
	Use:   "mcp-server",
	Short: "Start MCP server for AI agent integration",
	Long: `Start a Model Context Protocol (MCP) server that exposes Geoffrey's
capabilities as tools and resources for AI agents to autonomously build software.

The MCP server runs over stdio transport and follows the JSON-RPC 2.0 protocol.
It can be connected to by any MCP-compatible client such as Claude for Desktop.

Example configuration for Claude for Desktop (claude_desktop_config.json):
{
  "mcpServers": {
    "geoffrey": {
      "command": "geoffrussy",
      "args": ["mcp-server", "--project-path", "/path/to/project"]
    }
  }
}`,
	RunE: runMCPServer,
}

func init() {
	mcpCmd.Flags().StringVar(&mcpProjectPath, "project-path", "", "Project root path (defaults to current directory)")
}

func runMCPServer(cmd *cobra.Command, args []string) error {
	// Determine project path
	projectPath := mcpProjectPath
	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		projectPath = cwd
	}

	// Load configuration
	cfgMgr := config.NewManager()
	if err := cfgMgr.Load(nil); err != nil {
		// Config loading failure is not fatal for MCP server
		// We'll just continue without pre-configured providers
		fmt.Fprintf(os.Stderr, "Warning: Failed to load configuration: %v\n", err)
	}

	// Initialize database (create if doesn't exist)
	dbPath := filepath.Join(projectPath, ".geoffrussy", "state.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize state store: %w", err)
	}
	defer store.Close()

	// Create MCP server
	server := mcp.NewServer(mcp.ServerConfig{
		Name:    "geoffrey",
		Version: version,
		Store:   store,
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	})

	// Register tools
	toolHandlers := mcp.NewSimpleToolHandlers(cfgMgr, projectPath)
	if err := toolHandlers.RegisterBasicTools(server.GetToolRegistry()); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// Register resources
	resourceHandlers := mcp.NewResourceHandlers(cfgMgr, projectPath)
	if err := resourceHandlers.RegisterAllResources(server.GetResourceRegistry()); err != nil {
		return fmt.Errorf("failed to register resources: %w", err)
	}

	// Log startup to stderr (not stdout which is used for JSON-RPC)
	fmt.Fprintf(os.Stderr, "Geoffrey MCP Server v%s starting...\n", version)
	fmt.Fprintf(os.Stderr, "Project path: %s\n", projectPath)
	fmt.Fprintf(os.Stderr, "Listening for MCP requests on stdin/stdout\n")

	// Start server (blocks until stdin is closed)
	if err := server.Start(); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}

	return nil
}
