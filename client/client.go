package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/mcp4go/mcp4go/client/transport"
	"github.com/mcp4go/mcp4go/pkg/logger"
	"github.com/mcp4go/mcp4go/protocol"

	"github.com/ccheers/xpkg/sync/errgroup"
)

// Client is the main client implementation for the Model Context Protocol
type Client struct {
	options options
	log     *logger.LogHelper

	transport transport.ITransport

	requestID int64

	cancel      context.CancelFunc
	eg          *errgroup.Group
	initialized bool
	mu          sync.Mutex

	// Handlers
	notificationHandlers map[protocol.McpMethod]NotificationHandler
	responseHandlers     map[int]chan *protocol.JsonrpcResponse

	writeChan chan json.RawMessage

	// Server capabilities
	serverCapabilities protocol.ServerCapabilities
	serverInfo         protocol.Implementation
	instructions       string
}

// NotificationHandler is a function that processes notifications
type NotificationHandler func(context.Context, json.RawMessage) error

// ResponseHandler processes a response and returns a result or error
type ResponseHandler func(response *protocol.JsonrpcResponse) (interface{}, error)

// Option is a function that configures a Client
type Option func(*options)

// options holds the configurable options for a Client
type options struct {
	logger logger.ILogger

	clientInfo   protocol.Implementation
	capabilities protocol.ClientCapabilities

	// Notification handlers
	resourcesListChangedHandler func(context.Context, protocol.ResourceListChangedNotification) error
	resourcesUpdatedHandler     func(context.Context, protocol.ResourceUpdatedNotification) error
	toolsListChangedHandler     func(context.Context, protocol.ToolListChangedNotification) error
	promptsListChangedHandler   func(context.Context, protocol.PromptListChangedNotification) error
	rootsListChangedHandler     func(context.Context, protocol.RootsListChangedNotification) error
	loggingMessageHandler       func(context.Context, protocol.LoggingMessageNotification) error
}

// WithLogger sets the logger for the client
func WithLogger(log logger.ILogger) Option {
	return func(o *options) {
		o.logger = log
	}
}

// WithClientInfo sets client name and version
func WithClientInfo(name, version string) Option {
	return func(o *options) {
		o.clientInfo = protocol.Implementation{
			Name:    name,
			Version: version,
		}
	}
}

// WithRootsCapability enables client roots capabilities
func WithRootsCapability(listChanged bool) Option {
	return func(o *options) {
		if o.capabilities.Roots == nil {
			o.capabilities.Roots = &protocol.ClientRoots{}
		}
		o.capabilities.Roots.ListChanged = listChanged
	}
}

// WithSamplingCapability enables client sampling capabilities
func WithSamplingCapability() Option {
	return func(o *options) {
		o.capabilities.Sampling = &protocol.ClientSampling{}
	}
}

// WithResourcesListChangedHandler sets a handler for resource list changes
func WithResourcesListChangedHandler(handler func(context.Context, protocol.ResourceListChangedNotification) error) Option {
	return func(o *options) {
		o.resourcesListChangedHandler = handler
	}
}

// WithResourcesUpdatedHandler sets a handler for resource updates
func WithResourcesUpdatedHandler(handler func(context.Context, protocol.ResourceUpdatedNotification) error) Option {
	return func(o *options) {
		o.resourcesUpdatedHandler = handler
	}
}

// WithToolsListChangedHandler sets a handler for tool list changes
func WithToolsListChangedHandler(handler func(context.Context, protocol.ToolListChangedNotification) error) Option {
	return func(o *options) {
		o.toolsListChangedHandler = handler
	}
}

// WithPromptsListChangedHandler sets a handler for prompt list changes
func WithPromptsListChangedHandler(handler func(context.Context, protocol.PromptListChangedNotification) error) Option {
	return func(o *options) {
		o.promptsListChangedHandler = handler
	}
}

// WithRootsListChangedHandler sets a handler for roots list changes
func WithRootsListChangedHandler(handler func(context.Context, protocol.RootsListChangedNotification) error) Option {
	return func(o *options) {
		o.rootsListChangedHandler = handler
	}
}

