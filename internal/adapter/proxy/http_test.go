package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"api-gateway/internal/middleware"

	"github.com/gofiber/fiber/v3"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient(Options{
		DialTimeout:     5,
		ReadTimeout:     10,
		WriteTimeout:    10,
		IdleConnTimeout: 30,
		MaxIdleConns:    100,
	})

	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
}

func TestHTTPClient_Do(t *testing.T) {
	t.Skip("Skipping - requires running HTTP server")
}

func TestNewRequest(t *testing.T) {
	req := NewRequest("GET", "http://example.com/path", []byte(`{"test": true}`))

	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "http://example.com/path", req.URL)
	assert.NotNil(t, req.Header)
	assert.NotNil(t, req.Body)
}

func TestParseURL(t *testing.T) {
	tests := []struct {
		name      string
		upstream  string
		path      string
		query     string
		wantPath  string
		wantQuery string
		wantErr   bool
	}{
		{
			name:      "basic URL",
			upstream:  "http://example.com",
			path:      "/api/users",
			query:     "",
			wantPath:  "/api/users",
			wantQuery: "",
			wantErr:   false,
		},
		{
			name:      "with query string",
			upstream:  "http://example.com",
			path:      "/api/users",
			query:     "page=1",
			wantPath:  "/api/users",
			wantQuery: "page=1",
			wantErr:   false,
		},
		{
			name:     "invalid upstream",
			upstream: "://invalid",
			path:     "/api",
			query:    "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := parseURL(tt.upstream, tt.path, tt.query)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantPath, u.Path)
			assert.Equal(t, tt.wantQuery, u.RawQuery)
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string", "test", "test"},
		{"nil", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHTTPClient_Close(t *testing.T) {
	client := NewHTTPClient(Options{
		DialTimeout:     5,
		ReadTimeout:     10,
		WriteTimeout:    10,
		IdleConnTimeout: 30,
		MaxIdleConns:    100,
	})

	err := client.Close()
	assert.NoError(t, err)
}

func TestForward_RetryThenSuccess(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := calls.Add(1)
		if current == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("temporary"))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	app := fiber.New()
	client := NewHTTPClient(Options{
		DialTimeout:         1 * time.Second,
		ReadTimeout:         1 * time.Second,
		WriteTimeout:        1 * time.Second,
		IdleConnTimeout:     10 * time.Second,
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
	})

	app.Use(middleware.Retry(middleware.RetryConfig{
		Attempts:   1,
		Backoff:    1 * time.Millisecond,
		MaxBackoff: 10 * time.Millisecond,
	}))
	app.Get("/proxy", client.Forward(struct {
		Upstream    string
		StripPrefix string
		Headers     map[string]string
	}{
		Upstream:    upstream.URL,
		StripPrefix: "",
	}))

	req := httptest.NewRequest(http.MethodGet, "/proxy", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.GreaterOrEqual(t, calls.Load(), int32(2))
}

func TestForward_Timeout(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(40 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("slow"))
	}))
	defer upstream.Close()

	app := fiber.New()
	client := NewHTTPClient(Options{
		DialTimeout:         1 * time.Second,
		ReadTimeout:         1 * time.Second,
		WriteTimeout:        1 * time.Second,
		IdleConnTimeout:     10 * time.Second,
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
	})

	app.Use(middleware.Timeout(10 * time.Millisecond))
	app.Get("/proxy", client.Forward(struct {
		Upstream    string
		StripPrefix string
		Headers     map[string]string
	}{
		Upstream:    upstream.URL,
		StripPrefix: "",
	}))

	req := httptest.NewRequest(http.MethodGet, "/proxy", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

var _ = io.Discard
