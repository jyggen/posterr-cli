package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/browser"
)

type previewCmd struct {
	httpConfig
	metadbConfig
	ImdbID string `arg:"" name:"imdb-id" help:""`
}

func (p *previewCmd) Run(cli *posterrCli) error {
	client := newClient(cli.Preview.HTTPTimeout)
	metadbClient, err := newMetaDBClient(cli.Preview.ApiURL, cli.Preview.DnsResolver, cli.Preview.HTTPTimeout)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()
	posterUrl, err := getPosterByImdbId(ctx, metadbClient, cli.Preview.ImdbID, nil)

	if err != nil {
		return err
	}

	if posterUrl == "" {
		return fmt.Errorf("%s: unknown movie", cli.Preview.ImdbID)
	}

	posterPath, err := downloadOrCache(client.Get, cli.CacheBasePath, posterUrl)

	if err != nil {
		return fmt.Errorf("%s: %w", cli.Preview.ImdbID, err)
	}

	return browser.OpenFile(posterPath)
}
