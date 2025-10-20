package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/rs/zerolog/log"

    "github.com/kyx-api-quota-bridge/backend/internal/app"
)

func main() {
    container, err := app.InitializeApplication()
    if err != nil {
        log.Fatal().Err(err).Msg("failed to initialize application")
    }

    srvCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    go func() {
        if err := container.HTTPServer.Start(); err != nil {
            log.Fatal().Err(err).Msg("http server crashed")
        }
    }()

    <-srvCtx.Done()

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := container.Shutdown(shutdownCtx); err != nil {
        log.Error().Err(err).Msg("graceful shutdown failed")
        os.Exit(1)
    }

    log.Info().Msg("server stopped")
}
