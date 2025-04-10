package transport

import (
	"context"
	"io"
)

type ITransport interface {
	// Run starts the transport and blocks until it is stopped or context canceled
	// The handle function is the callback to process the connection. When handle returns, the connection will be closed
	Run(ctx context.Context, handle func(context.Context, io.Reader, io.Writer) error) error
}
