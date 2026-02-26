package main

import (
	"context"
	"flag"
	"os"

	"api-gateway/internal/adapter/config"
	"api-gateway/internal/server"

	"github.com/rs/zerolog"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger.Info().Str("config", *configPath).Msg("loading configuration")

	loader := config.NewViperLoader()
	cfg, err := loader.Load(context.Background(), *configPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	srv := server.New(cfg, logger)
	if err := srv.Start(); err != nil {
		logger.Fatal().Err(err).Msg("server failed")
	}
}
