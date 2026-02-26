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

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					v.Reload()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("config watcher error: %v\n", err)
			}
		}
	}()

	return nil
}

func (v *ViperLoader) Watch(callback func(*config.Config)) {
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
