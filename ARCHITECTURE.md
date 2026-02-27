# API Gateway Pro - Architecture

A high-performance, production-grade API Gateway built in Go using Fiber framework.

## Why a Custom Gateway?

Off-the-shelf solutions (Nginx, Kong, Traefik) are great but:
- Too heavy for simple use cases
- Complex configuration for specific needs
- Limited customization without plugins

This gateway provides:
- **Lightweight**: ~8-15MB single binary
- **Fast**: Built on fasthttp, >10k RPS
- **Observable**: OpenTelemetry + Prometheus metrics
- **Resilient**: Circuit breaker, retries, timeouts
- **Secure**: JWT auth, rate limiting, CORS

## Architecture Overview

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────────────────────────┐
│                    API Gateway                           │
├─────────────────────────────────────────────────────────┤
│  1. Request ID (X-Request-ID)                          │
│  2. Structured Logging (zerolog)                        │
│  3. Metrics (Prometheus)                                │
│  4. CORS                                               │
│  5. JWT Auth (if required)                              │
│  6. Rate Limiting (per-route token bucket)              │
│  7. OpenTelemetry Tracing                              │
│  8. Recovery (panic handler)                            │
│  9. Timeout (per-route)                                │
│ 10. Retry (with backoff)                               │
│ 11. Circuit Breaker                                    │
│ 12. Proxy Forwarding                                   │
└──────────────┬──────────────────────────────────────────┘
               │
       ┌───────┴───────┐
       ▼               ▼
┌────────────┐  ┌────────────┐
│  Service A │  │  Service B │
└────────────┘  └────────────┘
```

## Middleware Order

The middleware chain order is critical:

1. **Request ID** - Generate/carry X-Request-ID for tracing
2. **Logger** - Structured JSON logging with request info
3. **Metrics** - Prometheus metrics collection
4. **CORS** - Handle cross-origin requests
5. **JWT Auth** - Validate tokens, extract claims (if auth_required)
6. **Rate Limiting** - Per-route token bucket rate limiting
7. **OpenTelemetry** - Trace instrumentation
8. **Recovery** - Catch panics, return 500
9. **Timeout** - Per-request timeout
10. **Retry** - Retry with exponential backoff (if configured)
11. **Circuit Breaker** - Prevent cascading failures
12. **Proxy** - Forward to upstream service

## Project Structure

```
api-gateway/
├── cmd/
│   ├── gateway/          # Main gateway application
│   └── echo/             # Echo service for testing
├── internal/
│   ├── adapter/          # External adapters (implementations)
│   │   ├── auth/         # JWT implementation
│   │   ├── config/      # Viper config loader
│   │   ├── proxy/       # HTTP proxy client
│   │   ├── ratelimit/   # Rate limiting adapter
│   │   └── resilience/  # Circuit breaker
│   ├── domain/           # Business logic & interfaces
│   │   ├── auth/        # Auth port interface
│   │   ├── config/      # Config port interface
│   │   ├── proxy/       # Proxy port interface
│   │   ├── ratelimit/  # Rate limit port interface
│   │   └── resilience/  # Resilience port interface
│   ├── handler/          # HTTP handlers
│   ├── middleware/       # HTTP middleware
│   ├── router/           # Route configuration
│   └── server/           # Server initialization
├── tests/
│   └── integration/      # Integration tests
├── config.yaml           # Configuration file
├── docker-compose.yml    # Docker compose stack
├── Makefile             # Build commands
└── openapi.json         # OpenAPI specification
```

## Configuration

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout_ms: 5000
  write_timeout_ms: 5000

jwt:
  secret: "your-secret-key"
  issuer: "api-gateway"

otel:
  endpoint: "jaeger:4318"
  service_name: "api-gateway"

routes:
  - path: "/api/users/*"
    upstream: "http://users-service:8080"
    methods: ["GET", "POST", "PUT", "DELETE"]
    strip_prefix: "/api/users"
    auth_required: true
    rate_limit:
      rps: 100
      burst: 150
    timeout_ms: 5000
    retry:
      attempts: 3
      backoff_ms: 100
```

## Features

### Authentication
- JWT token validation
- Bearer token in Authorization header
- Claims extraction (user_id, roles, etc.)
- Issuer validation

### Rate Limiting
- Token bucket algorithm
- Per-route configuration
- Per-IP limiting
- Automatic cleanup of stale entries

### Resilience
- **Circuit Breaker**: Prevents cascading failures
- **Retry**: Exponential backoff on failure
- **Timeout**: Per-request timeout
- **Recovery**: Handles panics gracefully

### Observability
- **Traces**: OpenTelemetry + Jaeger
- **Metrics**: Prometheus at /metrics
- **Logs**: Structured JSON logging with zerolog
- **Request ID**: Correlation across services

### API Documentation
- Swagger UI at /docs
- OpenAPI spec at /openapi.json

## Building & Running

```bash
# Build
make build

# Run tests
make test

# Run with Docker
docker-compose up --build

# Run locally
go run cmd/gateway/main.go -config config.yaml
```

## Performance

- **Binary Size**: ~10MB
- **Throughput**: >10k RPS on single core
- **p99 Latency**: <10ms (proxy only)
- **Memory**: Low footprint, connection pooling

## Extending

### Adding a New Middleware

1. Create `internal/middleware/yourfeature.go`
2. Implement `fiber.Handler` interface
3. Add to router in `internal/router/router.go`

### Adding a New Route

Add to `config.yaml`:

```yaml
routes:
  - path: "/api/your-service/*"
    upstream: "http://your-service:8080"
    methods: ["GET"]
    auth_required: false
    rate_limit:
      rps: 100
      burst: 150
    timeout_ms: 3000
```

## License

MIT
