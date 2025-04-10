<p align="center">

<a href="https://github.com/mcp4go/mcp4go/actions/workflows/go.yml"><img src="https://github.com/mcp4go/mcp4go/actions/workflows/go.yml/badge.svg?v=1231" alt="Go Test"></a>
<a href="https://github.com/mcp4go/mcp4go/actions/workflows/lint.yml"><img src="https://github.com/mcp4go/mcp4go/actions/workflows/lint.yml/badge.svg?v=1231" alt="Lint"></a>
<a href="https://github.com/mcp4go/mcp4go/actions/workflows/codeql.yml"><img src="https://github.com/mcp4go/mcp4go/actions/workflows/codeql.yml/badge.svg?v=1231" alt="codeql"></a>
<a href="https://pkg.go.dev/github.com/mcp4go/mcp4go"><img src="https://pkg.go.dev/badge/github.com/mcp4go/mcp4go?v=1231" alt="GoDoc"></a>
<a href="https://codecov.io/gh/mcp4go/mcp4go"><img src="https://codecov.io/gh/mcp4go/mcp4go/master/graph/badge.svg?v=1231" alt="CodeCover"></a>
<a href="https://goreportcard.com/report/github.com/mcp4go/mcp4go"><img src="https://goreportcard.com/badge/github.com/mcp4go/mcp4go?v=1231" alt="Go Report Card"></a>
<a href="https://github.com/mcp4go/mcp4go/blob/main/LICENSE"><img src="https://img.shields.io/github/license/mcp4go/mcp4go?v=1231" alt="License"></a>
<a href="https://github.com/mcp4go/mcp4go/releases"><img src="https://img.shields.io/github/v/release/mcp4go/mcp4go?v=1231" alt="Release"></a>

</p>

# MCP4Go

MCP4Go is a Go implementation of the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/introduction), designed to simplify the development of AI applications by abstracting away protocol complexities.

## Features

- Complete MCP protocol implementation in pure Go
- High-level abstractions for common MCP resources
- Pluggable architecture for custom extensions
- Comprehensive documentation and examples
- Production-ready with robust error handling

## Installation

MCP4Go requires Go 1.18 or later. Install it using Go modules:

```bash
go get github.com/mcp4go/mcp4go
```

## Getting Started

To get started with MCP4Go, import the package in your Go application:

```go
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
```



## Protocol Implementation
MCP4Go provides a complete implementation of the Model Context Protocol with support for:
- JSON-RPC communication
- Resource lifecycle management
- Prompt engineering
- Tool definitions and invocations
- Sampling parameters
- Logging and diagnostics

## License
This project is licensed under the MIT License
## Contributing
Contributions are welcome! Please see our Contributing Guide for more information.
