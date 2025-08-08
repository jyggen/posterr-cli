package cmd

import (
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

var MaxThreads = (runtime.NumCPU() * 2) + 1

type CacheConfig struct {
	CacheBasePath string `default:"${cache}" type:"path" help:""`
}

func (c *CacheConfig) AfterApply(ctx *kong.Context) error {
	ctx.FatalIfErrorf(ctx.BindSingletonProvider(func() (*cache.Cache, error) {
		return cache.New(c.CacheBasePath)
	}))
	return nil
}

func (c *CacheConfig) AfterRun(cache *cache.Cache) error {
	return cache.Close()
}

type ConcurrencyConfig struct {
	Threads int `default:"${threads}" help:""`
}

func (c *ConcurrencyConfig) Validate() error {
	if c.Threads < 2 || c.Threads > MaxThreads {
		return fmt.Errorf("threads must be a number between 2 and %d", MaxThreads)
	}

	return nil
}

type HTTPConfig struct {
	HTTPTimeout time.Duration `help:"" default:"${timeout}"`
}

func (c *HTTPConfig) AfterApply(ctx *kong.Context, ca *cache.Cache) error {
	ctx.FatalIfErrorf(ctx.BindSingletonProvider(func() (*http.Client, error) {
		options := []http.Option{
			http.WithMiddleware(cache.NewCachingMiddleware(ca)),
			http.WithTimeout(c.HTTPTimeout),
			http.WithUserAgent(fmt.Sprintf("posterr/%s", internal.Version())),
		}

		return http.NewClient(options...), nil
	}))

	return nil
}

type MetaDBConfig struct {
	ApiURL url.URL `default:"https://posters.metadb.info" help:""`
}

func (c *MetaDBConfig) AfterApply(kongCtx *kong.Context, client *http.Client) error {
	kongCtx.FatalIfErrorf(kongCtx.BindSingletonProvider(func() (*metadb.Client, error) {
		return metadb.NewClient(c.ApiURL.String(), client), nil
	}))

	return nil
}

type PlexConfig struct {
	PlexBaseURL url.URL `required:"" name:"plex-base-url" help:""`
	PlexToken   string  `required:"" name:"plex-token" help:""`
}

func (c *PlexConfig) AfterApply(ctx *kong.Context, client *http.Client, cacheSvc *cache.Cache) error {
	ctx.FatalIfErrorf(ctx.BindSingletonProvider(func() (*plex.Client, error) {
		plexClient, err := plex.NewClient(c.PlexBaseURL.String(), c.PlexToken, client, cacheSvc)
		if err != nil {
			return nil, err
		}

		return plexClient, plexClient.CheckConnectivity()
	}))

	return nil
}
