package main

import (
	"fmt"
	"github.com/adrg/xdg"
	"github.com/alecthomas/kong"
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

type httpConfig struct {
	HttpTimeout time.Duration `help:"" default:"${timeout}"`
}

type plexConfig struct {
	PlexBaseUrl url.URL `arg:"" name:"plex-base-url" help:""`
	PlexToken   string  `arg:"" name:"plex-token" help:""`
}

type posterrCli struct {
	CacheBasePath string      `help:"" default:"${cache}"`
	Compare       *compareCmd `cmd:"" help:""`
	Threads       int         `help:"" default:"${threads}"`
	Update        *updateCmd  `cmd:"" help:""`
	Version       VersionFlag `name:"version" help:"Show version number."`
}

func (p *posterrCli) Validate() error {
	maxThreads := runtime.NumCPU()

	if p.Threads < 1 || p.Threads > maxThreads {
		return fmt.Errorf("threads must be a number between 1 and %d", runtime.NumCPU())
	}

	return nil
}

func main() {
	cli := &posterrCli{}
	ctx := kong.Parse(cli, kong.Name("posterr"), kong.UsageOnError(), kong.Vars{
		"cache":   filepath.Join(xdg.CacheHome, "posterr"),
		"threads": strconv.Itoa(runtime.NumCPU()),
		"timeout": (time.Second * 10).String(),
	})

	switch ctx.Command() {
	case "compare <plex-base-url> <plex-token>":
		ctx.FatalIfErrorf(compare(cli))
	case "update <plex-base-url> <plex-token>":
		ctx.FatalIfErrorf(update(cli))
	default:
		ctx.FatalIfErrorf(ctx.PrintUsage(true))
	}
}