// WithLoggingMessageHandler sets a handler for logging messages
func WithLoggingMessageHandler(handler func(context.Context, protocol.LoggingMessageNotification) error) Option {
	return func(o *options) {
		o.loggingMessageHandler = handler
	}
}

// defaultOptions returns the default client options
func defaultOptions() options {
	return options{
		clientInfo: protocol.Implementation{
			Name:    "mcp4go-client",
			Version: "0.1.0",
		},
		capabilities: protocol.ClientCapabilities{},
		logger:       logger.DefaultLog,
	}
}

// NewClient creates a new MCP client with the given transport and options
func NewClient(t transport.ITransport, opts ...Option) (*Client, func(), error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Client{
		options:              options,
		log:                  logger.NewLogHelper(options.logger),
		transport:            t,
		requestID:            1,
		cancel:               nil,
		eg:                   nil,
		initialized:          false,
		mu:                   sync.Mutex{},
		notificationHandlers: make(map[protocol.McpMethod]NotificationHandler),
		responseHandlers:     make(map[int]chan *protocol.JsonrpcResponse),
		writeChan:            make(chan json.RawMessage, 1024),
		serverCapabilities:   protocol.ServerCapabilities{},
		serverInfo:           protocol.Implementation{},
		instructions:         "",
	}, func() {}, nil
}

// Connect establishes a connection to the server and initializes it
func (x *Client) Connect(ctx context.Context) error {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(ctx)
	x.eg = errgroup.WithCancel(ctx)
	x.cancel = cancel

	// Connect transport
	var err error
	reader, writer, err := x.transport.Connect(ctx)
	if err != nil {
		return fmt.Errorf("transport connect failed: %w", err)
	}
	x.log.Debugf(ctx, "Connected to server")

	// Register notification handlers
	x.registerNotificationHandlers()

	// Start loop
	x.eg.Go(func(ctx context.Context) error {
		defer func() {
			r := recover()
			if r != nil {
				x.log.Errorf(ctx, "[Client][ReadLoop] panic: %v, stack:\n%s\n", r, debug.Stack())
			}
		}()
		x.readLoop(ctx, reader)
		return nil
	})
	x.eg.Go(func(ctx context.Context) error {
		defer func() {
			r := recover()
			if r != nil {
				x.log.Errorf(ctx, "[Client][WriteLoop] panic: %v, stack:\n%s\n", r, debug.Stack())
			}
		}()
		x.writeLoop(ctx, writer)
		return nil
	})

	defer x.log.Warnf(ctx, "Initialized server...")
	// Initialize the server
	return x.initialize(ctx)
}

func (x *Client) readLoop(ctx context.Context, reader io.Reader) {
	decoder := json.NewDecoder(reader)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var message json.RawMessage
			if err := decoder.Decode(&message); err != nil {
				if errors.Is(err, io.EOF) || ctx.Err() != nil {
					return
				}
				x.log.Errorf(ctx, "Error decoding message: %v\n", err)
				continue
			}
			x.eg.Go(func(ctx context.Context) error {
				defer func() {
					r := recover()
					if r != nil {
						x.log.Errorf(ctx, "[Client][handleMessage] panic: %v, stack:\n%s\n", r, debug.Stack())
					}
				}()
				x.handleMessage(ctx, message)
				return nil
			})
		}
	}
}

func (x *Client) writeLoop(ctx context.Context, writer io.Writer) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-x.writeChan:
			bs, _ := json.Marshal(req)
			bs = append(bs, '\n')
			if _, err := writer.Write(bs); err != nil {
				x.log.Errorf(ctx, "Error encoding request: %v\n", err)
			}
		}
	}
}

