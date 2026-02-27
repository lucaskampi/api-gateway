package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	app := fiber.New()
	app.Get("/health", Health())

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestReady(t *testing.T) {
	app := fiber.New()
	app.Get("/ready", Ready())

	req := httptest.NewRequest("GET", "/ready", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestOpenAPI(t *testing.T) {
	t.Skip("Skipping - requires openapi.json in working directory")
}

func TestSwaggerUI(t *testing.T) {
	app := fiber.New()
	app.Get("/docs", SwaggerUI())

	req := httptest.NewRequest("GET", "/docs", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")
}
