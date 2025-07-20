package main

import (
	"github.com/jyggen/posterr-cli/cmd"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/alecthomas/kong"
)

const applicationName = "posterr"
const defaultTimeout = 10 * time.Second

var maxThreads = (runtime.NumCPU() * 2) + 1

type cli struct {
	Compare *compareCmd `cmd:"" help:""`
	Preview *previewCmd `cmd:"" help:""`

	//Update  *update.Command `cmd:"" help:""`
	Version VersionFlag `name:"version" help:"Show version number."`
}

func main() {
	cacheDir, _ := os.UserCacheDir()
	ctx := kong.Parse(&cli{}, kong.Name(applicationName), kong.UsageOnError(), kong.Vars{
		"cache":   filepath.Join(cacheDir, applicationName),
		"threads": strconv.Itoa(cmd.MaxThreads),
		"timeout": defaultTimeout.String(),
	})

	ctx.FatalIfErrorf(ctx.Run())
}
