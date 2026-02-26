package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway/internal/domain/config"
	"api-gateway/internal/middleware"
	"api-gateway/internal/router"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/rs/zerolog"
)

type Server struct {
	app    *fiber.App
	cfg    *config.Config
	logger zerolog.Logger
}

func New(cfg *config.Config, logger zerolog.Logger) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout(),
		WriteTimeout: cfg.Server.WriteTimeout(),
		IdleTimeout:  cfg.Server.IdleTimeout(),
		AppName:      "api-gateway",
	})

	app.Use(recover.New())

	return &Server{
		app:    app,
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Server) Start() error {
	if s.cfg.OTel.Endpoint != "" {
		if err := middleware.InitOTel(s.cfg.OTel.Endpoint, s.cfg.OTel.ServiceName); err != nil {
			s.logger.Warn().Err(err).Msg("failed to initialize OTel")
		}
	}

	r := router.New(s.app, s.cfg, s.logger)
	r.Setup()

	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.logger.Info().Str("addr", addr).Msg("starting server")

	go func() {
		if err := s.app.Listen(addr); err != nil {
			s.logger.Error().Err(err).Msg("server error")
		}
	}()

	return s.gracefulShutdown()
}

func (s *Server) gracefulShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	if err := middleware.ShutdownOTel(); err != nil {
		s.logger.Warn().Err(err).Msg("OTel shutdown error")
	}

	s.logger.Info().Msg("server stopped")
	return nil
}
