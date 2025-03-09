package server

import (
	"context"

	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/server/iface"
)

// defaultOptions returns a new options struct with default values
func defaultOptions() options {
	return options{
		serverInfo: protocol.Implementation{
			Name:    "mcp4go",
			Version: "0.1.0",
		},
		instructions:    "Welcome to mcp4go!",
		resourceBuilder: &dummyIResourceBuilder{},
		promptBuilder:   &dummyIPromptBuilder{},
		toolBuilder:     &dummyIToolBuilder{},
	}
}

// ServerOption is a function that configures the server options
type ServerOption interface {
	apply(*options)
}

type ServerOptionFunc func(*options)

func (fn ServerOptionFunc) apply(o *options) {
	fn(o)
}

// WithServerInfo sets the server implementation info
func WithServerInfo(info protocol.Implementation) ServerOptionFunc {
	return func(o *options) {
		o.serverInfo = info
	}
}

// WithInstructions sets the server instructions
func WithInstructions(instructions string) ServerOptionFunc {
	return func(o *options) {
		o.instructions = instructions
	}
}

// WithResourceBuilder sets the resource builder interface
func WithResourceBuilder(resource iface.IResourceBuilder) ServerOptionFunc {
	return func(o *options) {
		o.resourceBuilder = resource
	}
}

// WithPromptBuilder sets the prompt builder interface
func WithPromptBuilder(prompt iface.IPromptBuilder) ServerOptionFunc {
	return func(o *options) {
		o.promptBuilder = prompt
	}
}

// WithToolBuilder sets the tool builder interface
func WithToolBuilder(tool iface.IToolBuilder) ServerOptionFunc {
	return func(o *options) {
		o.toolBuilder = tool
	}
}

type dummyIResourceBuilder struct{}

func (x *dummyIResourceBuilder) Build() iface.IResource {
	return &dummyIResource{}
}

func (x *dummyIResourceBuilder) Subscribe() bool {
	return false
}

func (x *dummyIResourceBuilder) ListChanged() bool {
	return false
}

type dummyIResource struct{}

func (x *dummyIResource) AccessList() []protocol.ResourceTemplate {
	return nil
}

func (x *dummyIResource) List(ctx context.Context, cursor string) ([]protocol.Resource, string, error) {
	return nil, "", nil
}

func (x *dummyIResource) Query(ctx context.Context, uri string) ([]protocol.ResourceContent, error) {
	return nil, nil
}

func (x *dummyIResource) Watch(ctx context.Context, uri string, ch chan<- protocol.ResourceUpdatedNotification) error {
	return nil
}

func (x *dummyIResource) CloseWatch(ctx context.Context, uri string) error {
	return nil
}

func (x *dummyIResource) StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.ResourceListChangedNotification) error {
	return nil
}

type dummyIPromptBuilder struct {
}

func (x *dummyIPromptBuilder) Build() iface.IPrompt {
	return &dummyIPrompt{}
}

func (x *dummyIPromptBuilder) ListChanged() bool {
	return false
}

type dummyIPrompt struct{}

func (x *dummyIPrompt) List(ctx context.Context, cursor string) ([]protocol.Prompt, string, error) {
	return nil, "", nil
}

func (x *dummyIPrompt) Get(ctx context.Context, name string, arguments map[string]string) (string, []protocol.PromptMessage, error) {
	return "", nil, nil
}

func (x *dummyIPrompt) StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.PromptListChangedNotification) error {
	return nil
}

type dummyIToolBuilder struct {
}

func (x *dummyIToolBuilder) Build() iface.ITool {
	return &dummyITool{}

}

func (x *dummyIToolBuilder) ListChanged() bool {
	return false
}

type dummyITool struct{}

func (x *dummyITool) List(ctx context.Context, cursor string) ([]protocol.Tool, string, error) {
	return nil, "", nil
}

func (x *dummyITool) Call(ctx context.Context, name string, arguments map[string]interface{}) ([]protocol.Content, error) {
	return nil, nil
}

func (x *dummyITool) StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.ToolListChangedNotification) error {
	return nil
}
