package transport

import (
	"context"
	"io"
)

// ITransport defines the interface for MCP client transports
type ITransport interface {
	// Connect initializes the transport and returns reader and writer for communication
	Connect(ctx context.Context) (io.Reader, io.Writer, error)

	// Close closes the transport
	Close() error
}
