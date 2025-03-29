package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/browser"
)

type previewCmd struct {
	httpConfig
	ImdbID string `arg:"" name:"imdb-id" help:""`
}

func (p *previewCmd) Run(cli *posterrCli) error {
	client := newClient(cli.Preview.HTTPTimeout)
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		cancel()
	}()
	posterPath, err := getPosterByImdbId(ctx, client, cli.CacheBasePath, cli.Preview.ImdbID, nil)

	if err != nil {
		return err
	}

	if posterPath == "" {
		return errors.New("unknown movie")
	}

	return browser.OpenFile(posterPath)
}
