# MCP4Go Client

This is a Go implementation of a Model Context Protocol (MCP) client, designed to communicate with MCP servers.

## Features

- JSON-RPC 2.0 compliant implementation
- Supports standard MCP lifecycle (initialize, ping)
- Handles logging messages from the server
- Provides roots capability for filesystem access
- Implements sampling capability for LLM integration
- Flexible transport architecture (currently supports stdio)

## Usage

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mcp4go/mcp4go/client"
	"github.com/mcp4go/mcp4go/client/transport"
)

func main() {
	// Set up logger
	log.SetOutput(os.Stderr)
	log.SetPrefix("[MCP4GO-Client] ")

	// Create a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		log.Println("Received termination signal, shutting down...")
		cancel()
		// Force exit after a timeout
		time.AfterFunc(5*time.Second, func() {
			log.Println("Forced shutdown after timeout")
			os.Exit(1)
		})
	}()

	// Create client transport
	t := transport.NewStdioTransport()

	// Create client with options
	c := client.NewClient(
		t,
		client.WithClientInfo("my-mcp-client", "1.0.0"),
		client.WithRootsCapability(true),  // Enable roots capability
		client.WithSamplingCapability(),   // Enable LLM sampling capability
	)

	// Connect to the server
	log.Println("Connecting to MCP server...")
	if err := c.Connect(ctx); err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}

	log.Println("Client connection closed")
}
```

## Extending the Client

### Adding New Handlers

To add a new handler for an MCP method:

1. Create a new handler in `client/internal/handlers` that implements the `IHandler` interface
2. Register the handler in `client/wire.go`

### Adding New Transport

To add a new transport mechanism:

1. Create a new transport in `client/transport` that implements the `ITransport` interface
2. Use the new transport when creating the client

## License

See the project's LICENSE file.
