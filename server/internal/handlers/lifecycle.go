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

// ShutdownHandler handles shutdown requests
type ShutdownHandler struct{}

// NewShutdownHandler creates a new ShutdownHandler instance
func NewShutdownHandler() *ShutdownHandler {
	return &ShutdownHandler{}
}

// Handle processes shutdown requests
func (x *ShutdownHandler) Handle(_ context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.ShutdownRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	// Actual cleanup work should be done here
	// In this example, we just return a success response

	result := protocol.ShutdownResult{}
	return json.Marshal(result)
}

// Method returns the handler's method
func (x *ShutdownHandler) Method() protocol.McpMethod {
	return "shutdown" // Note: This method is not defined in constants.go, can be added as needed
}

// CancelHandler handles cancel requests
type CancelHandler struct{}

// NewCancelHandler creates a new CancelHandler instance
func NewCancelHandler() *CancelHandler {
	return &CancelHandler{}
}

// Handle processes cancel requests
func (x *CancelHandler) Handle(_ context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.CancelRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	// Logic to cancel requests with the specified ID should be implemented here
	// In this example, we just return a success response

	result := protocol.CancelResult{}
	return json.Marshal(result)
}

// Method returns the handler's method
func (x *CancelHandler) Method() protocol.McpMethod {
	return "$/cancelRequest" // Note: This method is not defined in constants.go, can be added as needed
}

// ProgressNotificationHandler sends progress notifications
type ProgressNotificationHandler struct {
	gateway interface {
		SendNotification(method string, params interface{}) error
	}
}

// NewProgressNotificationHandler creates a new ProgressNotificationHandler instance
func NewProgressNotificationHandler(gateway interface {
	SendNotification(method string, params interface{}) error
},
) *ProgressNotificationHandler {
	//nolint:whitespace
	return &ProgressNotificationHandler{
		gateway: gateway,
	}
}

// SendProgress sends a progress notification
func (x *ProgressNotificationHandler) SendProgress(token protocol.ProgressToken, value int, message string) error {
	params := protocol.ProgressParams{
		Token:   token,
		Value:   value,
		Message: message,
	}

	return x.gateway.SendNotification(protocol.NotificationProgress, params)
}
