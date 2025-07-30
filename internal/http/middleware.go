package http

import "net/http"

type Middleware func(*http.Request) (*http.Response, error)

type MiddlewareFunc func(Middleware) Middleware
