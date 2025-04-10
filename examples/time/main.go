package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mcp4go/mcp4go/pkg/logger"
	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/protocol/jsonschema"
	"github.com/mcp4go/mcp4go/server"
	"github.com/mcp4go/mcp4go/server/iface"
	"github.com/mcp4go/mcp4go/server/transport"

	_ "time/tzdata" // Load all time zones
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_logger := logger.DefaultLog

	// Listen for interrupt signals for graceful exit
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		_logger.Logf(ctx, logger.LevelWarn, "Received shutdown signal")
		cancel()
	}()

	// Create standard input/output transport layer
	stdioTransport := transport.NewStdioTransport()
	// Create MCP server and configure options
	srv, cleanup, err := server.NewServer(
		stdioTransport,
		server.WithServerInfo(protocol.Implementation{
			Name:    "time-mcp",
			Version: "0.1.0",
		}),
		server.WithInstructions("Welcome to Time MCP! This server provides Time tools."),
		server.WithToolBuilder(newSimpleTimeBuilder()),
	)
	if err != nil {
		log.Printf("Failed to create server: %v\n", err)
		return
	}
	defer cleanup()

	// Start the server
	_logger.Logf(ctx, logger.LevelWarn, "Starting Time MCP server")
	if err := srv.Run(ctx); err != nil {
		_logger.Logf(ctx, logger.LevelError, "Server error: %v", err)
		return
	}

	_logger.Logf(ctx, logger.LevelWarn, "Server shutdown complete")
}

type simpleTime struct {
}

func newSimpleTime() *simpleTime {
	return &simpleTime{}
}

type timeRequest struct {
	TimeZone string `json:"time_zone,omitempty" description:"time zone default is Asia/Shanghai"`
}

func (x *simpleTime) List(ctx context.Context, cursor string) ([]protocol.Tool, string, error) {
	define, err := jsonschema.GenerateSchemaForType(timeRequest{})
	if err != nil {
		return nil, "", err
	}
	return []protocol.Tool{
		{
			Name:        "simple_time",
			Description: "get current time",
			InputSchema: define,
		},
	}, "", nil
}

func (x *simpleTime) Call(ctx context.Context, name string, argsJSON json.RawMessage) ([]protocol.Content, error) {
	var args timeRequest
	if err := json.Unmarshal(argsJSON, &args); err != nil {
		return nil, err
	}
	if args.TimeZone == "" {
		args.TimeZone = "Asia/Shanghai"
	}
	currentTime := x.getTimezoneTime(ctx, args.TimeZone)

	return []protocol.Content{
		protocol.NewTextContent(fmt.Sprintf("current time is %s", currentTime), nil),
	}, nil
}
func (x *simpleTime) getTimezoneTime(_ context.Context, timeZone string) time.Time {
	if timeZone == "" {
		timeZone = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return time.Now()
	}
	return time.Now().In(loc)
}

func (x *simpleTime) StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.ToolListChangedNotification) error {
	return fmt.Errorf("not support")
}

type simpleTimeBuilder struct {
}

func newSimpleTimeBuilder() *simpleTimeBuilder {
	return &simpleTimeBuilder{}
}

func (x *simpleTimeBuilder) Build() iface.ITool {
	return newSimpleTime()
}

func (x *simpleTimeBuilder) ListChanged() bool {
	return true
}
