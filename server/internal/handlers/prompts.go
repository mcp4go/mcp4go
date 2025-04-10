package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/server/iface"
)

// Handle prompts/list request
type ListPromptsHandler struct {
	prompt iface.IPrompt
}

// Create a new instance
func NewListPromptsHandler(prompt iface.IPrompt) *ListPromptsHandler {
	return &ListPromptsHandler{prompt: prompt}
}

// Handle prompts/list request
func (x *ListPromptsHandler) Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.ListPromptsRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	prompts, nextCursor, err := x.prompt.List(ctx, req.Cursor)
	if err != nil {
		return nil, fmt.Errorf("list prompts failed: %w", err)
	}

	result := protocol.ListPromptsResult{
		Prompts:    prompts,
		NextCursor: nextCursor,
	}

	return json.Marshal(result)
}

// Returns the result
func (x *ListPromptsHandler) Method() protocol.McpMethod {
	return protocol.MethodListPrompts
}

// Handle prompts/get request
type GetPromptHandler struct {
	prompt iface.IPrompt
}

// Create a new instance
func NewGetPromptHandler(prompt iface.IPrompt) *GetPromptHandler {
	return &GetPromptHandler{prompt: prompt}
}

// Handle prompts/get request
func (x *GetPromptHandler) Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.GetPromptRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	description, messages, err := x.prompt.Get(ctx, req.Name, req.Arguments)
	if err != nil {
		return nil, fmt.Errorf("get prompt failed: %w", err)
	}

	result := protocol.GetPromptResult{
		Description: description,
		Messages:    messages,
	}

	return json.Marshal(result)
}

// Returns the result
func (x *GetPromptHandler) Method() protocol.McpMethod {
	return protocol.MethodGetPrompt
}
