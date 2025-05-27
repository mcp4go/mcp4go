package transport

import (
	"context"
	"io"
)

// MemoryServerTransport implements ITransport for memory communication (server side)
type MemoryServerTransport struct {
	reader io.Reader
	writer io.Writer
}

func NewMemoryServerTransport(reader io.Reader, writer io.Writer) *MemoryServerTransport {
	return &MemoryServerTransport{reader: reader, writer: writer}
}

// Run implements the server ITransport interface
func (t *MemoryServerTransport) Run(ctx context.Context, handle func(context.Context, io.Reader, io.Writer) error) error {
	// Create a done channel to signal completion
	done := make(chan error, 1)

	// Start handling the connection in a goroutine
	go func() {
		err := handle(ctx, t.reader, t.writer)
		done <- err
	}()

	// Wait for either context cancellation or handler completion
	select {
	case <-ctx.Done():
		// Context was canceled
		return ctx.Err()
	case err := <-done:
		// Handler completed
		return err
	}
}
