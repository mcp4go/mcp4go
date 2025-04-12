package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"
	"sync"

	"github.com/ccheers/xpkg/generic/arrayx"
	"github.com/ccheers/xpkg/sync/errgroup"

	"github.com/mcp4go/mcp4go/pkg/logger"
	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/server/iface"
)

type IRouter interface {
	Handle(ctx context.Context, reader io.Reader, writer io.Writer) error
}

type NotFoundHandleFunc func(ctx context.Context, method string, message json.RawMessage) (json.RawMessage, error)

type IHandler interface {
	Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error)
	Method() protocol.McpMethod
}

type IHandlerFunc func(ctx context.Context, message json.RawMessage) (json.RawMessage, error)

type IHandlerFuncWrapper struct {
	fn     IHandlerFunc
	method protocol.McpMethod
}

func NewIHandlerFuncWrapper(fn IHandlerFunc, method protocol.McpMethod) *IHandlerFuncWrapper {
	return &IHandlerFuncWrapper{fn: fn, method: method}
}

func (x *IHandlerFuncWrapper) Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error) {
	return x.fn(ctx, message)
}

func (x *IHandlerFuncWrapper) Method() protocol.McpMethod {
	return x.method
}

type Router struct {
	log *logger.LogHelper

	writePackCH chan *protocol.JsonrpcPack

	handlers      map[protocol.McpMethod]IHandler
	bus           iface.EventBus
	processingReq sync.Map
}

func NewIRouter(x *Router) IRouter {
	return x
}

func NewRouter(list []IHandler, bus iface.EventBus, _logger logger.ILogger) (*Router, error) {
	x := &Router{
		log:         logger.NewLogHelper(_logger),
		writePackCH: make(chan *protocol.JsonrpcPack, 2048),
		handlers:    nil,
		bus:         bus,
	}

	// add canceled handlers
	list = append(list, NewIHandlerFuncWrapper(x.cancelHandler(), protocol.NotificationCancelled))

	x.handlers = arrayx.BuildMap(list, func(t IHandler) protocol.McpMethod {
		return t.Method()
	})

	return x, nil
}

func (x *Router) Handle(ctx context.Context, reader io.Reader, writer io.Writer) error {
	eg := errgroup.WithCancel(ctx)
	eg.Go(func(ctx context.Context) error {
		return x.readLoop(ctx, reader)
	})
	eg.Go(func(ctx context.Context) error {
		return x.writeLoop(ctx, writer)
	})
	eg.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case msg := <-x.bus.PromptListChangedNotificationChan:
				bs, _ := json.Marshal(msg)
				x.writePackCH <- (*protocol.JsonrpcPack)(protocol.NewJsonrpcNotification(protocol.NotificationPromptsListChanged, bs))
			}
		}
	})
	eg.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case msg := <-x.bus.ResourceUpdatedNotificationChan:
				bs, _ := json.Marshal(msg)
				x.writePackCH <- (*protocol.JsonrpcPack)(protocol.NewJsonrpcNotification(protocol.NotificationResourcesUpdated, bs))
			}
		}
	})
	eg.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case msg := <-x.bus.ResourceListChangedNotificationChan:
				bs, _ := json.Marshal(msg)
				x.writePackCH <- (*protocol.JsonrpcPack)(protocol.NewJsonrpcNotification(protocol.NotificationResourcesListChanged, bs))
			}
		}
	})
	eg.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case msg := <-x.bus.ToolListChangedNotificationChan:
				bs, _ := json.Marshal(msg)
				x.writePackCH <- (*protocol.JsonrpcPack)(protocol.NewJsonrpcNotification(protocol.NotificationToolsListChanged, bs))
			}
		}
	})
	eg.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case msg := <-x.bus.LoggingMessageNotificationChan:
				bs, _ := json.Marshal(msg)
				x.writePackCH <- (*protocol.JsonrpcPack)(protocol.NewJsonrpcNotification(protocol.NotificationLoggingMessage, bs))
			}
		}
	})
	eg.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case msg := <-x.bus.ProgressNotificationChan:
				bs, _ := json.Marshal(msg)
				x.writePackCH <- (*protocol.JsonrpcPack)(protocol.NewJsonrpcNotification(protocol.NotificationProgress, bs))
			}
		}
	})
	return eg.Wait()
}

func (x *Router) readLoop(ctx context.Context, reader io.Reader) error {
	decoder := json.NewDecoder(reader)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		err := func() error {
			defer func() {
				if r := recover(); r != nil {
					x.log.Errorf(ctx, "[Router][handle] panic: %v, stack:\n%s\n", r, debug.Stack())
				}
			}()

			var req protocol.JsonrpcRequest
			err := decoder.Decode(&req)
			if err != nil {
				return fmt.Errorf("decode error: %w", err)
			}

			x.log.Debugf(ctx, "#%s. method[%s] params[%s]\n", req.GetID(), req.Method, string(req.Params))
			go func() {
				defer func() {
					if r := recover(); r != nil {
						x.log.Errorf(ctx, "[Router][handle] panic: %v, stack:\n%s\n", r, debug.Stack())
					}
				}()

				//nolint:govet
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()

				// add cancel callback
				x.processingReq.Store(string(req.GetID()), cancel)
				defer x.processingReq.Delete(string(req.GetID()))

				respBs, err := x.handle(ctx, &req)
				if err != nil {
					x.log.Errorf(ctx, "handle error: %v\n", err)
				}
				if req.IsNotification() {
					return
				}
				if err != nil {
					code := int64(-1)
					if errCode, ok := err.(interface{ Code() int64 }); ok {
						code = errCode.Code()
					}
					x.writePackCH <- (*protocol.JsonrpcPack)(
						protocol.NewJsonrpcResponse(req.GetID(), nil, &protocol.JsonrpcError{
							Code:    code,
							Message: err.Error(),
							Data:    nil,
						}),
					)
					return
				}
				x.writePackCH <- (*protocol.JsonrpcPack)(
					protocol.NewJsonrpcResponse(req.GetID(), respBs, nil),
				)
			}()
			return nil
		}()
		if err != nil {
			return fmt.Errorf("[Router][readLoop] read error: %w", err)
		}
	}
}

func (x *Router) writeLoop(ctx context.Context, writer io.Writer) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case pack := <-x.writePackCH:
			encoder := json.NewEncoder(writer)
			bs, _ := json.Marshal(pack)
			x.log.Debugf(ctx, "write response: %+v\n", string(bs))
			err := encoder.Encode(pack)
			if err != nil {
				return err
			}
		}
	}
}

func (x *Router) handle(ctx context.Context, req *protocol.JsonrpcRequest) (json.RawMessage, error) {
	// handle request
	handler, ok := x.handlers[req.Method]
	if !ok {
		return x.notFoundHandleFunc(ctx, req.Method, req.Params)
	}
	return handler.Handle(ctx, req.Params)
}

func (x *Router) notFoundHandleFunc(ctx context.Context, method protocol.McpMethod, message json.RawMessage) (json.RawMessage, error) {
	x.log.Errorf(ctx, "method(%s) not found, message=%s", method, message)
	return nil, fmt.Errorf("method(%s) not found", method)
}

func (x *Router) cancelHandler() IHandlerFunc {
	return func(_ context.Context, message json.RawMessage) (json.RawMessage, error) {
		var req protocol.CancelRequest
		err := json.Unmarshal(message, &req)
		if err != nil {
			return nil, err
		}
		cancel, ok := x.processingReq.Load(string(req.ID))
		if ok {
			cancel.(context.CancelFunc)()
		}
		return nil, nil
	}
}
