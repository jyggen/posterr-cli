package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/alecthomas/kong"
)

type httpConfig struct {
	HTTPTimeout time.Duration `help:"" default:"${timeout}"`
}

type plexConfig struct {
	PlexBaseURL url.URL `arg:"" name:"plex-base-url" help:""`
	PlexToken   string  `arg:"" name:"plex-token" help:""`
}

type posterrCli struct {
	CacheBasePath string      `help:"" default:"${cache}"`
	Compare       *compareCmd `cmd:"" help:""`
	Preview       *previewCmd `cmd:"" help:""`
	Threads       int         `help:"" default:"${threads}"`
	Update        *updateCmd  `cmd:"" help:""`
	Version       VersionFlag `name:"version" help:"Show version number."`
}

const defaultTimeout = 10 * time.Second

var maxThreads int

func init() {
	maxThreads = (runtime.NumCPU() * 2) + 1
}

func (p *posterrCli) Validate() error {
	if p.Threads < 1 || p.Threads > maxThreads {
		return fmt.Errorf("threads must be a number between 1 and %d", maxThreads)
	}

	return nil
}

func main() {
	cacheDir, err := os.UserCacheDir()

	if err != nil {
		panic(err)
	}

	cli := &posterrCli{}
	ctx := kong.Parse(cli, kong.Name("posterr"), kong.UsageOnError(), kong.Vars{
		"cache":   filepath.Join(cacheDir, "posterr"),
		"threads": strconv.Itoa(maxThreads),
		"timeout": defaultTimeout.String(),
	})

	ctx.FatalIfErrorf(ctx.Run())
}
