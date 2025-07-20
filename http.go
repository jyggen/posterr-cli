package main

import (
	"fmt"
	"github.com/jyggen/posterr-cli/internal/http"
	"time"
)

func newClient(timeout time.Duration) *http.Client {
	return http.NewClient([]http.Option{
		http.WithTimeout(timeout),
		http.WithUserAgent(fmt.Sprintf("posterr/%s", version)),
	}...)
}
