package cmd

import (
	"fmt"
	"runtime"

	"github.com/alecthomas/kong"
	"github.com/jyggen/posterr-cli/internal"
)

type VersionFlag bool

func (v VersionFlag) BeforeReset(ctx *kong.Context) error {
	if _, err := fmt.Fprintf(ctx.Stderr, "%s (%s/%s)", internal.Version(), runtime.GOOS, runtime.GOARCH); err != nil {
		return err
	}

	ctx.Exit(0)
	return nil
}
