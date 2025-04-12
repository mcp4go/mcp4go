package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mcp4go/mcp4go/pkg/logger"
	"github.com/mcp4go/mcp4go/protocol"
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

	type timeRequest struct {
		TimeZone string `json:"time_zone,omitempty" description:"time zone default is Asia/Shanghai"`
	}

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
		server.WithToolBuilder(iface.NewFunctionalToolsBuilder(
			iface.NewFunctionalToolWrapper(
				"time",
				"get current time",
				func(ctx context.Context, args timeRequest) ([]protocol.Content, error) {
					if args.TimeZone == "" {
						args.TimeZone = "Asia/Shanghai"
					}

					getTimezoneTime := func(_ context.Context, timeZone string) time.Time {
						if timeZone == "" {
							timeZone = "Asia/Shanghai"
						}
						loc, err := time.LoadLocation(timeZone)
						if err != nil {
							return time.Now()
						}
						return time.Now().In(loc)
					}

					currentTime := getTimezoneTime(ctx, args.TimeZone)

					return []protocol.Content{
						protocol.NewTextContent(fmt.Sprintf("current time is %s", currentTime), nil),
					}, nil
				},
			),
		)),
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
