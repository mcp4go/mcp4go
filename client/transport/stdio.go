package transport

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// StdioTransport implements ITransport for stdio communication
type StdioTransport struct {
	cmd        *exec.Cmd
	stdinPipe  io.WriteCloser
	stdoutPipe io.ReadCloser
}

// StdioTransportOption is a function that configures a StdioTransport
type StdioTransportOption func(*StdioTransport)

// WithCommand sets the command to execute
func WithCommand(command string, args ...string) StdioTransportOption {
	return func(t *StdioTransport) {
		t.cmd = exec.Command(command, args...)
	}
}

// NewStdioTransport creates a new transport that uses standard I/O
func NewStdioTransport(opts ...StdioTransportOption) *StdioTransport {
	t := &StdioTransport{}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Connect starts the command if one is set, otherwise uses os.Stdin/os.Stdout
func (t *StdioTransport) Connect(ctx context.Context) (io.Reader, io.Writer, error) {
	if t.cmd != nil {
		// Use the command's stdin/stdout
		var err error

		// Create stdin pipe
		t.stdinPipe, err = t.cmd.StdinPipe()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create stdin pipe: %w", err)
		}

		// Create stdout pipe
		t.stdoutPipe, err = t.cmd.StdoutPipe()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create stdout pipe: %w", err)
		}

		// Redirect stderr to parent process stderr for debugging
		t.cmd.Stderr = os.Stderr

		// Start the command
		if err := t.cmd.Start(); err != nil {
			return nil, nil, fmt.Errorf("failed to start command: %w", err)
		}

		// Create a goroutine to handle command completion
		go func() {
			// Use a separate goroutine to avoid blocking
			<-ctx.Done()
			_ = t.Close()
		}()

		return t.stdoutPipe, t.stdinPipe, nil
	}

	// Use os.Stdin/os.Stdout if no command is set
	return os.Stdin, os.Stdout, nil
}

// Close terminates the transport
func (t *StdioTransport) Close() error {
	if t.cmd == nil {
		return nil
	}

	// Close pipes
	if t.stdinPipe != nil {
		_ = t.stdinPipe.Close()
	}

	if t.stdoutPipe != nil {
		_ = t.stdoutPipe.Close()
	}

	// Wait for command to exit or kill it
	if t.cmd.Process != nil {
		_ = t.cmd.Process.Kill()
		_ = t.cmd.Wait()
	}

	return nil
}
