package main

import (
	"context"
	"os"
	"strings"

	"github.com/chelnak/ysmrr"
	"github.com/jyggen/go-plex-client"
)

type updateCmd struct {
	httpConfig
	plexConfig
	Force bool `help:""`
}

func (u *updateCmd) Run(cli *posterrCli) error {
	client := newClient(cli.Update.HTTPTimeout)
	connection, err := plex.New(strings.TrimSuffix(cli.Update.PlexBaseURL.String(), "/"), cli.Update.PlexToken)

	if err != nil {
		return err
	}

	connection.HTTPClient = *client

	return withThreads(func(ctx context.Context, queue chan plex.Metadata) error {
		return produceMoviesMetadata(ctx, connection, queue)
	}, func(ctx context.Context, queue chan plex.Metadata, s *ysmrr.Spinner) error {
		for job := range queue {
			imdbID := getImdbID(job)

			if imdbID == "" {
				continue
			}

			metadbPath, err := getPosterByImdbId(ctx, client, cli.CacheBasePath, imdbID, s)

			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err != nil {
				return err
			}

			s.UpdateMessagef("%s: Downloading current poster from Plex...", imdbID)
			plexPath, err := getPosterByMetadata(connection, cli.CacheBasePath, job)

			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err != nil {
				return err
			}

			if !cli.Update.Force {
				s.UpdateMessagef("%s: Comparing poster checksums...", imdbID)
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
			}

			s.UpdateMessagef("%s: Uploading poster to Plex...", imdbID)
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
