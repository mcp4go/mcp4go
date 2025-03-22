package server

import (
	"context"
	"encoding/json"

	"github.com/mcp4go/mcp4go/pkg/logger"
	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/server/iface"
)

// defaultOptions returns a new options struct with default values
func defaultOptions() (options, error) {
	_log, err := logger.NewDefaultLogger("server", logger.LevelDebug)
	if err != nil {
		return options{}, err
	}
	return options{
		serverInfo: protocol.Implementation{
			Name:    "mcp4go",
			Version: "0.1.0",
		},
		instructions:    "Welcome to mcp4go!",
		logger:          _log,
		resourceBuilder: &dummyIResourceBuilder{},
		promptBuilder:   &dummyIPromptBuilder{},
		toolBuilder:     &dummyIToolBuilder{},
	}, nil
}

// Option is a function that configures the server options
type Option interface {
	apply(*options)
}

type OptionFunc func(*options)

func (fn OptionFunc) apply(o *options) {
	fn(o)
}

// WithLogger sets the logger interface
func WithLogger(logger logger.ILogger) OptionFunc {
	return func(o *options) {
		o.logger = logger
	}
}

// WithServerInfo sets the server implementation info
func WithServerInfo(info protocol.Implementation) OptionFunc {
	return func(o *options) {
		o.serverInfo = info
	}
}

// WithInstructions sets the server instructions
func WithInstructions(instructions string) OptionFunc {
	return func(o *options) {
		o.instructions = instructions
	}
}

// WithResourceBuilder sets the resource builder interface
func WithResourceBuilder(resource iface.IResourceBuilder) OptionFunc {
	return func(o *options) {
		o.resourceBuilder = resource
	}
}

// WithPromptBuilder sets the prompt builder interface
func WithPromptBuilder(prompt iface.IPromptBuilder) OptionFunc {
	return func(o *options) {
		o.promptBuilder = prompt
	}
}

// WithToolBuilder sets the tool builder interface
func WithToolBuilder(tool iface.IToolBuilder) OptionFunc {
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

func (x *dummyIResource) List(_ context.Context, _ string) ([]protocol.Resource, string, error) {
	return nil, "", nil
}

func (x *dummyIResource) Query(_ context.Context, _ string) ([]protocol.ResourceContent, error) {
	return nil, nil
}

func (x *dummyIResource) Watch(_ context.Context, _ string, _ chan<- protocol.ResourceUpdatedNotification) error {
	return nil
}

func (x *dummyIResource) CloseWatch(_ context.Context, _ string) error {
	return nil
}

func (x *dummyIResource) StartWatchListChanged(_ context.Context, _ string, _ chan<- protocol.ResourceListChangedNotification) error {
	return nil
}

type dummyIPromptBuilder struct{}

func (x *dummyIPromptBuilder) Build() iface.IPrompt {
	return &dummyIPrompt{}
}

func (x *dummyIPromptBuilder) ListChanged() bool {
	return false
}

type dummyIPrompt struct{}

func (x *dummyIPrompt) List(_ context.Context, _ string) ([]protocol.Prompt, string, error) {
	return nil, "", nil
}

func (x *dummyIPrompt) Get(_ context.Context, _ string, _ map[string]string) (string, []protocol.PromptMessage, error) {
	return "", nil, nil
}

func (x *dummyIPrompt) StartWatchListChanged(_ context.Context, _ string, _ chan<- protocol.PromptListChangedNotification) error {
	return nil
}

type dummyIToolBuilder struct{}

func (x *dummyIToolBuilder) Build() iface.ITool {
	return &dummyITool{}
}

func (x *dummyIToolBuilder) ListChanged() bool {
	return false
}

type dummyITool struct{}

func (x *dummyITool) List(_ context.Context, _ string) ([]protocol.Tool, string, error) {
	return nil, "", nil
}

func (x *dummyITool) Call(_ context.Context, _ string, _ json.RawMessage) ([]protocol.Content, error) {
	return nil, nil
}

func (x *dummyITool) StartWatchListChanged(_ context.Context, _ string, _ chan<- protocol.ToolListChangedNotification) error {
	return nil
}
