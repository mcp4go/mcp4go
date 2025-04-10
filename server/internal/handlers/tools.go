package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/server/iface"
)

// Handle tools/list request
type ListToolsHandler struct {
	tool iface.ITool
}

// Create a new instance
func NewListToolsHandler(tool iface.ITool) *ListToolsHandler {
	return &ListToolsHandler{tool: tool}
}

// Handle tools/list request
func (x *ListToolsHandler) Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.ListToolsRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	tools, nextCursor, err := x.tool.List(ctx, req.Cursor)
	if err != nil {
		return nil, fmt.Errorf("list tools failed: %w", err)
	}

	result := protocol.ListToolsResult{
		Tools:      tools,
		NextCursor: nextCursor,
	}

	return json.Marshal(result)
}

// Returns the result
func (x *ListToolsHandler) Method() protocol.McpMethod {
	return protocol.MethodListTools
}

// Handle tools/call request
type CallToolHandler struct {
	tool iface.ITool
}

// Create a new instance
func NewCallToolHandler(tool iface.ITool) *CallToolHandler {
	return &CallToolHandler{tool: tool}
}

// Handle tools/call request
func (x *CallToolHandler) Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.CallToolRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	content, err := x.tool.Call(ctx, req.Name, req.Arguments)
	if err != nil {
		return nil, fmt.Errorf("call tool failed: %w", err)
	}
	result := protocol.CallToolResult{
		IsError: err != nil,
		Content: content,
	}

	return json.Marshal(result)
}

// Returns the result
func (x *CallToolHandler) Method() protocol.McpMethod {
	return protocol.MethodCallTool
}
