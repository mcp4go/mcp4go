package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mcp4go/mcp4go/pkg/logger"
	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/server/iface"
)

// 模拟处理程序，用于测试
type mockHandler struct {
	method     protocol.McpMethod
	handleFunc func(ctx context.Context, message json.RawMessage) (json.RawMessage, error)
	handleErr  error
}

func (m *mockHandler) Handle(ctx context.Context, message json.RawMessage) (json.RawMessage, error) {
	if m.handleFunc != nil {
		return m.handleFunc(ctx, message)
	}
	if m.handleErr != nil {
		return nil, m.handleErr
	}
	return json.RawMessage(`{"result":"ok"}`), nil
}

func (m *mockHandler) Method() protocol.McpMethod {
	return m.method
}

// 模拟事件总线，用于测试
type mockEventBus struct {
	iface.EventBus
	ResourceUpdatedNotificationSubscribers map[string][]chan<- protocol.ResourceUpdatedNotification
	ResourceListChangedSubscribers         map[string][]chan<- protocol.ResourceListChangedNotification
	ToolListChangedSubscribers             map[string][]chan<- protocol.ToolListChangedNotification
	PromptListChangedSubscribers           map[string][]chan<- protocol.PromptListChangedNotification
	mutex                                  sync.RWMutex
}

func newMockEventBus() *mockEventBus {
	return &mockEventBus{
		EventBus:                               iface.NewEventBus(),
		ResourceUpdatedNotificationSubscribers: make(map[string][]chan<- protocol.ResourceUpdatedNotification),
		ResourceListChangedSubscribers:         make(map[string][]chan<- protocol.ResourceListChangedNotification),
		ToolListChangedSubscribers:             make(map[string][]chan<- protocol.ToolListChangedNotification),
		PromptListChangedSubscribers:           make(map[string][]chan<- protocol.PromptListChangedNotification),
	}
}

func (m *mockEventBus) PublishResourceUpdated(_ context.Context, notification protocol.ResourceUpdatedNotification) {
	m.ResourceUpdatedNotificationChan <- notification
}

func (m *mockEventBus) PublishResourceListChanged(_ context.Context, notification protocol.ResourceListChangedNotification) {
	m.ResourceListChangedNotificationChan <- notification
}

func (m *mockEventBus) PublishToolListChanged(_ context.Context, notification protocol.ToolListChangedNotification) {
	m.ToolListChangedNotificationChan <- notification
}

func (m *mockEventBus) PublishPromptListChanged(_ context.Context, notification protocol.PromptListChangedNotification) {
	m.PromptListChangedNotificationChan <- notification
}

func (m *mockEventBus) PublishLogMessage(_ context.Context, notification protocol.LoggingMessageNotification) {
	m.LoggingMessageNotificationChan <- notification
}

func (m *mockEventBus) SubscribeResourceUpdated(_ context.Context, uri string, ch chan<- protocol.ResourceUpdatedNotification) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ResourceUpdatedNotificationSubscribers[uri] = append(m.ResourceUpdatedNotificationSubscribers[uri], ch)
}

func (m *mockEventBus) SubscribeResourceListChanged(_ context.Context, path string, ch chan<- protocol.ResourceListChangedNotification) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ResourceListChangedSubscribers[path] = append(m.ResourceListChangedSubscribers[path], ch)
}

func (m *mockEventBus) SubscribeToolListChanged(_ context.Context, path string, ch chan<- protocol.ToolListChangedNotification) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ToolListChangedSubscribers[path] = append(m.ToolListChangedSubscribers[path], ch)
}

func (m *mockEventBus) SubscribePromptListChanged(_ context.Context, path string, ch chan<- protocol.PromptListChangedNotification) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.PromptListChangedSubscribers[path] = append(m.PromptListChangedSubscribers[path], ch)
}

func (m *mockEventBus) UnsubscribeResourceUpdated(_ context.Context, uri string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.ResourceUpdatedNotificationSubscribers, uri)
}

func (m *mockEventBus) UnsubscribeResourceListChanged(_ context.Context, path string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.ResourceListChangedSubscribers, path)
}

