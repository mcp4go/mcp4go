package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/server/iface"
)

// ListResourcesHandler 处理resources/list请求
type ListResourcesHandler struct {
	resource iface.IResource
}

// NewListResourcesHandler 创建一个ListResourcesHandler实例
func NewListResourcesHandler(resource iface.IResource) *ListResourcesHandler {
	return &ListResourcesHandler{resource: resource}
}

// Handle 处理resources/list请求
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

// Method 返回此处理程序对应的MCP方法
func (x *ListResourcesHandler) Method() protocol.McpMethod {
	return protocol.MethodListResources
}

// ReadResourceHandler 处理resources/read请求
type ReadResourceHandler struct {
	resource iface.IResource
}

// NewReadResourceHandler 创建一个ReadResourceHandler实例
func NewReadResourceHandler(resource iface.IResource) *ReadResourceHandler {
	return &ReadResourceHandler{
		resource: resource,
	}
}

// Handle 处理resources/read请求
func (x *ReadResourceHandler) Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.ReadResourceRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	// 根据URI读取相应的资源内容
	contents, err := x.resource.Query(ctx, req.URI)
	if err != nil {
		return nil, fmt.Errorf("query resources failed: %w", err)
	}

	result := protocol.ReadResourceResult{
		Contents: contents,
	}

	return json.Marshal(result)
}

// Method 返回此处理程序对应的MCP方法
func (x *ReadResourceHandler) Method() protocol.McpMethod {
	return protocol.MethodReadResource
}

// ListResourceTemplatesHandler 处理resources/templates/list请求
type ListResourceTemplatesHandler struct {
	resource iface.IResource
}

// NewListResourceTemplatesHandler 创建一个ListResourceTemplatesHandler实例
func NewListResourceTemplatesHandler(resource iface.IResource) *ListResourceTemplatesHandler {
	return &ListResourceTemplatesHandler{
		resource: resource,
	}
}

// Handle 处理resources/templates/list请求
func (x *ListResourceTemplatesHandler) Handle(_ context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.ListResourceTemplatesRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	// 示例资源模板列表，在实际应用中应该从服务或数据库获取
	templates := x.resource.AccessList()

	result := protocol.ListResourceTemplatesResult{
		ResourceTemplates: templates,
	}

	return json.Marshal(result)
}

// Method 返回此处理程序对应的MCP方法
func (x *ListResourceTemplatesHandler) Method() protocol.McpMethod {
	return protocol.MethodListResourceTemplates
}

// SubscribeHandler 处理resources/subscribe请求
type SubscribeHandler struct {
	resource iface.IResource
	ch       chan<- protocol.ResourceUpdatedNotification
}

// NewSubscribeHandler 创建一个SubscribeHandler实例
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

// Handle 处理resources/subscribe请求
func (x *SubscribeHandler) Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.SubscribeRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	// 这里应该实现实际的订阅逻辑，例如将URI添加到订阅列表中
	// 在本示例中，我们只是返回一个成功响应
	err = x.resource.Watch(ctx, req.URI, x.ch)
	if err != nil {
		return nil, fmt.Errorf("subscribe failed: %w", err)
	}
	result := protocol.SubscribeResult{}
	return json.Marshal(result)
}

// Method 返回此处理程序对应的MCP方法
func (x *SubscribeHandler) Method() protocol.McpMethod {
	return protocol.MethodSubscribe
}

// UnsubscribeHandler 处理resources/unsubscribe请求
type UnsubscribeHandler struct {
	resource iface.IResource
}

// NewUnsubscribeHandler 创建一个UnsubscribeHandler实例
func NewUnsubscribeHandler(resource iface.IResource) *UnsubscribeHandler {
	return &UnsubscribeHandler{resource: resource}
}

// Handle 处理resources/unsubscribe请求
func (x *UnsubscribeHandler) Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error) {
	var req protocol.UnsubscribeRequest
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}

	// 这里应该实现实际的取消订阅逻辑，例如从订阅列表中移除URI
	// 在本示例中，我们只是返回一个成功响应
	err = x.resource.CloseWatch(ctx, req.URI)
	if err != nil {
		return nil, fmt.Errorf("unsubscribe failed: %w", err)
	}

	result := protocol.UnsubscribeResult{}
	return json.Marshal(result)
}

// Method 返回此处理程序对应的MCP方法
func (x *UnsubscribeHandler) Method() protocol.McpMethod {
	return protocol.MethodUnsubscribe
}
