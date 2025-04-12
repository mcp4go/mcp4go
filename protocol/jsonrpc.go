package protocol

import (
	"encoding/json"
)

// JsonrpcRequest represents a request that expects a response
// JsonrpcRequest 表示一个期望响应的请求
type JsonrpcRequest JsonrpcPack

func NewJsonrpcRequest(id json.RawMessage, method McpMethod, params json.RawMessage) *JsonrpcRequest {
	return &JsonrpcRequest{
		Jsonrpc: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

// IsNotification checks if this request is a notification (which doesn't require a response)
// IsNotification 检查此请求是否为通知（不需要响应）
func (x *JsonrpcRequest) IsNotification() bool {
	return len(x.ID) == 0
}

// GetID returns the request ID or nil if it's a notification
// GetID 返回请求ID，如果是通知则返回nil
func (x *JsonrpcRequest) GetID() json.RawMessage {
	return x.ID
}

// JsonrpcResponse represents a successful (non-error) response to a request
// JsonrpcResponse 表示对请求的成功（非错误）响应
type JsonrpcResponse JsonrpcPack

// JsonrpcPack represents a JSON-RPC packet
// JsonrpcPack 表示一个 JSON-RPC 数据包
type JsonrpcPack struct {
	// JSON-RPC version (JSON-RPC 版本)
	Jsonrpc string `json:"jsonrpc"`
	// Communication ID, can be string or integer (通信ID，可以是字符串或整数)
	ID json.RawMessage `json:"id,omitempty"`
	// Method to be called (调用的方法) (only for request)
	Method McpMethod `json:"method,omitempty"`
	// Parameters for the method (方法的参数) (only for request)
	Params json.RawMessage `json:"params,omitempty"`
	// Result of the method call (方法调用的结果) (only for response)
	Result json.RawMessage `json:"result,omitempty"`
	// Error information, if any (错误信息，如果有的话) (only for response)
	Error *JsonrpcError `json:"error,omitempty"`
}

// NewJsonrpcResponse creates a new JSON-RPC response
// NewJsonrpcResponse 创建一个新的 JSON-RPC 响应
func NewJsonrpcResponse(id json.RawMessage, result json.RawMessage, errInfo *JsonrpcError) *JsonrpcResponse {
	return &JsonrpcResponse{
		Jsonrpc: "2.0",
		ID:      id,
		Result:  result,
		Error:   errInfo,
	}
}

// JsonrpcError represents an error response to a JSON-RPC request
// JsonrpcError 表示对 JSON-RPC 请求的错误响应
type JsonrpcError struct {
	// Error code (错误代码)
	Code int64 `json:"code"`
	// Short description of the error (错误的简短描述)
	Message string `json:"message"`
	// Additional information about the error (关于错误的附加信息)
	Data json.RawMessage `json:"data,omitempty"`
}

// JsonrpcNotification represents a notification which does not expect a response
// JsonrpcNotification 表示不需要响应的通知
type JsonrpcNotification JsonrpcPack

// NewJsonrpcNotification creates a new JSON-RPC notification
// NewJsonrpcNotification 创建一个新的 JSON-RPC 通知
func NewJsonrpcNotification(method McpMethod, params json.RawMessage) *JsonrpcNotification {
	return &JsonrpcNotification{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
	}
}
