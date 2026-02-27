package handler

import (
	"bytes"
	_ "embed"
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

var openAPISpec map[string]interface{}

//go:embed openapi.json
var openAPISpecData []byte

func init() {
	decoder := json.NewDecoder(bytes.NewReader(openAPISpecData))
	decoder.UseNumber()
	if err := decoder.Decode(&openAPISpec); err != nil {
		openAPISpec = nil
	}
}

func OpenAPI() fiber.Handler {
	return func(c fiber.Ctx) error {
		if openAPISpec == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "OpenAPI spec not loaded",
			})
		}
		return c.JSON(openAPISpec)
	}
}

func SwaggerUI() fiber.Handler {
	return func(c fiber.Ctx) error {
		html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Gateway Pro - Swagger UI</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.11.0/swagger-ui.css">
    <style>
        body { margin: 0; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: '/openapi.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	}
}
