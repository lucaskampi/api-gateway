package proxy

import (
	"context"
	"io"
	"net/http"
)

type Request struct {
	Method string
	URL    string
	Header http.Header
	Body   io.Reader
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

type HTTPClient interface {
	Do(ctx context.Context, req *Request) (*Response, error)
}
