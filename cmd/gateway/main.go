package main

import (
	"context"
	"errors"
	"flag"
	"os"

	adapterconfig "api-gateway/internal/adapter/config"
	domainconfig "api-gateway/internal/domain/config"
	"api-gateway/internal/server"

	"github.com/rs/zerolog"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger.Info().Str("config", *configPath).Msg("loading configuration")

	loader := adapterconfig.NewViperLoader()
	_, err := loader.Load(context.Background(), *configPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	reloadCh := make(chan struct{}, 1)
	loader.Watch(func(_ *domainconfig.Config) {
		select {
		case reloadCh <- struct{}{}:
		default:
		}
	})

	for {
		cfg := loader.Get()
		if cfg == nil {
			logger.Fatal().Msg("configuration is not available")
		}

		srv := server.New(cfg, logger)
		err = srv.Start(reloadCh)
		if errors.Is(err, server.ErrReloadRequested) {
			logger.Info().Msg("configuration reloaded")
			continue
		}
		if err != nil {
			logger.Fatal().Err(err).Msg("server failed")
		}
		break
	}
}
