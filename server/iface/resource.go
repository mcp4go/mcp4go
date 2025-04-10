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

// IPrompt defines the prompt functionality interface
type IPrompt interface {
	// List returns available prompts
	List(ctx context.Context, cursor string) ([]protocol.Prompt, string, error)

	// Get returns a specific prompt with the given name and arguments
	Get(ctx context.Context, name string, arguments map[string]string) (string, []protocol.PromptMessage, error)

	StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.PromptListChangedNotification) error
}

type IResourceBuilder interface {
	Build() IResource
	Subscribe() bool
	ListChanged() bool
}

// IResource defines the resource interface
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

// ITool defines the tool functionality interface
type ITool interface {
	// List returns available tools
	List(ctx context.Context, cursor string) ([]protocol.Tool, string, error)

	// Call invokes the specified tool operation
	Call(ctx context.Context, name string, argsJSON json.RawMessage) ([]protocol.Content, error)

	StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.ToolListChangedNotification) error
}
