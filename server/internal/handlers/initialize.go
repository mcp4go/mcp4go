package handlers

import (
	"context"
	"encoding/json"

	"github.com/mcp4go/mcp4go/protocol"
)

// InitializeHandler 处理initialize请求
type InitializeHandler struct {
	serverCapabilities protocol.ServerCapabilities
	serverInfo         protocol.Implementation
	instructions       string
}

// NewInitializeHandler 创建一个InitializeHandler实例
func NewInitializeHandler(
	serverCapabilities protocol.ServerCapabilities,
	serverInfo protocol.Implementation,
	instructions string,
) *InitializeHandler {
	//nolint:whitespace
	return &InitializeHandler{
		serverCapabilities: serverCapabilities,
		serverInfo:         serverInfo,
		instructions:       instructions,
	}
}

// Handle 处理initialize请求
func (x *InitializeHandler) Handle(_ context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.InitializeRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	// 构建初始化响应
	response := protocol.InitializeResult{
		ProtocolVersion: "2024-11-05", // 使用MCP协议版本
		Capabilities:    x.serverCapabilities,
		ServerInfo:      x.serverInfo,
		Instructions:    x.instructions,
	}

	return json.Marshal(response)
}

// Method 返回此处理程序对应的MCP方法
func (x *InitializeHandler) Method() protocol.McpMethod {
	return protocol.MethodInitialize
}
