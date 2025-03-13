package iface

import (
	"context"

	"github.com/mcp4go/mcp4go/protocol"
)

// IClient defines the interface for MCP clients
type IClient interface {
	// Connect establishes a connection to the server and initializes it
	Connect(ctx context.Context) error

	// Close terminates the client connection
	Close() error

	// Information methods
	ServerInfo() protocol.Implementation
	ServerCapabilities() protocol.ServerCapabilities
	Instructions() string
	IsInitialized() bool

	// Tool methods
	ListTools(ctx context.Context) (protocol.ListToolsResult, error)
	CallTool(ctx context.Context, request protocol.CallToolRequest) (protocol.CallToolResult, error)

	// Resource methods
	ListResources(ctx context.Context) (protocol.ListResourcesResult, error)
	ReadResource(ctx context.Context, request protocol.ReadResourceRequest) (protocol.ReadResourceResult, error)
	SubscribeResource(ctx context.Context, request protocol.SubscribeRequest) error
	UnsubscribeResource(ctx context.Context, request protocol.UnsubscribeRequest) error

	// Prompt methods
	ListPrompts(ctx context.Context) (protocol.ListPromptsResult, error)
	GetPrompt(ctx context.Context, request protocol.GetPromptRequest) (protocol.GetPromptResult, error)
}
