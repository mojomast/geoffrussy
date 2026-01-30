package mcp

import (
	"context"
	"fmt"
	"sync"
)

// ToolHandler is a function that executes a tool
type ToolHandler func(ctx context.Context, arguments map[string]interface{}) (*CallToolResult, error)

// ToolRegistry manages registered tools
type ToolRegistry struct {
	mu       sync.RWMutex
	tools    map[string]Tool
	handlers map[string]ToolHandler
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:    make(map[string]Tool),
		handlers: make(map[string]ToolHandler),
	}
}

// RegisterTool registers a new tool with its handler
func (r *ToolRegistry) RegisterTool(tool Tool, handler ToolHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("tool already registered: %s", tool.Name)
	}

	r.tools[tool.Name] = tool
	r.handlers[tool.Name] = handler
	return nil
}

// ListTools returns all registered tools
func (r *ToolRegistry) ListTools() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// CallTool executes a tool with the given arguments
func (r *ToolRegistry) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*CallToolResult, error) {
	r.mu.RLock()
	handler, exists := r.handlers[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return handler(ctx, arguments)
}

// Helper function to create text content
func TextContent(text string) Content {
	return Content{
		Type: "text",
		Text: text,
	}
}

// Helper function to create error content
func ErrorContent(message string) Content {
	return Content{
		Type: "text",
		Text: fmt.Sprintf("Error: %s", message),
	}
}

// Helper function to create a success result
func SuccessResult(text string) *CallToolResult {
	return &CallToolResult{
		Content: []Content{TextContent(text)},
		IsError: false,
	}
}

// Helper function to create an error result
func ErrorResult(message string) *CallToolResult {
	return &CallToolResult{
		Content: []Content{ErrorContent(message)},
		IsError: true,
	}
}

// Common input schema builders

// StringParam creates a string parameter schema
func StringParam(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": description,
	}
}

// NumberParam creates a number parameter schema
func NumberParam(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "number",
		"description": description,
	}
}

// BooleanParam creates a boolean parameter schema
func BooleanParam(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "boolean",
		"description": description,
	}
}

// ObjectParam creates an object parameter schema
func ObjectParam(description string, properties map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":        "object",
		"description": description,
		"properties":  properties,
	}
}

// ArrayParam creates an array parameter schema
func ArrayParam(description string, items map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": description,
		"items":       items,
	}
}

// CreateInputSchema creates a complete input schema
func CreateInputSchema(properties map[string]interface{}, required []string) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}
