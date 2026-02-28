package router

import (
	"time"

	"api-gateway/internal/adapter/proxy"
	"api-gateway/internal/domain/config"
	"api-gateway/internal/handler"
	"api-gateway/internal/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type Router struct {
	app    *fiber.App
	cfg    *config.Config
	logger zerolog.Logger
	proxy  *proxy.HTTPClient
}

func New(app *fiber.App, cfg *config.Config, logger zerolog.Logger) *Router {
	httpClient := proxy.NewHTTPClient(proxy.Options{
		DialTimeout:         5,
		ReadTimeout:         10,
		WriteTimeout:        10,
		IdleConnTimeout:     30,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	})

	return &Router{
		app:    app,
		cfg:    cfg,
		logger: logger,
		proxy:  httpClient,
	}
}

func (r *Router) Setup() {
	r.app.Get("/health", handler.Health())
	r.app.Get("/ready", handler.Ready())
	r.app.Get("/metrics", handler.Metrics())
	r.app.Get("/docs", handler.SwaggerUI())
	r.app.Get("/openapi.json", handler.OpenAPI())

	r.setupRoutes()
}

func (r *Router) setupRoutes() {
	for _, route := range r.cfg.Routes {
		methods := route.Methods
		if len(methods) == 0 {
			methods = []string{"GET"}
		}

		handlers := r.buildMiddlewareList(&route)

		for _, method := range methods {
			switch method {
			case "GET":
				r.app.Get(route.Path, handlers[0], handlers[1:]...)
			case "POST":
				r.app.Post(route.Path, handlers[0], handlers[1:]...)
			case "PUT":
				r.app.Put(route.Path, handlers[0], handlers[1:]...)
			case "DELETE":
				r.app.Delete(route.Path, handlers[0], handlers[1:]...)
			case "PATCH":
				r.app.Patch(route.Path, handlers[0], handlers[1:]...)
			}
		}
	}
}

type routeConfig struct {
	Upstream    string
	StripPrefix string
	Headers     map[string]string
}

func (r *Router) buildMiddlewareList(route *config.Route) []fiber.Handler {
	var handlers []fiber.Handler

	handlers = append(handlers, func(c fiber.Ctx) error {
		c.Locals("upstream", route.Upstream)
		return c.Next()
	})

	handlers = append(handlers, middleware.RequestID())
	handlers = append(handlers, middleware.Logger(r.logger))
	handlers = append(handlers, middleware.Metrics())
	handlers = append(handlers, middleware.CORS(middleware.CORSConfig{
		AllowOrigins:     r.cfg.CORS.AllowOrigins,
		AllowMethods:     r.cfg.CORS.AllowMethods,
		AllowHeaders:     r.cfg.CORS.AllowHeaders,
		AllowCredentials: r.cfg.CORS.AllowCredentials,
		ExposeHeaders:    r.cfg.CORS.ExposeHeaders,
		MaxAge:           r.cfg.CORS.MaxAge,
	}))

	if route.AuthRequired {
		handlers = append(handlers, middleware.JWT(middleware.JWTConfig{
			Secret: r.cfg.JWT.Secret,
			Issuer: r.cfg.JWT.Issuer,
		}))
	}

	if r.cfg.GlobalRateLimit != nil {
		handlers = append(handlers, middleware.RateLimitWithConfig(middleware.RateLimitConfig{
			GlobalRPS:   r.cfg.GlobalRateLimit.RPS,
			GlobalBurst: r.cfg.GlobalRateLimit.Burst,
			GlobalKeyBy: r.cfg.GlobalRateLimit.KeyBy,
		}))
	}

	if route.RateLimit != nil {
		handlers = append(handlers, middleware.RateLimitWithConfig(middleware.RateLimitConfig{
			RouteID:    route.Path,
			RouteRPS:   route.RateLimit.RPS,
			RouteBurst: route.RateLimit.Burst,
			RouteKeyBy: route.RateLimit.KeyBy,
		}))
	}

	handlers = append(handlers, middleware.OTel())
	handlers = append(handlers, middleware.Timeout(route.Timeout()))
	handlers = append(handlers, middleware.Recovery(r.logger))

	if route.Retry != nil && route.Retry.Attempts > 0 {
		handlers = append(handlers, middleware.CircuitBreakerMiddleware(route.Retry.Attempts, route.Retry.Backoff()))
		handlers = append(handlers, middleware.Retry(middleware.RetryConfig{
			Attempts:   route.Retry.Attempts,
			Backoff:    route.Retry.Backoff(),
			MaxBackoff: 5 * time.Second,
		}))
	}

	routeCfg := routeConfig{
		Upstream:    route.Upstream,
		StripPrefix: route.StripPrefix,
		Headers:     route.Headers,
	}
	handlers = append(handlers, r.proxy.Forward(routeCfg))

	return handlers
}
