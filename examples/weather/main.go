package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/server"
	"github.com/mcp4go/mcp4go/server/transport"
)

func main() {
	// Create context, supporting graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	homeDIR, _ := os.UserHomeDir()
	logDIR := filepath.Join(homeDIR, ".mcp4go", "logs")
	os.MkdirAll(logDIR, os.ModePerm)
	logFile, err := os.OpenFile(filepath.Join(logDIR, "weather.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(logFile)

	// Listen for interrupt signals for graceful exit
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Create weather resources, prompts, and tool instances
	weatherResource := NewWeatherResourceBuilder()
	weatherPrompt := NewWeatherPromptBuilder()
	weatherTool := NewWeatherToolBuilder()

	// Create standard input/output transport layer
	stdioTransport := transport.NewStdioTransport()

	// Create MCP server and configure options
	srv, cleanup, err := server.NewServer(
		stdioTransport,
		server.WithServerInfo(protocol.Implementation{
			Name:    "weather-mcp",
			Version: "0.1.0",
		}),
		server.WithInstructions("Welcome to Weather MCP! This server provides weather data, prompts, and tools."),
		server.WithResourceBuilder(weatherResource),
		server.WithPromptBuilder(weatherPrompt),
		server.WithToolBuilder(weatherTool),
	)
	if err != nil {
		log.Printf("Failed to create server: %v\n", err)
		return
	}
	defer cleanup()

	// Start the server
	log.Println("Starting Weather MCP server")
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server shutdown complete")
}
