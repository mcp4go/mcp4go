package handlers

import (
	"github.com/google/wire"
	"github.com/mcp4go/mcp4go/server/internal/router"
)

var Provider = wire.NewSet(
	NewIHandlers,
	NewInitializedHandler, NewShutdownHandler, NewCancelHandler, NewProgressNotificationHandler,
	NewListPromptsHandler, NewGetPromptHandler,
	NewListResourcesHandler, NewReadResourceHandler, NewListResourceTemplatesHandler, NewSubscribeHandler, NewUnsubscribeHandler,
	NewListToolsHandler, NewCallToolHandler,
)

func NewIHandlers(
	initializeHandler *InitializeHandler,
	initializedHandler *InitializedHandler,
	shutdownHandler *ShutdownHandler,
	cancelHandler *CancelHandler,
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
	return []router.IHandler{
		initializeHandler,
		initializedHandler,
		shutdownHandler,
		cancelHandler,
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
