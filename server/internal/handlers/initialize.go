package handlers

import (
	"context"
	"encoding/json"

	"github.com/mcp4go/mcp4go/protocol"
)

// InitializeHandler handles initialize requests
type InitializeHandler struct {
	serverCapabilities protocol.ServerCapabilities
	serverInfo         protocol.Implementation
	instructions       string
}

// NewInitializeHandler creates a new InitializeHandler instance
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

// Handle processes initialize requests
func (x *InitializeHandler) Handle(_ context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.InitializeRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	// Build initialization response
	response := protocol.InitializeResult{
		ProtocolVersion: "2024-11-05", // Using MCP protocol version
		Capabilities:    x.serverCapabilities,
		ServerInfo:      x.serverInfo,
		Instructions:    x.instructions,
	}

	return json.Marshal(response)
}

// Method returns the handler's method
func (x *InitializeHandler) Method() protocol.McpMethod {
	return protocol.MethodInitialize
}
