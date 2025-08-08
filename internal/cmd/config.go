package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"time"

	"github.com/alecthomas/kong"
	"github.com/jyggen/posterr-cli/internal"
	"github.com/jyggen/posterr-cli/internal/cache"
	"github.com/jyggen/posterr-cli/internal/http"
	"github.com/jyggen/posterr-cli/internal/metadb"
	"github.com/jyggen/posterr-cli/internal/plex"
)

var MaxWorkers = (runtime.NumCPU() * 2) + 1

type CacheConfig struct {
	BasePath string `default:"${cache}" type:"path" help:"Where to store cache data between consecutive runs."`
}

func (c *CacheConfig) AfterApply(ctx *kong.Context) error {
	ctx.FatalIfErrorf(ctx.BindSingletonProvider(func() (*cache.Cache, error) {
		return cache.New(c.BasePath)
	}))
	return nil
}

func (c *CacheConfig) AfterRun(cache *cache.Cache) error {
	return cache.Close()
}

type ConcurrencyConfig struct {
	Workers int `default:"${workers}" help:"Number of workers to use. Must be a number between 2 and ${workers}."`
}

func (c *ConcurrencyConfig) AfterApply() error {
	if c.Workers < 2 || c.Workers > MaxWorkers {
		return fmt.Errorf("workers must be a number between 2 and %d", MaxWorkers)
	}

	return nil
}

type HTTPConfig struct {
	Timeout time.Duration `default:"${timeout}" help:"Maximum duration to wait for any HTTP request made. Must be specified in the string format of Go's time.Duration type."`
}

func (c *HTTPConfig) AfterApply(ctx *kong.Context, ca *cache.Cache) error {
	ctx.FatalIfErrorf(ctx.BindSingletonProvider(func() (*http.Client, error) {
		options := []http.Option{
			http.WithMiddleware(cache.NewCachingMiddleware(ca)),
			http.WithTimeout(c.Timeout),
			http.WithUserAgent(fmt.Sprintf("posterr/%s", internal.Version())),
		}

		return http.NewClient(options...), nil
	}))

	return nil
}

type PlexConfig struct {
	BaseURL url.URL `required:"" help:"Base URL of the Plex Media Server instance."`
	Token   string  `required:"" help:"Token used to authenticate against the Plex Media Server instance."`
}

func (c *PlexConfig) AfterApply(ctx *kong.Context, client *http.Client, cacheSvc *cache.Cache) error {
	if c.BaseURL.Scheme != "https" && c.BaseURL.Scheme != "http" {
		return errors.New("plex base url must be a http(s) url")
	}

	ctx.FatalIfErrorf(ctx.BindSingletonProvider(func() (*plex.Client, error) {
		plexClient, err := plex.NewClient(c.BaseURL.String(), c.Token, client, cacheSvc)
		if err != nil {
			return nil, err
		}

		return plexClient, plexClient.CheckConnectivity()
	}))

	return nil
}

type PostersApiConfig struct {
	ApiURL url.URL `default:"https://posters.metadb.info" help:"Base URL of the API used to fetch a movie's recommended poster."`
}

func (c *PostersApiConfig) AfterApply(kongCtx *kong.Context, client *http.Client) error {
	if c.ApiURL.Scheme != "https" && c.ApiURL.Scheme != "http" {
		return errors.New("api url must be a http(s) url")
	}

	kongCtx.FatalIfErrorf(kongCtx.BindSingletonProvider(func() (*metadb.Client, error) {
		return metadb.NewClient(c.ApiURL.String(), client), nil
	}))

	return nil
}
