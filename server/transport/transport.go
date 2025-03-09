package transport

import (
	"context"
	"io"
)

type ITransport interface {
	// Run starts the transport and blocks until it is stopped or context canceled
	// the handle if the callback to handle the connection, if handle over, the connection will be closed
	Run(ctx context.Context, handle func(context.Context, io.Reader, io.Writer) error) error
}
