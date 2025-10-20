package app

import (
    "context"
    "fmt"

    "github.com/google/wire"

    "github.com/kyx-api-quota-bridge/backend/internal/config"
    "github.com/kyx-api-quota-bridge/backend/internal/http/routes"
    "github.com/kyx-api-quota-bridge/backend/internal/service"
    "github.com/kyx-api-quota-bridge/backend/internal/storage"
)

type Container struct {
    Config     *config.Config
    HTTPServer *routes.Server
}

func (c *Container) Shutdown(ctx context.Context) error {
    return c.HTTPServer.Shutdown(ctx)
}

func newContainer(cfg *config.Config, srv *routes.Server) (*Container, error) {
    return &Container{
        Config:     cfg,
        HTTPServer: srv,
    }, nil
}

var appSet = wire.NewSet(
    config.Provide,
    storage.Provide,
    service.Provide,
    routes.Provide,
    newContainer,
)

func InitializeApplication() (*Container, error) {
    wire.Build(appSet)
    return nil, fmt.Errorf("wire build failed")
}
