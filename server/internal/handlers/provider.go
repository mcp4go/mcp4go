package handlers

import (
	"encoding/json"

	"github.com/google/wire"

	"github.com/mcp4go/mcp4go/server/internal/router"
)

type RequestDecodeFunc func(data json.RawMessage, v any) error

var Provider = wire.NewSet(
	NewIHandlers,
	NewInitializedHandler,
	NewListPromptsHandler, NewGetPromptHandler,
	NewListResourcesHandler, NewReadResourceHandler, NewListResourceTemplatesHandler, NewSubscribeHandler, NewUnsubscribeHandler,
	NewListToolsHandler, NewCallToolHandler,
)

func NewIHandlers(
	initializeHandler *InitializeHandler,
	initializedHandler *InitializedHandler,
	setLevelHandler *SetLevelHandler,
	listPromptsHandler *ListPromptsHandler,
	getPromptHandler *GetPromptHandler,
	listResourcesHandler *ListResourcesHandler,
	readResourceHandler *ReadResourceHandler,
	listResourceTemplatesHandler *ListResourceTemplatesHandler,
	subscribeHandler *SubscribeHandler,
	unsubscribeHandler *UnsubscribeHandler,
	listToolsHandler *ListToolsHandler,
	callToolHandler *CallToolHandler,
) []router.IHandler {
	//nolint:whitespace
	return []router.IHandler{
		initializeHandler,
		initializedHandler,
		setLevelHandler,
		listPromptsHandler,
		getPromptHandler,
		listResourcesHandler,
		readResourceHandler,
		listResourceTemplatesHandler,
		subscribeHandler,
		unsubscribeHandler,
		listToolsHandler,
		callToolHandler,
	}
}
