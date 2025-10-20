package routes

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/cors"
    "github.com/rs/zerolog/log"

    "github.com/kyx-api-quota-bridge/backend/internal/config"
    "github.com/kyx-api-quota-bridge/backend/internal/http/handlers"
    "github.com/kyx-api-quota-bridge/backend/internal/http/middleware"
)

type Server struct {
    cfg      *config.Config
    router   chi.Router
    server   *http.Server
    handlers *handlers.Container
}

func Provide(cfg *config.Config, handlers *handlers.Container) (*Server, error) {
    router := chi.NewRouter()
    router.Use(cors.Handler(cors.Options{
        AllowedOrigins: []string{"*"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
        AllowCredentials: true,
    }))
    router.Use(middleware.RequestLogger)
    router.Use(middleware.Recoverer)

    handlers.RegisterRoutes(router)

    srv := &http.Server{
        Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
        Handler: router,
    }

    return &Server{
        cfg:      cfg,
        router:   router,
        server:   srv,
        handlers: handlers,
    }, nil
}

func (s *Server) Start() error {
    log.Info().Int("port", s.cfg.HTTPPort).Msg("starting http server")
    return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
    shutdownCtx, cancel := context.WithTimeout(ctx, s.cfg.ShutdownTimeout)
    defer cancel()
    return s.server.Shutdown(shutdownCtx)
}
