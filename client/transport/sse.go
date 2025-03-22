package transport

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// SSETransport implements ITransport for Server-Sent Events
type SSETransport struct {
	baseURL     string
	httpClient  *http.Client
	sseReader   io.ReadCloser
	messageURL  string
	ctx         context.Context
	cancel      context.CancelFunc
	closeMutex  sync.Mutex
	closed      bool
	messageID   int
	messagePipe *pipe
}

// pipe implements a simple read/write pipe for messages
type pipe struct {
	reader     *io.PipeReader
	writer     *io.PipeWriter
	writeMutex sync.Mutex
}

// newPipe creates a new pipe
func newPipe() *pipe {
	r, w := io.Pipe()
	return &pipe{
		reader: r,
		writer: w,
	}
}

// Write writes data to the pipe
func (p *pipe) Write(data []byte) (int, error) {
	p.writeMutex.Lock()
	defer p.writeMutex.Unlock()
	return p.writer.Write(data)
}

// Read reads data from the pipe
func (p *pipe) Read(data []byte) (int, error) {
	return p.reader.Read(data)
}

// Close closes the pipe
func (p *pipe) Close() error {
	p.writeMutex.Lock()
	defer p.writeMutex.Unlock()
	_ = p.reader.Close()
	return p.writer.Close()
}

// SSETransportOption is a function that configures an SSETransport
type SSETransportOption func(*SSETransport)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) SSETransportOption {
	return func(t *SSETransport) {
		t.httpClient = client
	}
}

// WithMessageURL sets a custom message URL path
func WithMessageURL(path string) SSETransportOption {
	return func(t *SSETransport) {
		t.messageURL = path
	}
}

// NewSSETransport creates a new transport that uses SSE
func NewSSETransport(baseURL string, opts ...SSETransportOption) *SSETransport {
	t := &SSETransport{
		baseURL:     baseURL,
		httpClient:  http.DefaultClient,
		messageURL:  "/mcp/message", // Default path
		messageID:   1,
		messagePipe: newPipe(),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Connect establishes an SSE connection
func (t *SSETransport) Connect(ctx context.Context) (io.Reader, io.Writer, error) {
	t.ctx, t.cancel = context.WithCancel(ctx)

	// Establish SSE connection
	sseURL := fmt.Sprintf("%s/sse", strings.TrimRight(t.baseURL, "/"))
	req, err := http.NewRequestWithContext(t.ctx, "GET", sseURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create SSE request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to establish SSE connection: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	t.sseReader = resp.Body

	// Start SSE message reader
	go t.handleSSEMessages()

	return t.messagePipe.reader, t, nil
}

// handleSSEMessages processes incoming SSE messages
func (t *SSETransport) handleSSEMessages() {
	defer func() {
		t.Close()
	}()

	scanner := bufio.NewScanner(t.sseReader)
	var buffer bytes.Buffer
	var dataLine bool

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line indicates the end of a message
			if buffer.Len() > 0 {
				data := buffer.String()
				buffer.Reset()
				dataLine = false

				// Process the message
				if _, err := t.messagePipe.Write([]byte(data)); err != nil {
					return
				}
			}
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			// Data line
			data := line[6:] // Remove "data: " prefix
			if dataLine {
				buffer.WriteByte('\n')
			}
			buffer.WriteString(data)
			dataLine = true
		}
	}

	if err := scanner.Err(); err != nil {
		// Only log the error if we're not shutting down
		select {
		case <-t.ctx.Done():
			// Context was canceled, don't log the error
		default:
			fmt.Fprintf(io.Discard, "SSE scanner error: %v", err)
		}
	}
}

// Write sends a message to the server
func (t *SSETransport) Write(data []byte) (int, error) {
	messageURL := fmt.Sprintf("%s%s", strings.TrimRight(t.baseURL, "/"), t.messageURL)

	req, err := http.NewRequestWithContext(t.ctx, "POST", messageURL, bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("failed to create POST request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Check for JSON-RPC response
	var responseObj map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseObj); err != nil {
		// Not a valid JSON response
		return len(data), nil
	}

	// If we get a JSON-RPC response, write it to the pipe
	if jsonrpc, ok := responseObj["jsonrpc"].(string); ok && jsonrpc == "2.0" {
		responseBytes, err := json.Marshal(responseObj)
		if err != nil {
			return len(data), nil
		}

		if _, err := t.messagePipe.Write(responseBytes); err != nil {
			return 0, err
		}
	}

	return len(data), nil
}

// Close terminates the SSE connection
func (t *SSETransport) Close() error {
	t.closeMutex.Lock()
	defer t.closeMutex.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true

	if t.cancel != nil {
		t.cancel()
	}

	if t.sseReader != nil {
		_ = t.sseReader.Close()
	}

	if t.messagePipe != nil {
		_ = t.messagePipe.Close()
	}

	return nil
}
