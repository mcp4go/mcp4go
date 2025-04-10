package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mcp4go/mcp4go/pkg/logger"
	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/server/iface"
)

// 模拟传输层，用于测试服务器
type mockTransport struct {
	reader     io.Reader
	writer     io.Writer
	handleFunc func(ctx context.Context, reader io.Reader, writer io.Writer) error
	mutex      sync.Mutex
	runCalled  bool
	runErr     error
}

func newMockTransport() *mockTransport {
	return &mockTransport{
		reader:    bytes.NewBuffer(nil),
		writer:    bytes.NewBuffer(nil),
		mutex:     sync.Mutex{},
		runCalled: false,
	}
}

func (m *mockTransport) Run(ctx context.Context, handler func(ctx context.Context, reader io.Reader, writer io.Writer) error) error {
	m.mutex.Lock()
	m.runCalled = true
	m.handleFunc = handler
	m.mutex.Unlock()

	if m.runErr != nil {
		return m.runErr
	}

	// 如果没有预设的错误，直接调用处理程序
	if m.handleFunc != nil {
		return m.handleFunc(ctx, m.reader, m.writer)
	}

	return nil
}

func (m *mockTransport) SetRunError(err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.runErr = err
}

func (m *mockTransport) SetReaderWriter(reader io.Reader, writer io.Writer) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.reader = reader
	m.writer = writer
}

func (m *mockTransport) IsRunCalled() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.runCalled
}

// 模拟资源构建器
type mockResourceBuilder struct {
	subscribeValue   bool
	listChangedValue bool
	resource         iface.IResource
}

func newMockResourceBuilder() *mockResourceBuilder {
	return &mockResourceBuilder{
		subscribeValue:   true,
		listChangedValue: true,
		resource:         &mockResource{},
	}
}

func (m *mockResourceBuilder) Build() iface.IResource {
	return m.resource
}

func (m *mockResourceBuilder) Subscribe() bool {
	return m.subscribeValue
}

func (m *mockResourceBuilder) ListChanged() bool {
	return m.listChangedValue
}

// 模拟资源接口
type mockResource struct {
	listResult   []protocol.Resource
	queryResult  []protocol.ResourceContent
	watchError   error
	listToken    string
	listError    error
	queryError   error
	closeError   error
	startError   error
	accessResult []protocol.ResourceTemplate
}

func (m *mockResource) AccessList() []protocol.ResourceTemplate {
	return m.accessResult
}

func (m *mockResource) List(_ context.Context, _ string) ([]protocol.Resource, string, error) {
	return m.listResult, m.listToken, m.listError
}

func (m *mockResource) Query(_ context.Context, _ string) ([]protocol.ResourceContent, error) {
	return m.queryResult, m.queryError
}

func (m *mockResource) Watch(_ context.Context, _ string, _ chan<- protocol.ResourceUpdatedNotification) error {
	return m.watchError
}

func (m *mockResource) CloseWatch(_ context.Context, _ string) error {
	return m.closeError
}

func (m *mockResource) StartWatchListChanged(_ context.Context, _ string, _ chan<- protocol.ResourceListChangedNotification) error {
	return m.startError
}

// 模拟工具构建器
type mockToolBuilder struct {
	listChangedValue bool
	tool             iface.ITool
}

func newMockToolBuilder() *mockToolBuilder {
	return &mockToolBuilder{
		listChangedValue: true,
		tool:             &mockTool{},
	}
}

func (m *mockToolBuilder) Build() iface.ITool {
	return m.tool
}

func (m *mockToolBuilder) ListChanged() bool {
	return m.listChangedValue
}

// 模拟工具接口
type mockTool struct {
	listResult []protocol.Tool
	callResult []protocol.Content
	listToken  string
	listError  error
	callError  error
	startError error
}

func (m *mockTool) List(_ context.Context, _ string) ([]protocol.Tool, string, error) {
	return m.listResult, m.listToken, m.listError
}

func (m *mockTool) Call(_ context.Context, _ string, _ json.RawMessage) ([]protocol.Content, error) {
	return m.callResult, m.callError
}

func (m *mockTool) StartWatchListChanged(_ context.Context, _ string, _ chan<- protocol.ToolListChangedNotification) error {
	return m.startError
}

// 模拟提示构建器
type mockPromptBuilder struct {
	listChangedValue bool
	prompt           iface.IPrompt
}

func newMockPromptBuilder() *mockPromptBuilder {
	return &mockPromptBuilder{
		listChangedValue: true,
		prompt:           &mockPrompt{},
	}
}

func (m *mockPromptBuilder) Build() iface.IPrompt {
	return m.prompt
}

func (m *mockPromptBuilder) ListChanged() bool {
	return m.listChangedValue
}

// 模拟提示接口
type mockPrompt struct {
	listResult []protocol.Prompt
	getText    string
	listToken  string
	listError  error
	getError   error
	startError error
	messages   []protocol.PromptMessage
}

func (m *mockPrompt) List(_ context.Context, _ string) ([]protocol.Prompt, string, error) {
	return m.listResult, m.listToken, m.listError
}

func (m *mockPrompt) Get(_ context.Context, _ string, _ map[string]string) (string, []protocol.PromptMessage, error) {
	return m.getText, m.messages, m.getError
}

func (m *mockPrompt) StartWatchListChanged(_ context.Context, _ string, _ chan<- protocol.PromptListChangedNotification) error {
	return m.startError
}