// registerNotificationHandlers sets up the notification handlers
func (x *Client) registerNotificationHandlers() {
	// Register resource notification handlers
	if x.options.resourcesListChangedHandler != nil {
		x.notificationHandlers[protocol.NotificationResourcesListChanged] = func(ctx context.Context, message json.RawMessage) error {
			var dst protocol.ResourceListChangedNotification
			if err := json.Unmarshal(message, &dst); err != nil {
				return fmt.Errorf("failed to unmarshal resources list changed notification: %w", err)
			}
			return x.options.resourcesListChangedHandler(ctx, dst)
		}
	}
	if x.options.resourcesUpdatedHandler != nil {
		x.notificationHandlers[protocol.NotificationResourcesUpdated] = func(ctx context.Context, message json.RawMessage) error {
			var dst protocol.ResourceUpdatedNotification
			if err := json.Unmarshal(message, &dst); err != nil {
				return fmt.Errorf("failed to unmarshal resources updated notification: %w", err)
			}
			return x.options.resourcesUpdatedHandler(ctx, dst)
		}
	}

	// Register tool notification handlers
	if x.options.toolsListChangedHandler != nil {
		x.notificationHandlers[protocol.NotificationToolsListChanged] = func(ctx context.Context, message json.RawMessage) error {
			var dst protocol.ToolListChangedNotification
			if err := json.Unmarshal(message, &dst); err != nil {
				return fmt.Errorf("failed to unmarshal tools list changed notification: %w", err)
			}
			return x.options.toolsListChangedHandler(ctx, dst)
		}
	}

	// Register prompt notification handlers
	if x.options.promptsListChangedHandler != nil {
		x.notificationHandlers[protocol.NotificationPromptsListChanged] = func(ctx context.Context, message json.RawMessage) error {
			var dst protocol.PromptListChangedNotification
			if err := json.Unmarshal(message, &dst); err != nil {
				return fmt.Errorf("failed to unmarshal prompts list changed notification: %w", err)
			}
			return x.options.promptsListChangedHandler(ctx, dst)
		}
	}

	// Register roots notification handlers
	if x.options.rootsListChangedHandler != nil {
		x.notificationHandlers[protocol.NotificationRootsListChanged] = func(ctx context.Context, message json.RawMessage) error {
			var dst protocol.RootsListChangedNotification
			if err := json.Unmarshal(message, &dst); err != nil {
				return fmt.Errorf("failed to unmarshal roots list changed notification: %w", err)
			}
			return x.options.rootsListChangedHandler(ctx, dst)
		}
	}

	// Register logging notification handlers
	if x.options.loggingMessageHandler != nil {
		x.notificationHandlers[protocol.NotificationLoggingMessage] = func(ctx context.Context, message json.RawMessage) error {
			var dst protocol.LoggingMessageNotification
			if err := json.Unmarshal(message, &dst); err != nil {
				return fmt.Errorf("failed to unmarshal logging message notification: %w", err)
			}
			return x.options.loggingMessageHandler(ctx, dst)
		}
	}
}

// initialize sends the initialize request to the server
func (x *Client) initialize(ctx context.Context) error {
	// Create initialize request
	initRequest := protocol.InitializeRequest{
		ProtocolVersion: protocol.ProtocolVersion,
		Capabilities:    x.options.capabilities,
		ClientInfo:      x.options.clientInfo,
	}

	// Send initialize request
	var initResult protocol.InitializeResult
	err := x.sendRequest(ctx, protocol.MethodInitialize, initRequest, &initResult)
	if err != nil {
		return fmt.Errorf("initialize request failed: %w", err)
	}

	// Store server info
	x.serverCapabilities = initResult.Capabilities
	x.serverInfo = initResult.ServerInfo
	x.instructions = initResult.Instructions

	// Send initialized notification
	initialized := protocol.InitializedNotification{}
	err = x.sendNotification(ctx, protocol.NotificationInitialized, initialized)
	if err != nil {
		return fmt.Errorf("initialized notification failed: %w", err)
	}

	// Mark as initialized
	x.initialized = true

	return nil
}

// handleMessage processes a JSON-RPC message
func (x *Client) handleMessage(ctx context.Context, message json.RawMessage) {
	// Try to parse as a response
	var response protocol.JsonrpcResponse
	x.log.Errorf(ctx, "handleMessage: %s", string(message))
	if err := json.Unmarshal(message, &response); err == nil && response.ID != nil {
		x.handleResponse(ctx, &response)
		return
	}

	// Try to parse as a notification
	var notification protocol.JsonrpcNotification
	if err := json.Unmarshal(message, &notification); err == nil && notification.Method != "" {
		x.handleNotification(ctx, notification.Method, notification.Params)
		return
	}

	// Unknown message type
	x.log.Debugf(ctx, "Received unknown message type: %s\n", string(message))
}

