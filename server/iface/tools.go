package iface

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/protocol/jsonschema"

	"github.com/ccheers/xpkg/generic/arrayx"
)

type IToolUnit interface {
	// Name returns the name of the tool
	Name() string

	// Description returns the description of the tool
	Description() string

	// Schema returns the JSON schema of the tool's arguments
	Schema() *jsonschema.Definition

	// Call invokes the specified tool operation
	Call(ctx context.Context, argsJSON json.RawMessage) ([]protocol.Content, error)
}

type ToolsCallbackFunc[T any] func(ctx context.Context, args T) ([]protocol.Content, error)

type FunctionalToolDecodeFunc func(argsJSON json.RawMessage, receiver any) error

type FunctionalToolWrapperOptions struct {
	decodeFunc FunctionalToolDecodeFunc
}

type IFunctionalToolWrapperOption interface {
	apply(*FunctionalToolWrapperOptions)
}

type FunctionalToolWrapperOptionFunc func(*FunctionalToolWrapperOptions)

func (f FunctionalToolWrapperOptionFunc) apply(opts *FunctionalToolWrapperOptions) {
	f(opts)
}

func WithFunctionalToolWrapperDecodeFunc(decodeFunc FunctionalToolDecodeFunc) FunctionalToolWrapperOptionFunc {
	return func(opts *FunctionalToolWrapperOptions) {
		opts.decodeFunc = decodeFunc
	}
}

func defaultFunctionalToolWrapperOptions() FunctionalToolWrapperOptions {
	return FunctionalToolWrapperOptions{
		decodeFunc: func(argsJSON json.RawMessage, receiver any) error {
			return json.Unmarshal(argsJSON, receiver)
		},
	}
}

type FunctionalToolWrapper[T any] struct {
	fn          ToolsCallbackFunc[T]
	name        string
	description string
	schema      *jsonschema.Definition

	// options
	options FunctionalToolWrapperOptions
}

func NewFunctionalToolWrapper[T any](name string, description string, fn ToolsCallbackFunc[T],
	opts ...IFunctionalToolWrapperOption,
) *FunctionalToolWrapper[T] { //nolint:whitespace
	options := defaultFunctionalToolWrapperOptions()
	for _, opt := range opts {
		opt.apply(&options)
	}

	var zero T

	return &FunctionalToolWrapper[T]{
		fn:          fn,
		name:        name,
		description: description,
		options:     options,
		schema:      mustGenSchema(zero),
	}
}

func mustGenSchema(input any) *jsonschema.Definition {
	schema, err := jsonschema.GenerateSchemaForType(input)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate schema: %v", err))
	}
	return schema
}

func (x *FunctionalToolWrapper[T]) Name() string {
	return x.name
}

func (x *FunctionalToolWrapper[T]) Description() string {
	return x.description
}

func (x *FunctionalToolWrapper[T]) Schema() *jsonschema.Definition {
	return x.schema
}

func (x *FunctionalToolWrapper[T]) Call(ctx context.Context, argsJSON json.RawMessage) ([]protocol.Content, error) {
	var args T
	if err := x.options.decodeFunc(argsJSON, &args); err != nil {
		return nil, err
	}
	return x.fn(ctx, args)
}

type FunctionalTools struct {
	tools       map[string]IToolUnit
	toolsDefine []protocol.Tool
}

func NewFunctionalTools(tools map[string]IToolUnit, toolsDefine []protocol.Tool) *FunctionalTools {
	return &FunctionalTools{
		tools:       tools,
		toolsDefine: toolsDefine,
	}
}

func (x *FunctionalTools) List(_ context.Context, _ string) ([]protocol.Tool, string, error) {
	return x.toolsDefine, "", nil
}

func (x *FunctionalTools) Call(ctx context.Context, name string, argsJSON json.RawMessage) ([]protocol.Content, error) {
	for _, tool := range x.tools {
		if tool.Name() == name {
			return tool.Call(ctx, argsJSON)
		}
	}
	return nil, fmt.Errorf("tool %s not found", name)
}

func (x *FunctionalTools) StartWatchListChanged(_ context.Context, _ string, _ chan<- protocol.ToolListChangedNotification) error {
	// This method is not implemented in this example
	return nil
}

type FunctionalToolsBuilder struct {
	tools       map[string]IToolUnit
	toolsDefine []protocol.Tool
}

func NewFunctionalToolsBuilder(tools ...IToolUnit) *FunctionalToolsBuilder {
	return &FunctionalToolsBuilder{
		tools: arrayx.BuildMap(tools, func(t IToolUnit) string {
			return t.Name()
		}),
		toolsDefine: func() []protocol.Tool {
			toolsDefine := make([]protocol.Tool, len(tools))
			for i, tool := range tools {
				toolsDefine[i] = protocol.Tool{
					Name:        tool.Name(),
					Description: tool.Description(),
					InputSchema: tool.Schema(),
				}
			}
			return toolsDefine
		}(),
	}
}

func (x *FunctionalToolsBuilder) Build() ITool {
	return NewFunctionalTools(x.tools, x.toolsDefine)
}

func (x *FunctionalToolsBuilder) ListChanged() bool {
	return len(x.tools) > 0
}
