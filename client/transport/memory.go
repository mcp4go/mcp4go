package transport

import (
	"context"
	"io"
)

// MemoryClientTransport implements ITransport for memory communication (client side)
type MemoryClientTransport struct {
	reader io.Reader
	writer io.Writer
}

// NewMemoryClientTransport creates a new memory client transport
func NewMemoryClientTransport(reader io.Reader, writer io.Writer) *MemoryClientTransport {
	return &MemoryClientTransport{
		reader: reader,
		writer: writer,
	}
}

// Connect implements the client ITransport interface
func (t *MemoryClientTransport) Connect(ctx context.Context) (io.Reader, io.Writer, error) {
	// Monitor context cancellation
	go func() {
		<-ctx.Done()
		t.Close()
	}()

	return t.reader, t.writer, nil
}

// Close implements the client ITransport interface
func (t *MemoryClientTransport) Close() error {
	return nil
}
