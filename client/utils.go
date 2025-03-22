package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mcp4go/mcp4go/protocol"
)

// ToolHelper provides convenient methods for calling tools
type ToolHelper struct {
	client *Client
}

// NewToolHelper creates a helper for working with tools
func NewToolHelper(client *Client) *ToolHelper {
	return &ToolHelper{
		client: client,
	}
}

// CallWithJSON executes a tool with JSON-encoded arguments
func (h *ToolHelper) CallWithJSON(ctx context.Context, name string, args interface{}) (protocol.CallToolResult, error) {
	// Marshal arguments to JSON
	jsonBytes, err := json.Marshal(args)
	if err != nil {
		return protocol.CallToolResult{}, fmt.Errorf("failed to marshal arguments: %w", err)
	}

	// Call the tool
	request := protocol.CallToolRequest{
		Name:      name,
		Arguments: jsonBytes,
	}

	return h.client.CallTool(ctx, request)
}

// CallWithTimeout executes a tool with a timeout
func (h *ToolHelper) CallWithTimeout(name string, args interface{}, timeout time.Duration) (protocol.CallToolResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Marshal arguments to JSON
	jsonBytes, err := json.Marshal(args)
	if err != nil {
		return protocol.CallToolResult{}, fmt.Errorf("failed to marshal arguments: %w", err)
	}

	request := protocol.CallToolRequest{
		Name:      name,
		Arguments: jsonBytes,
	}

	return h.client.CallTool(ctx, request)
}

// ResourceHelper provides convenient methods for working with resources
type ResourceHelper struct {
	client *Client
}

// NewResourceHelper creates a helper for working with resources
func NewResourceHelper(client *Client) *ResourceHelper {
	return &ResourceHelper{
		client: client,
	}
}

// ReadByURI reads a resource by its URI
func (h *ResourceHelper) ReadByURI(ctx context.Context, uri string) (protocol.ReadResourceResult, error) {
	request := protocol.ReadResourceRequest{
		URI: uri,
	}

	return h.client.ReadResource(ctx, request)
}

// ReadTextContent reads and extracts the text content of a resource
func (h *ResourceHelper) ReadTextContent(ctx context.Context, uri string) (string, error) {
	result, err := h.ReadByURI(ctx, uri)
	if err != nil {
		return "", err
	}

	// Find the content
	for _, content := range result.Contents {
		if content.URI == uri && content.Text != "" {
			return content.Text, nil
		}
	}

	return "", fmt.Errorf("no text content found for resource: %s", uri)
}

// PromptHelper provides convenient methods for working with prompts
type PromptHelper struct {
	client *Client
}

// NewPromptHelper creates a helper for working with prompts
func NewPromptHelper(client *Client) *PromptHelper {
	return &PromptHelper{
		client: client,
	}
}

// GetPromptMessages extracts messages from a prompt result
func (h *PromptHelper) GetPromptMessages(ctx context.Context, name string, args map[string]string) ([]protocol.PromptMessage, error) {
	result, err := h.client.GetPrompt(ctx, protocol.GetPromptRequest{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		return nil, err
	}

	return result.Messages, nil
}
