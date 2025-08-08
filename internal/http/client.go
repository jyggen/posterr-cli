package http

import (
	"context"
	"net/http"
)

var _ http.RoundTripper = (*Client)(nil)

type Client struct {
	client      *http.Client
	middlewares []MiddlewareFunc
}

func NewClient(options ...Option) *Client {
	client := &Client{}
	client.client = &http.Client{
		Transport: client,
	}

	return client.WithOptions(options...)
}

func (c *Client) clone() *Client {
	clone := *c
	clone.client = c.Client()
	clone.client.Transport = &clone

	return &clone
}

func (c *Client) WithOptions(options ...Option) *Client {
	clone := c.clone()

	for _, o := range options {
		o.apply(clone)
	}

	return clone
}

func (c *Client) Client() *http.Client {
	clone := *c.client

	return &clone
}

func (c *Client) RoundTrip(req *http.Request) (*http.Response, error) {
	next := http.DefaultTransport.RoundTrip
	for _, m := range c.middlewares {
		next = m(next)
	}

	return next(req)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}
