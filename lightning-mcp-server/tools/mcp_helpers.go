// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

package tools

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolInputSchema keeps schema literals compact while targeting go-sdk.
type ToolInputSchema struct {
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties,omitempty"`
	Required   []string       `json:"required,omitempty"`
}

// requestArguments parses the JSON arguments from an MCP tool call request.
func requestArguments(request *mcp.CallToolRequest) map[string]any {
	if request == nil || request.Params == nil ||
		len(request.Params.Arguments) == 0 {
		return map[string]any{}
	}

	var args map[string]any
	if err := json.Unmarshal(request.Params.Arguments, &args); err != nil {
		return map[string]any{}
	}
	if args == nil {
		return map[string]any{}
	}
	return args
}

// marshalJSON marshals v to a JSON string. If marshaling fails it returns
// the error text so callers never produce silently broken output.
func marshalJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return `{"error": "failed to marshal JSON: ` + err.Error() + `"}`
	}
	return string(b)
}

// newToolResultJSON marshals v to JSON and wraps it in a CallToolResult.
func newToolResultJSON(v any) *mcp.CallToolResult {
	return newToolResultText(marshalJSON(v))
}

// newToolResultText wraps a plain text string in a CallToolResult.
func newToolResultText(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

// newToolResultError wraps an error message in a CallToolResult with
// the IsError flag set.
func newToolResultError(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
		IsError: true,
	}
}
