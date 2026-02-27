package config

import (
	"context"
	"time"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	OTel   OTelConfig   `mapstructure:"otel"`
	Routes []Route      `mapstructure:"routes"`
}

type ServerConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	ReadTimeoutMs  int    `mapstructure:"read_timeout_ms"`
	WriteTimeoutMs int    `mapstructure:"write_timeout_ms"`
	IdleTimeoutMs  int    `mapstructure:"idle_timeout_ms"`
}

func (s ServerConfig) ReadTimeout() time.Duration {
	return time.Duration(s.ReadTimeoutMs) * time.Millisecond
}

func (s ServerConfig) WriteTimeout() time.Duration {
	return time.Duration(s.WriteTimeoutMs) * time.Millisecond
}

func (s ServerConfig) IdleTimeout() time.Duration {
	return time.Duration(s.IdleTimeoutMs) * time.Millisecond
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Issuer string `mapstructure:"issuer"`
}

type OTelConfig struct {
	Endpoint    string `mapstructure:"endpoint"`
	ServiceName string `mapstructure:"service_name"`
}

type Route struct {
	Path         string            `mapstructure:"path"`
	Upstream     string            `mapstructure:"upstream"`
	Methods      []string          `mapstructure:"methods"`
	StripPrefix  string            `mapstructure:"strip_prefix"`
	AuthRequired bool              `mapstructure:"auth_required"`
	RateLimit    *RateLimitConfig  `mapstructure:"rate_limit"`
	TimeoutMs    int               `mapstructure:"timeout_ms"`
	Retry        *RetryConfig      `mapstructure:"retry"`
	Headers      map[string]string `mapstructure:"headers"`
}

func (r Route) Timeout() time.Duration {
	return time.Duration(r.TimeoutMs) * time.Millisecond
}

type RateLimitConfig struct {
	RPS   int `mapstructure:"rps"`
	Burst int `mapstructure:"burst"`
}

type RetryConfig struct {
	Attempts  int `mapstructure:"attempts"`
	BackoffMs int `mapstructure:"backoff_ms"`
}

func (r RetryConfig) Backoff() time.Duration {
	return time.Duration(r.BackoffMs) * time.Millisecond
}

type ConfigLoader interface {
	Load(ctx context.Context, path string) (*Config, error)
	Watch(callback func(*Config))
	Get() *Config
	Reload() *Config
}