// handleNotification processes a notification from the server
func (x *Client) handleNotification(ctx context.Context, method protocol.McpMethod, params json.RawMessage) {
	// Find handler for this notification
	handler, ok := x.notificationHandlers[method]
	if !ok {
		// No handler registered, ignore notification
		return
	}

	// Execute handler
	err := handler(ctx, params)
	if err != nil {
		x.log.Errorf(ctx, "Error handling notification %s: %v\n", method, err)
	}
}

// handleResponse processes a response from the server
func (x *Client) handleResponse(ctx context.Context, response *protocol.JsonrpcResponse) {
	// Find and remove handler for this response
	var respID int
	_ = json.Unmarshal(response.ID, &respID)
	x.mu.Lock()

	ch, ok := x.responseHandlers[respID]
	if ok {
		delete(x.responseHandlers, respID)
	}
	x.mu.Unlock()

	// If no handler found, ignore response
	if !ok {
		return
	}

	// Send response to handler
	select {
	case ch <- response:
	// Response sent successfully
	case <-ctx.Done():
	}
}

// sendRequest sends a request to the server and waits for the response
func (x *Client) sendRequest(ctx context.Context, method protocol.McpMethod, params interface{}, result interface{}) error {
	// Generate request ID
	id := atomic.AddInt64(&x.requestID, 1)

	// Marshal params
	var paramsBytes json.RawMessage
	if params != nil {
		var err error
		paramsBytes, err = json.Marshal(params)
		if err != nil {
			return fmt.Errorf("failed to marshal params: %w", err)
		}
	} else {
		paramsBytes = json.RawMessage("{}")
	}

	idBs, _ := json.Marshal(id)
	// Create JSON-RPC request
	request := protocol.NewJsonrpcRequest(
		idBs,
		method,
		paramsBytes,
	)

	// Create response channel
	responseCh := make(chan *protocol.JsonrpcResponse, 1)

	// Register response handler
	x.mu.Lock()
	x.responseHandlers[int(id)] = responseCh
	x.mu.Unlock()

	// Marshal and send request
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	select {
	case x.writeChan <- requestBytes:
	case <-ctx.Done():
		return ctx.Err()
	}

	// Wait for response or context cancellation
	select {
	case response := <-responseCh:
		// Process response
		if response.Error != nil {
			return fmt.Errorf("server error: %s (code: %d)", response.Error.Message, response.Error.Code)
		}

		// Unmarshal result
		if result != nil && response.Result != nil {
			if err := json.Unmarshal(response.Result, result); err != nil {
				return fmt.Errorf("failed to unmarshal result: %w", err)
			}
		}

		return nil

	case <-ctx.Done():
		// Context canceled
		x.mu.Lock()
		delete(x.responseHandlers, int(id))
		x.mu.Unlock()

		return ctx.Err()
	}
}

