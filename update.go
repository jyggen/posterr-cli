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

	if _, err = connection.Test(); err != nil {
		return err
	}

	withThreads(func(ctx context.Context, queue chan plex.Metadata) {
		produceMoviesMetadata(ctx, connection, queue)
	}, func(ctx context.Context, queue chan plex.Metadata, s *ysmrr.Spinner) {
		for job := range queue {
			imdbId := getImdbId(job)

			if imdbId == "" {
				continue
			}

			metadbPath, err := getPosterByImdbId(ctx, client, cli.CacheBasePath, imdbId, s)

			if ctx.Err() != nil {
				s.ErrorWithMessage("Cancelled.")
				return
			}

			if err != nil {
				s.UpdateMessagef("Errored.")
				continue
			}

			s.UpdateMessagef("%s: Downloading current poster from Plex...", imdbId)
			plexPath, err := getPosterByMetadata(connection, cli.CacheBasePath, job)

			if ctx.Err() != nil {
				s.ErrorWithMessage("Cancelled.")
				return
			}

			if err != nil {
				s.UpdateMessagef("Errored.")
				continue
			}

			s.UpdateMessagef("%s: Comparing poster checksums...", imdbId)
			metadbHash, err := hashFile(metadbPath)

			if err != nil {
				s.UpdateMessagef("Errored.")
				continue
			}

			plexHash, err := hashFile(plexPath)

			if err != nil {
				s.UpdateMessagef("Errored.")
				continue
			}

			if plexHash == metadbHash {
				continue
			}

			s.UpdateMessagef("%s: Uploading poster to Plex...", imdbId)
			f, err := os.Open(metadbPath)

			if err != nil {
				s.UpdateMessagef("Errored.")
				continue
			}

			if err = connection.UploadPoster(job.RatingKey, f); err != nil {
				s.UpdateMessagef("Errored.")
			}
		}

		s.CompleteWithMessage("Done.")
	}, cli.Threads)

	return nil
}
