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
	// 创建上下文，支持优雅关闭
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

	// 监听中断信号，以便优雅退出
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		log.Println("Received shutdown signal")
		cancel()
	}()

	// 创建 weather 资源、提示和工具实例
	weatherResource := NewWeatherResourceBuilder()
	weatherPrompt := NewWeatherPromptBuilder()
	weatherTool := NewWeatherToolBuilder()

	// 创建标准输入/输出传输层
	stdioTransport := transport.NewStdioTransport()

	// 创建 MCP 服务器，并配置选项
	srv := server.NewServer(
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

	// 启动服务器
	log.Println("Starting Weather MCP server")
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server shutdown complete")
}