func (m *mockEventBus) UnsubscribeToolListChanged(_ context.Context, path string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.ToolListChangedSubscribers, path)
}

func (m *mockEventBus) UnsubscribePromptListChanged(_ context.Context, path string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.PromptListChangedSubscribers, path)
}

// 测试创建路由器
func TestNewRouter(t *testing.T) {
	// 创建模拟处理程序
	handlers := []IHandler{
		&mockHandler{method: "test/method1"},
		&mockHandler{method: "test/method2"},
	}

	// 创建模拟事件总线
	bus := newMockEventBus()

	// 创建路由器
	router, err := NewRouter(handlers, bus.EventBus, logger.DefaultLog)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	// 验证路由器是否正确初始化
	if router == nil {
		t.Fatal("Router is nil")
	}

	// 验证处理程序是否正确注册
	if len(router.handlers) != 3 {
		t.Errorf("Expected 3 handlers, got %d", len(router.handlers))
	}

	// 验证处理程序映射是否正确
	if _, ok := router.handlers["test/method1"]; !ok {
		t.Error("Handler for 'test/method1' not registered")
	}

	if _, ok := router.handlers["test/method2"]; !ok {
		t.Error("Handler for 'test/method2' not registered")
	}
}

// 测试路由器处理请求
func TestRouterHandle(t *testing.T) {
	// 创建模拟处理程序，返回特定结果
	handlers := []IHandler{
		&mockHandler{
			method: "test/method",
			handleFunc: func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return json.RawMessage(`{"result":"success"}`), nil
			},
		},
	}

	// 创建模拟事件总线
	bus := newMockEventBus()

	// 创建路由器
	router, err := NewRouter(handlers, bus.EventBus, logger.DefaultLog)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	// 创建测试请求
	req := &protocol.JsonrpcRequest{
		Jsonrpc: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "test/method",
		Params:  json.RawMessage(`{"param":"value"}`),
	}

	// 调用处理方法
	result, err := router.handle(context.Background(), req)
	if err != nil {
		t.Fatalf("Router.handle failed: %v", err)
	}

	// 验证结果
	expected := `{"result":"success"}`
	if string(result) != expected {
		t.Errorf("Expected result '%s', got '%s'", expected, string(result))
	}
}

