package handlers

import (
	"context"
	"encoding/json"

	"github.com/mcp4go/mcp4go/protocol"
)

// InitializedHandler handles initialized notifications
type InitializedHandler struct{}

// NewInitializedHandler creates a new InitializedHandler instance
func NewInitializedHandler() *InitializedHandler {
	return &InitializedHandler{}
}

// Handle processes initialized notifications
func (x *InitializedHandler) Handle(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
	// Initialization work can be done here
	return nil, nil
}

// Method returns the handler's method
func (x *InitializedHandler) Method() protocol.McpMethod {
	return protocol.NotificationInitialized
}
