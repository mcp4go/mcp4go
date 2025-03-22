//go:build wireinject

package server

import (
	"github.com/google/wire"
	"github.com/mcp4go/mcp4go/pkg/logger"
	"github.com/mcp4go/mcp4go/server/iface"
	"github.com/mcp4go/mcp4go/server/internal/handlers"
	"github.com/mcp4go/mcp4go/server/internal/router"
)

func initRouter(logger.ILogger, *handlers.InitializeHandler, *handlers.SetLevelHandler, iface.IResource, iface.IPrompt, iface.ITool, iface.EventBus) (router.IRouter, error) {
	panic(wire.Build(router.ProviderSet, handlers.Provider))
}
