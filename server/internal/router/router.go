package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	incrID := func() uint32 {
		return atomic.AddUint32(&id, 1)
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
			var req protocol.JsonrpcRequest
			err := decoder.Decode(&req)
			if err != nil {
				return err
			}
			log.Println("req====", req)
			// handle request
			handler, ok := x.handlers[protocol.McpMethod(req.Method)]
			var respBs json.RawMessage
			if !ok {
				respBs, err = x.options.handleNotFound(ctx, req.Method, req.Params)
			} else {
				respBs, err = handler.Handle(ctx, req.Params)
			}
			if req.IsNotification() {
				return nil
			}
			if err != nil {
				x.writePackCH <- protocol.NewJsonrpcResponse(req.GetID(), nil, &protocol.JsonrpcError{
					Code:    -1,
					Message: err.Error(),
					Data:    "",
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

func DefaultNotFoundHandleFunc(ctx context.Context, method string, message json.RawMessage) (json.RawMessage, error) {
	log.Println("message====", string(message))
	return nil, fmt.Errorf("method(%s) not found", method)
}
