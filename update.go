package main

import (
	"context"
	"github.com/chelnak/ysmrr"
	"github.com/jyggen/go-plex-client"
	"os"
	"strings"
)

type updateCmd struct {
	httpConfig
	plexConfig
	Force bool `help:""`
}

func update(cli *posterrCli) error {
	client := newClient(cli.Update.HttpTimeout)
	connection, err := plex.New(strings.TrimSuffix(cli.Update.PlexBaseUrl.String(), "/"), cli.Update.PlexToken)

	if err != nil {
		return err
	}

	connection.HTTPClient = *client

	return withThreads(func(ctx context.Context, queue chan plex.Metadata) error {
		return produceMoviesMetadata(ctx, connection, queue)
	}, func(ctx context.Context, queue chan plex.Metadata, s *ysmrr.Spinner) error {
		for job := range queue {
			imdbId := getImdbId(job)

			if imdbId == "" {
				continue
			}

			metadbPath, err := getPosterByImdbId(ctx, client, cli.CacheBasePath, imdbId, s)

			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err != nil {
				return err
			}

			s.UpdateMessagef("%s: Downloading current poster from Plex...", imdbId)
			plexPath, err := getPosterByMetadata(connection, cli.CacheBasePath, job)

			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err != nil {
				return err
			}

			s.UpdateMessagef("%s: Comparing poster checksums...", imdbId)
			metadbHash, err := hashFile(metadbPath)

			if err != nil {
				return err
			}

			plexHash, err := hashFile(plexPath)

			if err != nil {
				return err
			}

			if plexHash == metadbHash {
				continue
			}

			s.UpdateMessagef("%s: Uploading poster to Plex...", imdbId)
			f, err := os.Open(metadbPath)

			if err != nil {
				return err
			}

			if err = connection.UploadPoster(job.RatingKey, f); err != nil {
				return err
			}
		}

		return nil
	}, cli.Threads)
}
