package router

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewIRouter, NewRouter)
