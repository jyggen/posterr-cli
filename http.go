package main

import (
	"net/http"
	"time"
)

type customTransport struct {
	transport http.RoundTripper
}

func (ct *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "posterr/"+version)

	return ct.transport.RoundTrip(req)
}

func newClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: &customTransport{http.DefaultTransport},
	}
}