// sendNotification sends a notification to the server
func (x *Client) sendNotification(ctx context.Context, method protocol.McpMethod, params interface{}) error {
	// Marshal params
	var paramsBytes json.RawMessage
	if params != nil {
		var err error
		paramsBytes, err = json.Marshal(params)
		if err != nil {
			return fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	// Create JSON-RPC notification
	notification := protocol.NewJsonrpcNotification(method, paramsBytes)

	// Marshal and send notification
	notificationBytes, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	select {
	case x.writeChan <- notificationBytes:
	case <-ctx.Done():
		return fmt.Errorf("failed to send notification: %w", ctx.Err())
	}

	return nil
}

// Close terminates the client connection
func (x *Client) Close() error {
	x.mu.Lock()
	if x.cancel != nil {
		x.cancel()
	}
	x.mu.Unlock()

	// Wait for message processor to finish
	_ = x.eg.Wait()

	// Close transport
	if err := x.transport.Close(); err != nil {
		return fmt.Errorf("transport close failed: %w", err)
	}

	return nil
}

// ServerInfo returns the server implementation details
func (x *Client) ServerInfo() protocol.Implementation {
	return x.serverInfo
}

// ServerCapabilities returns the server capabilities
func (x *Client) ServerCapabilities() protocol.ServerCapabilities {
	return x.serverCapabilities
}

// Instructions returns the server instructions
func (x *Client) Instructions() string {
	return x.instructions
}

// IsInitialized checks if the client is initialized
func (x *Client) IsInitialized() bool {
	return x.initialized
}

// ListTools retrieves the list of available tools from the server
func (x *Client) ListTools(ctx context.Context) (protocol.ListToolsResult, error) {
	var result protocol.ListToolsResult
	err := x.sendRequest(ctx, protocol.MethodListTools, nil, &result)
	if err != nil {
		return protocol.ListToolsResult{}, err
	}
	return result, nil
}

// CallTool executes a tool on the server
func (x *Client) CallTool(ctx context.Context, request protocol.CallToolRequest) (protocol.CallToolResult, error) {
	var result protocol.CallToolResult
	err := x.sendRequest(ctx, protocol.MethodCallTool, request, &result)
	if err != nil {
		return protocol.CallToolResult{}, err
	}
	return result, nil
}

// ListResources retrieves the list of available resources from the server
func (x *Client) ListResources(ctx context.Context) (protocol.ListResourcesResult, error) {
	var result protocol.ListResourcesResult
	err := x.sendRequest(ctx, protocol.MethodListResources, nil, &result)
	if err != nil {
		return protocol.ListResourcesResult{}, err
	}
	return result, nil
}

// ReadResource reads a resource from the server
func (x *Client) ReadResource(ctx context.Context, request protocol.ReadResourceRequest) (protocol.ReadResourceResult, error) {
	var result protocol.ReadResourceResult
	err := x.sendRequest(ctx, protocol.MethodReadResource, request, &result)
	if err != nil {
		return protocol.ReadResourceResult{}, err
	}
	return result, nil
}

// SubscribeResource subscribes to updates for a resource
func (x *Client) SubscribeResource(ctx context.Context, request protocol.SubscribeRequest) error {
	return x.sendRequest(ctx, protocol.MethodSubscribe, request, nil)
}

// UnsubscribeResource unsubscribes from updates for a resource
func (x *Client) UnsubscribeResource(ctx context.Context, request protocol.UnsubscribeRequest) error {
	return x.sendRequest(ctx, protocol.MethodUnsubscribe, request, nil)
}

// ListPrompts retrieves the list of available prompts from the server
func (x *Client) ListPrompts(ctx context.Context) (protocol.ListPromptsResult, error) {
	var result protocol.ListPromptsResult
	err := x.sendRequest(ctx, protocol.MethodListPrompts, nil, &result)
	if err != nil {
		return protocol.ListPromptsResult{}, err
	}
	return result, nil
}

// GetPrompt retrieves a prompt from the server
func (x *Client) GetPrompt(ctx context.Context, request protocol.GetPromptRequest) (protocol.GetPromptResult, error) {
	var result protocol.GetPromptResult
	err := x.sendRequest(ctx, protocol.MethodGetPrompt, request, &result)
	if err != nil {
		return protocol.GetPromptResult{}, err
	}
	return result, nil
}

// ListRoots retrieves the list of available roots from the client
func (x *Client) ListRoots(ctx context.Context) (protocol.ListRootsResult, error) {
	var result protocol.ListRootsResult
	err := x.sendRequest(ctx, protocol.MethodListRoots, nil, &result)
	if err != nil {
		return protocol.ListRootsResult{}, err
	}
	return result, nil
}

// CreateMessage requests a completion from the client's LLM
func (x *Client) CreateMessage(ctx context.Context, request protocol.CreateMessageRequest) (protocol.CreateMessageResult, error) {
	var result protocol.CreateMessageResult
	err := x.sendRequest(ctx, protocol.MethodCreateMessage, request, &result)
	if err != nil {
		return protocol.CreateMessageResult{}, err
	}
	return result, nil
}

func (x *Client) Logger() *logger.LogHelper {
	return x.log
}
