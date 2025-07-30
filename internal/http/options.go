package http

import (
	"net/http"
	"time"
)

type Option interface {
	apply(*Client)
}

type optionFunc func(*Client)

func (f optionFunc) apply(client *Client) {
	f(client)
}

func WithMiddleware(middleware MiddlewareFunc, options ...bool) Option {
	if len(options) > 0 && options[0] == true {
		return optionFunc(func(c *Client) {
			c.middlewares = append([]MiddlewareFunc{middleware}, c.middlewares...)
		})
	}

	return optionFunc(func(c *Client) {
		c.middlewares = append(c.middlewares, middleware)
	})
}

func WithSkipRedirects() Option {
	return optionFunc(func(c *Client) {
		c.client.CheckRedirect = func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}
	})
}

func WithTimeout(timeout time.Duration) Option {
	return optionFunc(func(c *Client) {
		c.client.Timeout = timeout
	})
}

func WithUserAgent(userAgent string) Option {
	return WithMiddleware(func(next Middleware) Middleware {
		return func(req *http.Request) (*http.Response, error) {
			req.Header.Set("User-Agent", userAgent)
			return next(req)
		}
	})
}
