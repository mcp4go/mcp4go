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
				}),
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
