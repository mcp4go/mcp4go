package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"runtime/debug"
	"sync/atomic"

	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/server/iface"

	"github.com/ccheers/xpkg/generic/arrayx"
	"github.com/ccheers/xpkg/sync/errgroup"
)

type IRouter interface {
	Handle(ctx context.Context, reader io.Reader, writer io.Writer) error
}

type Options struct {
	handleNotFound NotFoundHandleFunc
}

func defaultOptions() Options {
	return Options{
		handleNotFound: DefaultNotFoundHandleFunc,
	}
}

type NotFoundHandleFunc func(ctx context.Context, method string, message json.RawMessage) (json.RawMessage, error)

type IHandler interface {
	Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error)
	Method() protocol.McpMethod
}

type Router struct {
	options Options

	writePackCH chan *protocol.JsonrpcResponse

	handlers map[protocol.McpMethod]IHandler
	bus      iface.EventBus
}

func NewIRouter(x *Router) IRouter {
	return x
}

func NewRouter(list []IHandler, bus iface.EventBus) *Router {
	options := defaultOptions()
	return &Router{
		options:     options,
		writePackCH: make(chan *protocol.JsonrpcResponse, 2048),
		handlers: arrayx.BuildMap(list, func(t IHandler) protocol.McpMethod {
			return t.Method()
		}),
		bus: bus,
	}
}

func (x *Router) Handle(ctx context.Context, reader io.Reader, writer io.Writer) error {
	id := uint32(0)
	incrID := func() json.RawMessage {
		newID := atomic.AddUint32(&id, 1)
		bs, _ := json.Marshal(newID)
		return bs
	}
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
				x.writePackCH <- protocol.NewJsonrpcResponse(incrID(), bs, nil)
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
				x.writePackCH <- protocol.NewJsonrpcResponse(incrID(), bs, nil)
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
				x.writePackCH <- protocol.NewJsonrpcResponse(incrID(), bs, nil)
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
				x.writePackCH <- protocol.NewJsonrpcResponse(incrID(), bs, nil)
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
				x.writePackCH <- protocol.NewJsonrpcResponse(incrID(), bs, nil)
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
					log.Printf("panic: %v, stack:\n%s\n", r, debug.Stack())
				}
			}()
			var req protocol.JsonrpcRequest
			err := decoder.Decode(&req)
			if err != nil {
				return err
			}

			log.Printf("#%s. method[%s] params[%s]\n", req.GetID(), req.Method, string(req.Params))

			respBs, err := x.handle(ctx, &req)
			if err != nil {
				log.Printf("handle error: %v\n", err)
			}
			if req.IsNotification() {
				return nil
			}
			if err != nil {
				code := int64(-1)
				if errCode, ok := err.(interface{ Code() int64 }); ok {
					code = errCode.Code()
				}
				x.writePackCH <- protocol.NewJsonrpcResponse(req.GetID(), nil, &protocol.JsonrpcError{
					Code:    code,
					Message: err.Error(),
					Data:    nil,
				})
				return nil
			}
			x.writePackCH <- protocol.NewJsonrpcResponse(req.GetID(), respBs, nil)
			return nil
		}()
		if err != nil {
			return err
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
			err := encoder.Encode(pack)
			if err != nil {
				return err
			}
		}
	}
}

func (x *Router) handle(ctx context.Context, req *protocol.JsonrpcRequest) (json.RawMessage, error) {
	// handle request
	handler, ok := x.handlers[protocol.McpMethod(req.Method)]
	if !ok {
		return x.options.handleNotFound(ctx, req.Method, req.Params)
	}
	return handler.Handle(ctx, req.Params)
}

func DefaultNotFoundHandleFunc(_ context.Context, method string, message json.RawMessage) (json.RawMessage, error) {
	log.Printf("method(%s) not found, message=%s", method, message)
	return nil, fmt.Errorf("method(%s) not found", method)
}
