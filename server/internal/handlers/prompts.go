package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/server/iface"
)

// ListPromptsHandler 处理prompts/list请求
type ListPromptsHandler struct {
	prompt iface.IPrompt
}

// NewListPromptsHandler 创建一个ListPromptsHandler实例
func NewListPromptsHandler(prompt iface.IPrompt) *ListPromptsHandler {
	return &ListPromptsHandler{prompt: prompt}
}

// Handle 处理prompts/list请求
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

// Method 返回此处理程序对应的MCP方法
func (x *ListPromptsHandler) Method() protocol.McpMethod {
	return protocol.MethodListPrompts
}

// GetPromptHandler 处理prompts/get请求
type GetPromptHandler struct {
	prompt iface.IPrompt
}

// NewGetPromptHandler 创建一个GetPromptHandler实例
func NewGetPromptHandler(prompt iface.IPrompt) *GetPromptHandler {
	return &GetPromptHandler{prompt: prompt}
}

// Handle 处理prompts/get请求
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

// Method 返回此处理程序对应的MCP方法
func (x *GetPromptHandler) Method() protocol.McpMethod {
	return protocol.MethodGetPrompt
}
