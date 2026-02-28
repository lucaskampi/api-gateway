package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, values := range req.Header {
		for _, value := range values {
			httpReq.Header.Add(key, value)
		}
	}

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
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

		body := append([]byte(nil), ctx.Body()...)
		baseHeaders := make(http.Header)
		ctx.Request().Header.VisitAll(func(key, value []byte) {
			baseHeaders.Add(string(key), string(value))
		})

		if route.Headers != nil {
			userID := getUserID(ctx)
			userClaims := getUserClaims(ctx)
			for k, v := range route.Headers {
				v = strings.ReplaceAll(v, "{{.UserID}}", userID)
				for claimKey, claimValue := range userClaims {
					v = strings.ReplaceAll(v, "{{."+claimKey+"}}", toString(claimValue))
				}
				baseHeaders.Set(k, v)
			}
		}

		attempts := 1
		if v, ok := ctx.Locals("retry_attempts").(int); ok && v > 0 {
			attempts = v + 1
		}

		backoff, _ := ctx.Locals("retry_backoff").(time.Duration)
		maxBackoff, ok := ctx.Locals("retry_max_backoff").(time.Duration)
		if !ok || maxBackoff <= 0 {
			maxBackoff = 5 * time.Second
		}

		timeout, _ := ctx.Locals("request_timeout").(time.Duration)

		var lastErr error
		for attempt := 0; attempt < attempts; attempt++ {
			if attempt > 0 && backoff > 0 {
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}

			reqCtx := ctx.Context()
			cancel := func() {}
			if timeout > 0 {
				reqCtxWithTimeout, cancelFunc := context.WithTimeout(reqCtx, timeout)
				reqCtx = reqCtxWithTimeout
				cancel = cancelFunc
			}

			req, err := http.NewRequestWithContext(reqCtx, ctx.Method(), target.String(), bytes.NewReader(body))
			if err != nil {
				cancel()
				return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{
					"error": "failed to create request",
				})
			}
			req.Header = cloneHeaders(baseHeaders)

			resp, err := c.client.Do(req)
			cancel()
			if err != nil {
				lastErr = err
				continue
			}

			bodyBytes, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				lastErr = readErr
				continue
			}

			if resp.StatusCode >= fiber.StatusInternalServerError && attempt < attempts-1 {
				continue
			}

			return writeResponse(ctx, resp.StatusCode, resp.Header, bodyBytes)
		}

		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":   "failed to forward request",
			"details": errorMessage(lastErr),
		})
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

	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func cloneHeaders(source http.Header) http.Header {
	target := make(http.Header, len(source))
	for key, values := range source {
		for _, value := range values {
			target.Add(key, value)
		}
	}
	return target
}

func writeResponse(ctx fiber.Ctx, status int, headers http.Header, body []byte) error {
	ctx.Response().Reset()
	ctx.Response().SetStatusCode(status)
	for key, values := range headers {
		if len(values) > 0 {
			ctx.Response().Header.Set(key, values[0])
		}
	}

	_, err := ctx.Response().BodyWriter().Write(body)
	return err
}

func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