// 测试路由器处理未找到的方法
func TestRouterHandleNotFound(t *testing.T) {
	// 创建空的处理程序列表
	handlers := []IHandler{}

	// 创建模拟事件总线
	bus := newMockEventBus()

	// 创建路由器
	router, err := NewRouter(handlers, bus.EventBus, logger.DefaultLog)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	// 创建测试请求，请求一个不存在的方法
	req := &protocol.JsonrpcRequest{
		Jsonrpc: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "test/nonexistent",
		Params:  json.RawMessage(`{}`),
	}

	// 调用处理方法，应该返回方法未找到的错误
	_, err = router.handle(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for non-existent method, got nil")
	}

	// 验证错误消息
	expected := "method(test/nonexistent) not found"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

// 测试路由器处理错误
func TestRouterHandleError(t *testing.T) {
	// 创建模拟处理程序，返回错误
	expectedErr := errors.New("test error")
	handlers := []IHandler{
		&mockHandler{
			method:    "test/error",
			handleErr: expectedErr,
		},
	}

	// 创建模拟事件总线
	bus := newMockEventBus()

	// 创建路由器
	router, err := NewRouter(handlers, bus.EventBus, logger.DefaultLog)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	// 创建测试请求
	req := &protocol.JsonrpcRequest{
		Jsonrpc: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "test/error",
		Params:  json.RawMessage(`{}`),
	}

	// 调用处理方法，应该返回处理程序的错误
	_, err = router.handle(context.Background(), req)
	if err != expectedErr {
		t.Fatalf("Expected error '%v', got '%v'", expectedErr, err)
	}
}

// 测试路由器的Handle方法，模拟完整的读写循环
func TestRouterHandleWithReadWrite(t *testing.T) {
	// 创建模拟处理程序
	handlers := []IHandler{
		&mockHandler{
			method: "test/method",
			handleFunc: func(_ context.Context, _ json.RawMessage) (json.RawMessage, error) {
				return json.RawMessage(`{"name":"test-result"}`), nil
			},
		},
	}

	// 创建模拟事件总线
	bus := newMockEventBus()

	// 创建路由器
	router, err := NewRouter(handlers, bus.EventBus, logger.DefaultLog)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	// 创建测试请求
	requestJSON := `{"jsonrpc":"2.0","id":1,"method":"test/method","params":{}}`
	preader, pwriter := io.Pipe()
	go func() {
		_, _ = pwriter.Write([]byte(requestJSON))
	}()
	writer := &saveBuf{}

	// 创建上下文，带有超时
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// 启动一个goroutine运行Handle方法
	errorCh := make(chan error, 1)
	go func() {
		errorCh <- router.Handle(ctx, preader, writer)
	}()

	// 等待一段时间，让响应被写入
	time.Sleep(200 * time.Millisecond)

	// 取消上下文，结束处理
	cancel()

	// 等待Handle方法返回
	select {
	case err := <-errorCh:
		if err != nil && errors.Is(err, context.Canceled) {
			t.Fatalf("Router.Handle returned unexpected error: %v", err)
		}
	case <-ctx.Done():
	default:
	}

	// 验证响应
	response := writer.String()
	if !strings.Contains(response, `"result":{"name":"test-result"}`) {
		t.Errorf("Expected response to contain result, got: %s", response)
	}
}

// 测试路由器处理事件总线通知
func TestRouterHandleEventBusNotifications(t *testing.T) {
	// 创建模拟处理程序
	handlers := []IHandler{}

	// 创建模拟事件总线
	bus := newMockEventBus()

	// 创建路由器
	router, err := NewRouter(handlers, bus.EventBus, logger.DefaultLog)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	// 创建空的读取器和写入器
	writer := &saveBuf{}

	// 创建上下文，带有超时
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	preader, pwriter := io.Pipe()
	go func() {
		_, _ = pwriter.Write([]byte(""))
	}()
	// 启动一个goroutine运行Handle方法
	errorCh := make(chan error, 1)
	go func() {
		errorCh <- router.Handle(ctx, preader, writer)
	}()

	// 发送资源更新通知
	resourceNotification := protocol.ResourceUpdatedNotification{
		URI: "file://test.txt",
	}
	bus.PublishResourceUpdated(ctx, resourceNotification)

	// 发送工具列表更改通知
	toolsNotification := protocol.ToolListChangedNotification{}
	bus.PublishToolListChanged(ctx, toolsNotification)

	// 等待一段时间，让通知被处理并响应被写入
	time.Sleep(200 * time.Millisecond)

	// 取消上下文，结束处理
	cancel()

	// 等待Handle方法返回
	select {
	case err := <-errorCh:
		if err != nil && errors.Is(err, context.Canceled) {
			t.Fatalf("Router.Handle returned unexpected error: %v", err)
		}
	case <-ctx.Done():
	default:
	}

	// 验证响应包含通知
	response := writer.String()
	if !strings.Contains(response, `"file://test.txt"`) {
		t.Errorf("Expected response to contain resource notification, got: %s", response)
	}
}

// 测试NewIRouter函数
func TestNewIRouter(t *testing.T) {
	// 创建模拟处理程序和事件总线
	handlers := []IHandler{}
	bus := newMockEventBus()

	// 创建路由器
	router, err := NewRouter(handlers, bus.EventBus, logger.DefaultLog)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	// 通过NewIRouter创建接口实例
	iRouter := NewIRouter(router)

	// 验证返回的接口实例是否正确
	if iRouter == nil {
		t.Fatal("IRouter is nil")
	}
}

type saveBuf struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (x *saveBuf) Write(p []byte) (n int, err error) {
	x.mu.Lock()
	defer x.mu.Unlock()

	n, err = x.buf.Write(p)
	if err != nil {
		return
	}
	return n, nil
}

func (x *saveBuf) String() string {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.buf.String()
}
