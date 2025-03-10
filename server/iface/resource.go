package iface

import (
	"context"
	"encoding/json"

	"github.com/mcp4go/mcp4go/protocol"
)

type IPromptBuilder interface {
	Build() IPrompt
	ListChanged() bool
}

// IPrompt 定义了提示功能的接口
type IPrompt interface {
	// List 返回可用的提示列表
	List(ctx context.Context, cursor string) ([]protocol.Prompt, string, error)

	// Get 根据提示名称和参数获取提示内容
	Get(ctx context.Context, name string, arguments map[string]string) (string, []protocol.PromptMessage, error)

	StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.PromptListChangedNotification) error
}

type IResourceBuilder interface {
	Build() IResource
	Subscribe() bool
	ListChanged() bool
}

// IResource 定义了资源接口
type IResource interface {
	AccessList() []protocol.ResourceTemplate

	List(ctx context.Context, cursor string) ([]protocol.Resource, string, error)
	Query(ctx context.Context, uri string) ([]protocol.ResourceContent, error)

	Watch(ctx context.Context, uri string, ch chan<- protocol.ResourceUpdatedNotification) error
	CloseWatch(ctx context.Context, uri string) error

	StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.ResourceListChangedNotification) error
}

type IToolBuilder interface {
	Build() ITool
	ListChanged() bool
}

// ITool 定义了工具功能的接口
type ITool interface {
	// List 返回可用的工具列表
	List(ctx context.Context, cursor string) ([]protocol.Tool, string, error)

	// Call 调用指定工具执行操作
	Call(ctx context.Context, name string, argsJSON json.RawMessage) ([]protocol.Content, error)

	StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.ToolListChangedNotification) error
}