// 测试服务器创建
func TestNewServer(t *testing.T) {
	// 创建模拟传输层
	mockTransp := newMockTransport()

	// 创建服务器
	server, cleanup, err := NewServer(mockTransp)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer cleanup()

	// 验证服务器是否正确初始化
	if server == nil {
		t.Fatal("Server is nil")
	}

	// 验证默认选项
	if server.options.serverInfo.Name != "mcp4go" {
		t.Errorf("Expected server name 'mcp4go', got '%s'", server.options.serverInfo.Name)
	}

	if server.options.serverInfo.Version != "0.1.0" {
		t.Errorf("Expected server version '0.1.0', got '%s'", server.options.serverInfo.Version)
	}

	if server.options.instructions != "Welcome to mcp4go!" {
		t.Errorf("Expected instructions 'Welcome to mcp4go!', got '%s'", server.options.instructions)
	}
}

// 测试服务器选项
func TestServerOptions(t *testing.T) {
	// 创建模拟传输层
	mockTransp := newMockTransport()

	// 创建模拟构建器
	resourceBuilder := newMockResourceBuilder()
	promptBuilder := newMockPromptBuilder()
	toolBuilder := newMockToolBuilder()

	// 自定义服务器信息
	customInfo := protocol.Implementation{
		Name:    "custom-server",
		Version: "1.2.3",
	}

	// 创建服务器，带有自定义选项
	server, cleanup, err := NewServer(
		mockTransp,
		WithLogger(logger.DefaultLog),
		WithServerInfo(customInfo),
		WithInstructions("Custom instructions"),
		WithResourceBuilder(resourceBuilder),
		WithPromptBuilder(promptBuilder),
		WithToolBuilder(toolBuilder),
	)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer cleanup()

	// 验证自定义选项是否正确应用
	if server.options.serverInfo.Name != "custom-server" {
		t.Errorf("Expected server name 'custom-server', got '%s'", server.options.serverInfo.Name)
	}

	if server.options.serverInfo.Version != "1.2.3" {
		t.Errorf("Expected server version '1.2.3', got '%s'", server.options.serverInfo.Version)
	}

	if server.options.instructions != "Custom instructions" {
		t.Errorf("Expected instructions 'Custom instructions', got '%s'", server.options.instructions)
	}

	// 验证构建器是否正确设置
	if server.options.resourceBuilder != resourceBuilder {
		t.Error("Resource builder not set correctly")
	}

	if server.options.promptBuilder != promptBuilder {
		t.Error("Prompt builder not set correctly")
	}

	if server.options.toolBuilder != toolBuilder {
		t.Error("Tool builder not set correctly")
	}
}

// 测试服务器运行
func TestServerRun(t *testing.T) {
	// 创建模拟传输层
	mockTransp := newMockTransport()
	pr, pw := io.Pipe()
	go func() {
		_, _ = pw.Write(nil)
	}()
	mockTransp.SetReaderWriter(pr, mockTransp.writer)

	// 创建服务器
	server, cleanup, err := NewServer(mockTransp)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer cleanup()

	// 设置一个简单的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	errChan := make(chan error, 1)

	go func() {
		// 运行服务器
		err = server.Run(ctx)
		if err != nil {
			errChan <- err
		}
	}()
	time.Sleep(time.Millisecond * 50)

	select {
	case err := <-errChan:
		t.Fatalf("Server run failed: %v", err)
	default:
	}
	// 验证传输层的Run方法是否被调用
	if !mockTransp.IsRunCalled() {
		t.Error("Transport.Run was not called")
	}
}

// 测试服务器运行，传输层返回错误
func TestServerRunWithTransportError(t *testing.T) {
	// 创建模拟传输层
	mockTransp := newMockTransport()

	// 设置传输层返回错误
	expectedErr := io.EOF
	mockTransp.SetRunError(expectedErr)

	// 创建服务器
	server, cleanup, err := NewServer(mockTransp)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer cleanup()

	// 设置一个简单的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 运行服务器，应该返回传输层的错误
	err = server.Run(ctx)
	if err != expectedErr {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}
}

// 测试服务器的Logger方法
func TestServerLogger(t *testing.T) {
	// 创建模拟传输层
	mockTransp := newMockTransport()

	// 创建自定义日志记录器
	customLogger := &testLogger{}

	// 创建服务器，带有自定义日志记录器
	server, cleanup, err := NewServer(
		mockTransp,
		WithLogger(customLogger),
	)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer cleanup()

	// 验证日志记录器是否正确设置
	logHelper := server.Logger()
	if logHelper == nil {
		t.Fatal("Logger helper is nil")
	}
	ctx := context.Background()
	// 使用日志记录器
	logHelper.Debugf(ctx, "Test debug message")
	logHelper.Infof(ctx, "Test info message")
	logHelper.Warnf(ctx, "Test warning message")
	logHelper.Errorf(ctx, "Test error message")

	// 验证日志记录器是否收到消息
	if !strings.Contains(customLogger.messages.String(), "Test debug message") {
		t.Error("Debug message not logged")
	}

	if !strings.Contains(customLogger.messages.String(), "Test info message") {
		t.Error("Info message not logged")
	}

	if !strings.Contains(customLogger.messages.String(), "Test warning message") {
		t.Error("Warning message not logged")
	}

	if !strings.Contains(customLogger.messages.String(), "Test error message") {
		t.Error("Error message not logged")
	}
}

// 测试用的简单日志记录器
type testLogger struct {
	messages bytes.Buffer
}

func (l *testLogger) Logf(_ context.Context, level logger.Level, message string, args ...interface{}) {
	l.messages.WriteString(fmt.Sprintf("[%s] %s\n", level, fmt.Sprintf(message, args...)))
}
