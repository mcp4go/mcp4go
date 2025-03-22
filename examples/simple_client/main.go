package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mcp4go/mcp4go/client"
	"github.com/mcp4go/mcp4go/client/transport"
)

func main() {
	// Set up logger
	log.SetOutput(os.Stdout)
	log.SetPrefix("[MCP4Go-Client] ")

	const (
		argNum = 2
	)

	// Parse command line arguments
	if len(os.Args) < argNum {
		log.Fatalf("Usage: %s <command> [args...]", os.Args[0])
	}

	command := os.Args[1]
	args := os.Args[argNum:]

	// Create a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		log.Println("Received termination signal, shutting down...")
		cancel()
		// Force exit after a timeout
		time.AfterFunc(5*time.Second, func() {
			log.Println("Forced shutdown after timeout")
			os.Exit(1)
		})
	}()

	// Create transport
	log.Println("command====", command)
	t := transport.NewStdioTransport(transport.WithCommand(command, args...))

	// Create client with options
	c, cleanup, err := client.NewClient(
		t,
		client.WithClientInfo("mcp4go-simple-client", "0.1.0"),
		client.WithRootsCapability(true),
		client.WithSamplingCapability(),
	)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
		return
	}
	defer cleanup()

	// Connect to the server
	log.Println("Connecting to MCP server...")
	if err := c.Connect(ctx); err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer c.Close()

	// Get server info
	serverInfo := c.ServerInfo()
	log.Printf("Connected to server: %s v%s", serverInfo.Name, serverInfo.Version)

	if c.Instructions() != "" {
		log.Printf("Server instructions: %s", c.Instructions())
	}

	// Get server capabilities
	caps := c.ServerCapabilities()
	log.Printf("Server capabilities:")
	if caps.Tools != nil {
		log.Printf("  - Tools support: enabled (list changed: %v)", caps.Tools.ListChanged)
	}
	if caps.Resources != nil {
		log.Printf("  - Resources support: enabled (subscribe: %v, list changed: %v)",
			caps.Resources.Subscribe, caps.Resources.ListChanged)
	}
	if caps.Prompts != nil {
		log.Printf("  - Prompts support: enabled (list changed: %v)", caps.Prompts.ListChanged)
	}
	if caps.Logging != nil {
		log.Printf("  - Logging support: enabled")
	}

	// List available tools
	if caps.Tools != nil {
		toolResult, err := c.ListTools(ctx)
		if err != nil {
			log.Printf("Error listing tools: %v", err)
		} else {
			log.Printf("Available tools (%d):", len(toolResult.Tools))
			for _, tool := range toolResult.Tools {
				log.Printf("  - %s: %s", tool.Name, tool.Description)
			}
		}
	}

	// List available resources
	if caps.Resources != nil {
		resourceResult, err := c.ListResources(ctx)
		if err != nil {
			log.Printf("Error listing resources: %v", err)
		} else {
			log.Printf("Available resources (%d):", len(resourceResult.Resources))
			for _, resource := range resourceResult.Resources {
				log.Printf("  - %s: %s (%s)", resource.URI, resource.Name, resource.Description)
			}
		}
	}

	// List available prompts
	if caps.Prompts != nil {
		promptResult, err := c.ListPrompts(ctx)
		if err != nil {
			log.Printf("Error listing prompts: %v", err)
		} else {
			log.Printf("Available prompts (%d):", len(promptResult.Prompts))
			for _, prompt := range promptResult.Prompts {
				log.Printf("  - %s: %s", prompt.Name, prompt.Description)
			}
		}
	}

	// Create examples of each helper
	toolHelper := client.NewToolHelper(c)
	resourceHelper := client.NewResourceHelper(c)
	promptHelper := client.NewPromptHelper(c)

	// Interactive loop - ask user what they want to do
	fmt.Println("\nEnter a command (or 'exit' to quit):")
	fmt.Println("  tool <name> <args_json> - Call a tool")
	fmt.Println("  resource <uri> - Read a resource")
	fmt.Println("  prompt <name> <args_json> - Get a prompt")
	fmt.Println("  exit - Exit the program")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Print("> ")
			var command, param1, param2 string
			_, _ = fmt.Scanf("%s %s %s", &command, &param1, &param2)

			switch command {
			case "exit":
				return

			case "tool":
				if param1 == "" {
					fmt.Println("Missing tool name")
					continue
				}

				// Use empty args if none provided
				args := map[string]interface{}{}
				if param2 != "" {
					// Try to parse args as JSON
					if err := json.Unmarshal([]byte(param2), &args); err != nil {
						fmt.Printf("Error parsing arguments: %v\n", err)
						continue
					}
				}

				result, err := toolHelper.CallWithTimeout(param1, args, 10*time.Second)
				if err != nil {
					fmt.Printf("Error calling tool: %v\n", err)
				} else {
					fmt.Printf("Tool result: %+v\n", result)
				}

			case "resource":
				if param1 == "" {
					fmt.Println("Missing resource URI")
					continue
				}

				text, err := resourceHelper.ReadTextContent(ctx, param1)
				if err != nil {
					fmt.Printf("Error reading resource: %v\n", err)
				} else {
					fmt.Printf("Resource content:\n%s\n", text)
				}

			case "prompt":
				if param1 == "" {
					fmt.Println("Missing prompt name")
					continue
				}

				// Use empty args if none provided
				var args map[string]string
				if param2 != "" {
					// Try to parse args as JSON
					if err := json.Unmarshal([]byte(param2), &args); err != nil {
						fmt.Printf("Error parsing arguments: %v\n", err)
						continue
					}
				}

				messages, err := promptHelper.GetPromptMessages(ctx, param1, args)
				if err != nil {
					fmt.Printf("Error getting prompt: %v\n", err)
				} else {
					fmt.Printf("Prompt messages (%d):\n", len(messages))
					for i, msg := range messages {
						fmt.Printf("  [%d] %s: %+v\n", i, msg.Role, msg.Content)
					}
				}

			default:
				fmt.Println("Unknown command")
			}
		}
	}
}
