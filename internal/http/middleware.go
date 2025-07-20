package http

import "net/http"

var _ http.RoundTripper = (Middleware)(nil)

type Middleware func(*http.Request) (*http.Response, error)

func (m Middleware) RoundTrip(req *http.Request) (*http.Response, error) {
	return m(req)
}

type MiddlewareFunc func(Middleware) Middleware

func (m MiddlewareFunc) apply(c *Client) {
	c.client.Transport = func(t http.RoundTripper) http.RoundTripper {
		return m(t.RoundTrip)
	}(c.client.Transport)
}
