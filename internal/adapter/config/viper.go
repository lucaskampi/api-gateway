package config

import (
	"context"
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"api-gateway/internal/domain/config"
)

type ViperLoader struct {
	mu      sync.RWMutex
	cfg     *config.Config
	watcher *fsnotify.Watcher
	path    string
}

func NewViperLoader() *ViperLoader {
	return &ViperLoader{}
}

func (v *ViperLoader) Load(ctx context.Context, path string) (*config.Config, error) {
	v.path = path

	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout_ms", 5000)
	viper.SetDefault("server.write_timeout_ms", 5000)
	viper.SetDefault("server.idle_timeout_ms", 60000)
	viper.SetDefault("cors.allow_origins", []string{"*"})
	viper.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"})
	viper.SetDefault("cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"})
	viper.SetDefault("cors.allow_credentials", false)
	viper.SetDefault("cors.expose_headers", []string{})
	viper.SetDefault("cors.max_age", 300)
	viper.SetDefault("global_rate_limit.key_by", "global")
	viper.SetDefault("routes.*.rate_limit.key_by", "ip")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := &config.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	v.mu.Lock()
	v.cfg = cfg
	v.mu.Unlock()

	if err := v.setupWatcher(); err != nil {
		return cfg, nil
	}

	return cfg, nil
}

func (v *ViperLoader) setupWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := watcher.Add(v.path); err != nil {
		watcher.Close()
		return err
	}

	v.watcher = watcher

	return nil
}

func (v *ViperLoader) Watch(callback func(*config.Config)) {
	if v.watcher == nil {
		return
	}

	go func() {
		for {
			select {
			case event, ok := <-v.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					cfg := v.Reload()
					if cfg != nil {
						callback(cfg)
					}
				}
			case err, ok := <-v.watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("config watcher error: %v\n", err)
			}
		}
	}()
}

func (v *ViperLoader) Get() *config.Config {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.cfg
}

func (v *ViperLoader) Reload() *config.Config {
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("failed to reload config: %v\n", err)
		return nil
	}

	cfg := &config.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		fmt.Printf("failed to unmarshal config: %v\n", err)
		return nil
	}

	v.mu.Lock()
	v.cfg = cfg
	v.mu.Unlock()

	return cfg
}
