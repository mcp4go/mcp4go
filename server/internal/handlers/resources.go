package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/server/iface"
)

// Handle resources/list request
type ListResourcesHandler struct {
	resource iface.IResource
}

// Create a new instance
func NewListResourcesHandler(resource iface.IResource) *ListResourcesHandler {
	return &ListResourcesHandler{resource: resource}
}

// Handle resources/list request
func (x *ListResourcesHandler) Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.ListResourcesRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	resources, nextCursor, err := x.resource.List(ctx, req.Cursor)
	if err != nil {
		return nil, fmt.Errorf("list resources failed: %w", err)
	}

	result := protocol.ListResourcesResult{
		Resources:  resources,
		NextCursor: nextCursor,
		Meta:       nil,
	}

	return json.Marshal(result)
}

// Returns the result
func (x *ListResourcesHandler) Method() protocol.McpMethod {
	return protocol.MethodListResources
}

// Handle resources/read request
type ReadResourceHandler struct {
	resource iface.IResource
}

// Create a new instance
func NewReadResourceHandler(resource iface.IResource) *ReadResourceHandler {
	return &ReadResourceHandler{
		resource: resource,
	}
}

// Handle resources/read request
func (x *ReadResourceHandler) Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.ReadResourceRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	// Read the resource content based on the URI
	contents, err := x.resource.Query(ctx, req.URI)
	if err != nil {
		return nil, fmt.Errorf("query resources failed: %w", err)
	}

	result := protocol.ReadResourceResult{
		Contents: contents,
	}

	return json.Marshal(result)
}

// Returns the result
func (x *ReadResourceHandler) Method() protocol.McpMethod {
	return protocol.MethodReadResource
}

// Handle resources/templates/list request
type ListResourceTemplatesHandler struct {
	resource iface.IResource
}

// Create a new instance
func NewListResourceTemplatesHandler(resource iface.IResource) *ListResourceTemplatesHandler {
	return &ListResourceTemplatesHandler{
		resource: resource,
	}
}

// Handle resources/templates/list request
func (x *ListResourceTemplatesHandler) Handle(_ context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.ListResourceTemplatesRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	// Example resource template list, in real applications this should be retrieved from a service or database
	templates := x.resource.AccessList()

	result := protocol.ListResourceTemplatesResult{
		ResourceTemplates: templates,
	}

	return json.Marshal(result)
}

// Returns the result
func (x *ListResourceTemplatesHandler) Method() protocol.McpMethod {
	return protocol.MethodListResourceTemplates
}

// Handle resources/subscribe request
type SubscribeHandler struct {
	resource iface.IResource
	ch       chan<- protocol.ResourceUpdatedNotification
}

// Create a new instance
func NewSubscribeHandler(
	resource iface.IResource,
	bus iface.EventBus,
) *SubscribeHandler {
	//nolint:whitespace
	return &SubscribeHandler{
		resource: resource,
		ch:       bus.ResourceUpdatedNotificationChan,
	}
}

// Handle resources/subscribe request
func (x *SubscribeHandler) Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.SubscribeRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	// Here should implement the actual subscription logic, such as adding the URI to the subscription list
	// In this example, we just return a success response
	err = x.resource.Watch(ctx, req.URI, x.ch)
	if err != nil {
		return nil, fmt.Errorf("subscribe failed: %w", err)
	}
	result := protocol.SubscribeResult{}
	return json.Marshal(result)
}

// Returns the result
func (x *SubscribeHandler) Method() protocol.McpMethod {
	return protocol.MethodSubscribe
}

// Handle resources/unsubscribe request
type UnsubscribeHandler struct {
	resource iface.IResource
}

// Create a new instance
func NewUnsubscribeHandler(resource iface.IResource) *UnsubscribeHandler {
	return &UnsubscribeHandler{resource: resource}
}

// Handle resources/unsubscribe request
func (x *UnsubscribeHandler) Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.UnsubscribeRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	// Here should implement the actual unsubscribe logic, such as removing the URI from the subscription list
	// In this example, we just return a success response
	err = x.resource.CloseWatch(ctx, req.URI)
	if err != nil {
		return nil, fmt.Errorf("unsubscribe failed: %w", err)
	}

	result := protocol.UnsubscribeResult{}
	return json.Marshal(result)
}

// Returns the result
func (x *UnsubscribeHandler) Method() protocol.McpMethod {
	return protocol.MethodUnsubscribe
}
