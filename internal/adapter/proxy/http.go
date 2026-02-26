package proxy

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"api-gateway/internal/domain/proxy"

	"github.com/gofiber/fiber/v3"
)

type HTTPClient struct {
	client *http.Client
}

type Options struct {
	DialTimeout         time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	IdleConnTimeout     time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
}

func NewHTTPClient(opts Options) *HTTPClient {
	tr := &http.Transport{
		MaxIdleConns:        opts.MaxIdleConns,
		MaxIdleConnsPerHost: opts.MaxIdleConnsPerHost,
		IdleConnTimeout:     opts.IdleConnTimeout,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   opts.DialTimeout + opts.ReadTimeout + opts.WriteTimeout,
	}

	return &HTTPClient{
		client: client,
	}
}

func (c *HTTPClient) Do(ctx context.Context, req *proxy.Request) (*proxy.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, req.Body)
	if err != nil {
		return nil, err
	}

	for key, values := range req.Header {
		for _, value := range values {
			httpReq.Header.Add(key, value)
		}
	}

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	resp := &proxy.Response{
		StatusCode: httpResp.StatusCode,
		Header:     httpResp.Header,
		Body:       body,
	}

	return resp, nil
}

func (c *HTTPClient) Close() error {
	c.client.CloseIdleConnections()
	return nil
}

func NewRequest(method, url string, body []byte) *proxy.Request {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	return &proxy.Request{
		Method: method,
		URL:    url,
		Header: make(http.Header),
		Body:   bodyReader,
	}
}

func (c *HTTPClient) Forward(route struct {
	Upstream    string
	StripPrefix string
	Headers     map[string]string
}) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		upstream := route.Upstream
		path := ctx.Path()

		if route.StripPrefix != "" {
			path = strings.TrimPrefix(path, route.StripPrefix)
		}

		target, err := parseURL(upstream, path, string(ctx.Request().URI().QueryString()))
		if err != nil {
			return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": "invalid upstream URL",
			})
		}

		body := ctx.Body()
		req, err := http.NewRequestWithContext(ctx.Context(), ctx.Method(), target.String(), bytes.NewReader(body))
		if err != nil {
			return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": "failed to create request",
			})
		}

		req.Header = make(http.Header)
		ctx.Request().Header.VisitAll(func(key, value []byte) {
			req.Header.Add(string(key), string(value))
		})

		if route.Headers != nil {
			userID := getUserID(ctx)
			userClaims := getUserClaims(ctx)
			for k, v := range route.Headers {
				v = strings.ReplaceAll(v, "{{.UserID}}", userID)
				if userClaims != nil {
					for claimKey, claimValue := range userClaims {
						v = strings.ReplaceAll(v, "{{."+claimKey+"}}", toString(claimValue))
					}
				}
				req.Header.Set(k, v)
			}
		}

		resp, err := c.client.Do(req)
		if err != nil {
			return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": "failed to forward request",
			})
		}
		defer resp.Body.Close()

		ctx.Response().Reset()
		ctx.Response().SetStatusCode(resp.StatusCode)
		for k, v := range resp.Header {
			if len(v) > 0 {
				ctx.Response().Header.Set(k, v[0])
			}
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": "failed to read response body",
			})
		}

		_, err = ctx.Response().BodyWriter().Write(bodyBytes)
		return err
	}
}

func parseURL(upstream, path, query string) (*url.URL, error) {
	u, err := url.Parse(upstream)
	if err != nil {
		return nil, err
	}
	u.Path = path
	u.RawQuery = query
	return u, nil
}

func getUserID(ctx fiber.Ctx) string {
	if id, ok := ctx.Locals("user_id").(string); ok {
		return id
	}
	return ""
}

func getUserClaims(ctx fiber.Ctx) map[string]interface{} {
	if claims, ok := ctx.Locals("user_claims").(map[string]interface{}); ok {
		return claims
	}
	return nil
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return strings.Trim(strings.ReplaceAll(toStringRecursive(v), `"`, ""), `"`)
}

func toStringRecursive(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return string(rune(int(val)))
	case int:
		return string(rune(val))
	default:
		return ""
	}
}
