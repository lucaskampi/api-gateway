package router

import (
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
	handlers = append(handlers, middleware.CORS(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
	}))

	if route.AuthRequired {
		handlers = append(handlers, middleware.JWT(middleware.JWTConfig{
			Secret: r.cfg.JWT.Secret,
			Issuer: r.cfg.JWT.Issuer,
		}))
	}

	if route.RateLimit != nil {
		handlers = append(handlers, middleware.RateLimit(route.RateLimit.RPS, route.RateLimit.Burst))
	}

	handlers = append(handlers, middleware.OTel())
	handlers = append(handlers, middleware.Recovery(r.logger))
	handlers = append(handlers, middleware.Timeout(route.Timeout()))

	if route.Retry != nil && route.Retry.Attempts > 0 {
		handlers = append(handlers, middleware.CircuitBreakerMiddleware(route.Retry.Attempts, route.Retry.Backoff()))
	}

	routeCfg := routeConfig{
		Upstream:    route.Upstream,
		StripPrefix: route.StripPrefix,
		Headers:     route.Headers,
	}
	handlers = append(handlers, r.proxy.Forward(routeCfg))

	return handlers
}
