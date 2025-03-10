package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mcp4go/mcp4go/protocol"
	"github.com/mcp4go/mcp4go/protocol/jsonschema"
	"github.com/mcp4go/mcp4go/server"
	"github.com/mcp4go/mcp4go/server/iface"
)

type WeatherResourceBuilder struct{}

func NewWeatherResourceBuilder() *WeatherResourceBuilder {
	return &WeatherResourceBuilder{}
}

func (x *WeatherResourceBuilder) Build() iface.IResource {
	return NewWeatherResource()
}

func (x *WeatherResourceBuilder) Subscribe() bool {
	return true
}

func (x *WeatherResourceBuilder) ListChanged() bool {
	return false
}

// WeatherResource 实现了 IResource 接口，提供天气数据资源
type WeatherResource struct{}

// AccessList 返回可用的天气资源模板列表
func (w *WeatherResource) AccessList() []protocol.ResourceTemplate {
	return []protocol.ResourceTemplate{
		{
			URITemplate: "weather:city/{city}",
			Name:        "City Weather",
			Description: "Get weather information for a specific city",
			MimeType:    "application/json",
		},
	}
}

// List 返回可用的天气资源列表
func (w *WeatherResource) List(ctx context.Context, cursor string) ([]protocol.Resource, string, error) {
	// 这里可以返回一些示例城市的天气资源
	resources := []protocol.Resource{
		{
			URI:         "weather:city/beijing",
			Name:        "Beijing Weather",
			Description: "Current weather in Beijing",
			MimeType:    "application/json",
		},
		{
			URI:         "weather:city/shanghai",
			Name:        "Shanghai Weather",
			Description: "Current weather in Shanghai",
			MimeType:    "application/json",
		},
		{
			URI:         "weather:city/guangzhou",
			Name:        "Guangzhou Weather",
			Description: "Current weather in Guangzhou",
			MimeType:    "application/json",
		},
	}

	return resources, "", nil
}

// Query 查询特定URI的天气资源内容
func (w *WeatherResource) Query(ctx context.Context, uri string) ([]protocol.ResourceContent, error) {
	// 在实际应用中，这里应该调用外部API获取真实天气数据
	// 这里简单模拟一些天气数据
	weatherData := map[string]interface{}{
		"temperature": 25,
		"humidity":    70,
		"conditions":  "Sunny",
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(weatherData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal weather data: %w", err)
	}

	return []protocol.ResourceContent{
		protocol.NewTextResourceContent(uri, "application/json", string(jsonData)),
	}, nil
}

// Watch 开始监视特定URI的天气资源更新
func (w *WeatherResource) Watch(ctx context.Context, uri string, ch chan<- protocol.ResourceUpdatedNotification) error {
	// 在实际应用中，这里应该实现一个定时任务来更新天气数据
	// 简单模拟每10秒更新一次天气
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 发送资源更新通知
				ch <- protocol.ResourceUpdatedNotification{
					URI: uri,
				}
			}
		}
	}()

	return nil
}

// CloseWatch 关闭对特定URI的监视
func (w *WeatherResource) CloseWatch(ctx context.Context, uri string) error {
	// 这里应该清理与URI相关的监视资源
	return nil
}

// StartWatchListChanged 开始监视资源列表变更
func (w *WeatherResource) StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.ResourceListChangedNotification) error {
	// 在当前示例中不支持资源列表变更，直接返回nil
	return nil
}

type WeatherPromptBuilder struct{}

func NewWeatherPromptBuilder() *WeatherPromptBuilder {
	return &WeatherPromptBuilder{}
}

func (x *WeatherPromptBuilder) Build() iface.IPrompt {
	return NewWeatherPrompt()
}

func (x *WeatherPromptBuilder) ListChanged() bool {
	return false
}

// WeatherPrompt 实现了 IPrompt 接口，提供与天气相关的提示
type WeatherPrompt struct{}

// List 返回可用的天气提示列表
func (p *WeatherPrompt) List(ctx context.Context, cursor string) ([]protocol.Prompt, string, error) {
	return []protocol.Prompt{
		{
			Name:        "weather_report",
			Description: "Generate a weather report for a specific city",
			Arguments: []protocol.PromptArgument{
				{
					Name:        "city",
					Description: "The name of the city",
					Required:    true,
				},
			},
		},
	}, "", nil
}

// Get 获取特定名称和参数的天气提示
func (p *WeatherPrompt) Get(ctx context.Context, name string, arguments map[string]string) (string, []protocol.PromptMessage, error) {
	if name != "weather_report" {
		return "", nil, fmt.Errorf("unknown prompt: %s", name)
	}

	city, ok := arguments["city"]
	if !ok {
		return "", nil, fmt.Errorf("missing required argument: city")
	}

	// 生成提示信息
	description := fmt.Sprintf("Weather report for %s", city)
	messages := []protocol.PromptMessage{
		{
			Role: protocol.RoleSystem,
			Content: protocol.NewTextContent(
				fmt.Sprintf("You are a weather reporter for %s. Provide a detailed weather report based on the current conditions.", city),
				nil,
			),
		},
		{
			Role: protocol.RoleUser,
			Content: protocol.NewTextContent(
				fmt.Sprintf("Please provide today's weather report for %s.", city),
				nil,
			),
		},
	}

	return description, messages, nil
}

