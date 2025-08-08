package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/jyggen/posterr-cli/internal/cmd"
	"github.com/jyggen/posterr-cli/internal/cmd/compare"
	"github.com/jyggen/posterr-cli/internal/cmd/preview"
	"github.com/jyggen/posterr-cli/internal/cmd/update"
)

const (
	applicationName = "posterr"
	defaultTimeout  = 10 * time.Second
)

type cli struct {
	Compare *compare.Command `cmd:"" help:"Compare your current Plex posters against the best posters available and generate an HTML file with all posters that do not match'."`
	Preview *preview.Command `cmd:"" help:"Open the best poster available for the movie specified in a new browser window."`
	Update  *update.Command  `cmd:"" help:"Update any Plex poster that does not match the best poster available."`
	Version cmd.VersionFlag  `help:""`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		cancel()
	}()

	cacheDir, _ := os.UserCacheDir()
	command := cli{}
	kongCtx := kong.Parse(&command, kong.Name(applicationName), kong.UsageOnError(), kong.Vars{
		"cache":   filepath.Join(cacheDir, applicationName),
		"timeout": defaultTimeout.String(),
		"workers": strconv.Itoa(cmd.MaxWorkers),
	}, kong.BindTo(ctx, (*context.Context)(nil)))

	kongCtx.FatalIfErrorf(kongCtx.Run())
}
