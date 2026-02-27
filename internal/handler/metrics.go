package handler

import (
	"bufio"
	"errors"
	"net"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Metrics() fiber.Handler {
	return func(c fiber.Ctx) error {
		handler := promhttp.Handler()

		uri := c.OriginalURL()
		req, err := http.NewRequest(c.Method(), uri, nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		req.Header = make(http.Header)
		c.Request().Header.VisitAll(func(key, value []byte) {
			req.Header.Add(string(key), string(value))
		})
		req.Host = c.Hostname()
		req.RemoteAddr = c.IP()

		rec := &recorder{statusCode: 200, body: &[]byte{}}
		handler.ServeHTTP(rec, req)

		c.Response().SetStatusCode(rec.statusCode)
		for k, v := range rec.Header() {
			c.Response().Header.Set(k, v[0])
		}
		c.Response().SetBody(*rec.body)

		return nil
	}
}

type recorder struct {
	statusCode int
	header     http.Header
	body       *[]byte
}

func (r *recorder) Header() http.Header {
	if r.header == nil {
		r.header = make(http.Header)
	}
	return r.header
}

func (r *recorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func (r *recorder) Write(b []byte) (int, error) {
	*r.body = append(*r.body, b...)
	return len(b), nil
}

func (r *recorder) BytesWritten() int {
	return len(*r.body)
}

func (r *recorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("hijack not supported")
}

var _ http.Hijacker = (*recorder)(nil)
