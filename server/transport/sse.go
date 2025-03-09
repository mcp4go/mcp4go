package transport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// SSETransport implements the ITransport interface for Server-Sent Events
type SSETransport struct {
	addr   string
	path   string
	server *http.Server
}

// NewSSETransport creates a new SSETransport instance
func NewSSETransport(addr, path string) *SSETransport {
	return &SSETransport{
		addr: addr,
		path: path,
	}
}

// Run implements the ITransport interface
// It starts an HTTP server and handles SSE connections
func (s *SSETransport) Run(ctx context.Context, handle func(context.Context, io.Reader, io.Writer) error) error {
	// Create a new HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc(s.path, func(w http.ResponseWriter, r *http.Request) {
		//nolint:govet // Ignore error since we're just logging
		ctx := r.Context()
		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Create flush-supporting writer
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

		// Create an SSE connection
		conn := newSSEConnection(r.Body, w, flusher)

		// Handle the connection
		if err := handle(ctx, conn, conn); err != nil {
			// Log the error but continue
			fmt.Printf("Error handling SSE connection: %v\n", err)
		}
	})

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	// Start server in a goroutine
	serverErrCh := make(chan error, 1)
	go func() {
		err := s.server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- nil
			return
		}
		if err != nil {
			serverErrCh <- err
			return
		}
		serverErrCh <- nil
	}()

	// Process incoming connections
	for {
		select {
		case <-ctx.Done():
			// Context was canceled, shutdown server
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err := s.server.Shutdown(shutdownCtx)
			if err != nil {
				return fmt.Errorf("server shutdown error: %w", err)
			}
			return ctx.Err()
		case err := <-serverErrCh:
			// Server error
			return err
		}
	}
}

// sseConnection implements io.Reader and io.Writer for SSE
type sseConnection struct {
	reader  io.Reader
	writer  io.Writer
	flusher http.Flusher
	mu      sync.Mutex
}

func newSSEConnection(r io.Reader, w io.Writer, f http.Flusher) *sseConnection {
	return &sseConnection{
		reader:  r,
		writer:  w,
		flusher: f,
	}
}

// Read implements io.Reader
func (c *sseConnection) Read(p []byte) (n int, err error) {
	return c.reader.Read(p)
}

// Write implements io.Writer
// Formats the data as an SSE event and flushes it to the client
func (c *sseConnection) Write(p []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Write to the underlying writer
	n, err = c.writer.Write(p)
	if err != nil {
		return 0, err
	}

	// Flush to ensure data is sent immediately
	c.flusher.Flush()

	// Return original length since we're reporting on the input
	return n, nil
}
