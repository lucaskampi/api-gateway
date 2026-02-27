package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var (
	tracer     *sdktrace.TracerProvider
	propagator propagation.TextMapPropagator
	tracerName = "api-gateway"
)

func InitOTel(endpoint, serviceName string) error {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return err
	}

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return err
	}

	tracer = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracer)

	propagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(propagator)

	return nil
}

func OTel() fiber.Handler {
	tr := otel.Tracer(tracerName)
	return func(c fiber.Ctx) error {
		startTime := time.Now()
		ctx := c.Context()

		ctx = propagator.Extract(ctx, fiberCarrier{c})

		spanName := c.Method() + " " + c.Path()

		ctx, span := tr.Start(ctx, spanName)
		span.SetAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.url", c.OriginalURL()),
			attribute.String("http.route", c.Route().Path),
			attribute.String("http.host", c.Hostname()),
			attribute.String("http.scheme", c.Protocol()),
			attribute.String("net.peer.ip", c.IP()),
			attribute.String("user_agent.original", string(c.Request().Header.UserAgent())),
		)

		requestID := c.Get("X-Request-ID")
		if requestID != "" {
			span.SetAttributes(attribute.String("http.request_id", requestID))
		}

		c.Locals("ctx", ctx)

		err := c.Next()

		duration := time.Since(startTime)

		span.SetAttributes(
			attribute.Int("http.status_code", c.Response().StatusCode()),
			attribute.Int64("http.response_time_ms", duration.Milliseconds()),
		)

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("error", err.Error()))
		}

		if c.Response().StatusCode() >= 500 {
			span.SetAttributes(attribute.Bool("error", true))
		}

		span.End()

		return err
	}
}

type fiberCarrier struct {
	fiber.Ctx
}

func (f fiberCarrier) Get(key string) string {
	return f.Ctx.Get(key)
}

func (f fiberCarrier) Keys() []string {
	return nil
}

func ShutdownOTel() error {
	if tracer != nil {
		return tracer.Shutdown(context.Background())
	}
	return nil
}
