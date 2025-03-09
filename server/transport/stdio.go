package transport

import (
	"context"
	"io"
	"os"
)

// StdioTransport implements the ITransport interface for standard input/output
type StdioTransport struct{}

// NewStdioTransport creates a new StdioTransport instance
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{}
}

// Run implements the ITransport interface
// It uses stdin for reading and stdout for writing
func (s *StdioTransport) Run(ctx context.Context, handle func(context.Context, io.Reader, io.Writer) error) error {
	// Create a done channel to signal completion
	done := make(chan error, 1)

	// Start handling the connection in a goroutine
	go func() {
		// Use os.Stdin as the reader and os.Stdout as the writer
		err := handle(ctx, os.Stdin, os.Stdout)
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
