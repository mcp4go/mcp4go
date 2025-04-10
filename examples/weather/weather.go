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

// WeatherResource implements the IResource interface, providing weather data resources
type WeatherResource struct{}

// AccessList returns a list of available weather resource templates
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

// List returns a list of available weather resources
func (w *WeatherResource) List(ctx context.Context, cursor string) ([]protocol.Resource, string, error) {
	// This can return weather resources for some example cities
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

// Query queries the content of a specific URI weather resource
func (w *WeatherResource) Query(ctx context.Context, uri string) ([]protocol.ResourceContent, error) {
	// In a real application, this should call an external API to get real weather data
	// Here, we simply simulate some weather data
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

// Watch starts monitoring updates for a specific URI weather resource
func (w *WeatherResource) Watch(ctx context.Context, uri string, ch chan<- protocol.ResourceUpdatedNotification) error {
	// In a real application, this should implement a scheduled task to update weather data
	// Simply simulate weather updates every 10 seconds
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Send resource update notification
				ch <- protocol.ResourceUpdatedNotification{
					URI: uri,
				}
			}
		}
	}()

	return nil
}

// CloseWatch closes monitoring for a specific URI
func (w *WeatherResource) CloseWatch(ctx context.Context, uri string) error {
	// This should clean up resources related to the URI monitoring
	return nil
}

// StartWatchListChanged starts monitoring for resource list changes
func (w *WeatherResource) StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.ResourceListChangedNotification) error {
	// In the current example, resource list changes are not supported, directly return nil
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

// WeatherPrompt implements the IPrompt interface, providing weather-related prompts
type WeatherPrompt struct{}

// List returns a list of available weather prompts
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

// Get gets a specific prompt by name and arguments
func (p *WeatherPrompt) Get(ctx context.Context, name string, arguments map[string]string) (string, []protocol.PromptMessage, error) {
	if name != "weather_report" {
		return "", nil, fmt.Errorf("unknown prompt: %s", name)
	}

	city, ok := arguments["city"]
	if !ok {
		return "", nil, fmt.Errorf("missing required argument: city")
	}

	// Generate prompt information
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

// StartWatchListChanged starts monitoring for prompt list changes
func (p *WeatherPrompt) StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.PromptListChangedNotification) error {
	// In the current example, prompt list changes are not supported, directly return nil
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

// WeatherTool implements the ITool interface, providing weather-related tools
type WeatherTool struct{}

// List returns a list of available weather tools
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

// Call calls a specific weather tool with the given name and arguments
func (t *WeatherTool) Call(ctx context.Context, name string, arguments json.RawMessage) ([]protocol.Content, error) {
	var dst WeatherArgs
	err := json.Unmarshal(arguments, &dst)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}
	switch name {
	case "get_weather":
		city := dst.City

		// In a real application, this should call an external API to get real weather data
		// Here, we simply return some simulated weather data
		return []protocol.Content{
			protocol.NewTextContent(fmt.Sprintf("Current weather in %s: 25°C, Sunny, Humidity: 70%%", city), nil),
		}, nil

	case "get_forecast":
		city := dst.City

		days := 3 // Default value
		if dst.Days > 0 {
			days = dst.Days
			if days > 5 {
				days = 5 // Maximum 5 days
			}
		}

		// Simulate weather forecast data
		var forecast string
		for i := 0; i < days; i++ {
			date := time.Now().AddDate(0, 0, i).Format("2006-01-02")
			temp := 20 + i // Simple temperature simulation
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

// StartWatchListChanged starts monitoring for tool list changes
func (t *WeatherTool) StartWatchListChanged(ctx context.Context, uri string, ch chan<- protocol.ToolListChangedNotification) error {
	// In the current example, tool list changes are not supported, directly return nil
	return nil
}

// NewWeatherResource creates a new WeatherResource instance
func NewWeatherResource() *WeatherResource {
	return &WeatherResource{}
}

// NewWeatherPrompt creates a new WeatherPrompt instance
func NewWeatherPrompt() *WeatherPrompt {
	return &WeatherPrompt{}
}

// NewWeatherTool creates a new WeatherTool instance
func NewWeatherTool() *WeatherTool {
	return &WeatherTool{}
}

// WeatherServer contains MCP server configuration and components
type WeatherServer struct {
	Resource *WeatherResource
	Prompt   *WeatherPrompt
	Tool     *WeatherTool
	Server   *server.Server
}

// NewWeatherServer creates a new WeatherServer instance
func NewWeatherServer() *WeatherServer {
	return &WeatherServer{
		Resource: NewWeatherResource(),
		Prompt:   NewWeatherPrompt(),
		Tool:     NewWeatherTool(),
	}
}