// StartWatchListChanged 开始监视提示列表变更
func (p *WeatherPrompt) StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.PromptListChangedNotification) error {
	// 在当前示例中不支持提示列表变更，直接返回nil
	return nil
}

type WeatherToolBuilder struct{}

func NewWeatherToolBuilder() *WeatherToolBuilder {
	return &WeatherToolBuilder{}
}

func (x *WeatherToolBuilder) Build() iface.ITool {
	return NewWeatherTool()
}

func (x *WeatherToolBuilder) ListChanged() bool {
	return false
}

// WeatherTool 实现了 ITool 接口，提供与天气相关的工具
type WeatherTool struct{}

// List 返回可用的天气工具列表
func (t *WeatherTool) List(ctx context.Context, cursor string) ([]protocol.Tool, string, error) {
	return []protocol.Tool{
		{
			Name:        "get_weather",
			Description: "Get the current weather for a specific city",
			InputSchema: &jsonschema.Definition{
				Type: "object",
				Properties: map[string]jsonschema.Definition{
					"city": {
						Type:        "string",
						Description: "The name of the city",
					},
				},
				Required: []string{"city"},
			},
		},
		{
			Name:        "get_forecast",
			Description: "Get a 5-day weather forecast for a specific city",
			InputSchema: &jsonschema.Definition{
				Type: "object",
				Properties: map[string]jsonschema.Definition{
					"city": {
						Type:        jsonschema.String,
						Description: "The name of the city",
					},
					"days": {
						Type:        jsonschema.Integer,
						Description: "Number of days for the forecast (max 5)",
					},
				},
				Required: []string{"city"},
			},
		},
	}, "", nil
}

type WeatherArgs struct {
	City string `json:"city"`
	Days int    `json:"days"`
}

// Call 调用特定名称和参数的天气工具
func (t *WeatherTool) Call(ctx context.Context, name string, arguments json.RawMessage) ([]protocol.Content, error) {
	var dst WeatherArgs
	err := json.Unmarshal(arguments, &dst)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}
	switch name {
	case "get_weather":
		city := dst.City

		// 在实际应用中，这里应该调用外部API获取真实天气数据
		// 这里简单模拟返回一些天气数据
		return []protocol.Content{
			protocol.NewTextContent(fmt.Sprintf("Current weather in %s: 25°C, Sunny, Humidity: 70%%", city), nil),
		}, nil

	case "get_forecast":
		city := dst.City

		days := 3 // 默认值
		if dst.Days > 0 {
			days = dst.Days
			if days > 5 {
				days = 5 // 最多5天
			}
		}

		// 模拟天气预报数据
		var forecast string
		for i := 0; i < days; i++ {
			date := time.Now().AddDate(0, 0, i).Format("2006-01-02")
			temp := 20 + i // 简单模拟温度变化
			forecast += fmt.Sprintf("%s: %d°C, %s\n", date, temp, []string{"Sunny", "Cloudy", "Rainy"}[i%3])
		}

		return []protocol.Content{
			protocol.NewTextContent(fmt.Sprintf("Weather forecast for %s:\n%s", city, forecast), nil),
		}, nil

	default:
		return []protocol.Content{
			protocol.NewTextContent(fmt.Sprintf("Error: unknown tool: %s", name), nil),
		}, fmt.Errorf("unknown tool: %s", name)
	}
}

// StartWatchListChanged 开始监视工具列表变更
func (t *WeatherTool) StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.ToolListChangedNotification) error {
	// 在当前示例中不支持工具列表变更，直接返回nil
	return nil
}

// NewWeatherResource 创建一个新的 WeatherResource 实例
func NewWeatherResource() *WeatherResource {
	return &WeatherResource{}
}

// NewWeatherPrompt 创建一个新的 WeatherPrompt 实例
func NewWeatherPrompt() *WeatherPrompt {
	return &WeatherPrompt{}
}

// NewWeatherTool 创建一个新的 WeatherTool 实例
func NewWeatherTool() *WeatherTool {
	return &WeatherTool{}
}

// WeatherServer 包含 MCP 服务器的配置和组件
type WeatherServer struct {
	Resource *WeatherResource
	Prompt   *WeatherPrompt
	Tool     *WeatherTool
	Server   *server.Server
}

// NewWeatherServer 创建一个新的 WeatherServer 实例
func NewWeatherServer() *WeatherServer {
	return &WeatherServer{
		Resource: NewWeatherResource(),
		Prompt:   NewWeatherPrompt(),
		Tool:     NewWeatherTool(),
	}
}
