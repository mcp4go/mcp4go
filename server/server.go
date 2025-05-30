package server

import (
	"context"
	"fmt"
	"io"

	"github.com/mcp4go/mcp4go/pkg/logger"
	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/server/iface"
	"github.com/mcp4go/mcp4go/server/internal/handlers"
	"github.com/mcp4go/mcp4go/server/transport"
)

type options struct {
	serverInfo   protocol.Implementation
	instructions string

	requestDecodeFn handlers.RequestDecodeFunc

	logger logger.ILogger

	resourceBuilder iface.IResourceBuilder
	promptBuilder   iface.IPromptBuilder
	toolBuilder     iface.IToolBuilder
}

type Server struct {
	options   options
	log       *logger.LogHelper
	transport transport.ITransport
}

// NewServer creates a new server with the given transport and options
func NewServer(transport transport.ITransport, opts ...Option) (*Server, func(), error) {
	//nolint:govet
	options := defaultOptions()
	for _, opt := range opts {
		opt.apply(&options)
	}

	return &Server{
		options:   options,
		log:       logger.NewLogHelper(options.logger),
		transport: transport,
	}, func() {}, nil
}

func (x *Server) Run(ctx context.Context) error {
	return x.transport.Run(ctx, func(ctx context.Context, reader io.Reader, writer io.Writer) error {
		router, err := initRouter(
			x.options.logger,
			handlers.NewInitializeHandler(
				protocol.ServerCapabilities{
					Prompts: &protocol.ServerPrompts{
						ListChanged: x.options.promptBuilder.ListChanged(),
					},
					Resources: &protocol.ServerResources{
						Subscribe:   x.options.resourceBuilder.Subscribe(),
						ListChanged: x.options.resourceBuilder.ListChanged(),
					},
					Tools: &protocol.ServerTools{
						ListChanged: x.options.toolBuilder.ListChanged(),
					},
				},
				x.options.serverInfo,
				x.options.instructions,
				x.options.requestDecodeFn,
			),
			handlers.NewSetLevelHandler(x.options.requestDecodeFn),
			x.options.resourceBuilder.Build(),
			x.options.promptBuilder.Build(),
			x.options.toolBuilder.Build(),
			iface.NewEventBus(),
			x.options.requestDecodeFn,
		)
		if err != nil {
			return fmt.Errorf("failed to initialize router: %w", err)
		}
		return router.Handle(ctx, reader, writer)
	})
}

func (x *Server) Logger() *logger.LogHelper {
	return x.log
}
