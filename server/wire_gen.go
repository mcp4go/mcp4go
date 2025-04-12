// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package server

import (
	"github.com/mcp4go/mcp4go/pkg/logger"
	"github.com/mcp4go/mcp4go/server/iface"
	"github.com/mcp4go/mcp4go/server/internal/handlers"
	"github.com/mcp4go/mcp4go/server/internal/router"
)

// Injectors from wire.go:

func initRouter(iLogger logger.ILogger, initializeHandler *handlers.InitializeHandler, setLevelHandler *handlers.SetLevelHandler, iResource iface.IResource, iPrompt iface.IPrompt, iTool iface.ITool, eventBus iface.EventBus) (router.IRouter, error) {
	initializedHandler := handlers.NewInitializedHandler()
	listPromptsHandler := handlers.NewListPromptsHandler(iPrompt)
	getPromptHandler := handlers.NewGetPromptHandler(iPrompt)
	listResourcesHandler := handlers.NewListResourcesHandler(iResource)
	readResourceHandler := handlers.NewReadResourceHandler(iResource)
	listResourceTemplatesHandler := handlers.NewListResourceTemplatesHandler(iResource)
	subscribeHandler := handlers.NewSubscribeHandler(iResource, eventBus)
	unsubscribeHandler := handlers.NewUnsubscribeHandler(iResource)
	listToolsHandler := handlers.NewListToolsHandler(iTool)
	callToolHandler := handlers.NewCallToolHandler(iTool)
	v := handlers.NewIHandlers(initializeHandler, initializedHandler, setLevelHandler, listPromptsHandler, getPromptHandler, listResourcesHandler, readResourceHandler, listResourceTemplatesHandler, subscribeHandler, unsubscribeHandler, listToolsHandler, callToolHandler)
	routerRouter, err := router.NewRouter(v, eventBus, iLogger)
	if err != nil {
		return nil, err
	}
	iRouter := router.NewIRouter(routerRouter)
	return iRouter, nil
}
