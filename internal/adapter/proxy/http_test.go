package proxy

import (
	"io"
	"testing"

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

var _ = io.Discard
