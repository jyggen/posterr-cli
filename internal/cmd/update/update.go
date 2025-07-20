package update

import (
	"github.com/jyggen/posterr-cli/internal/cmd"
)

type Command struct {
	cmd.PlexConfig
	Force bool `help:""`
}

func (cmd *Command) Run() error {
	return nil
	/*client := newClient(cmd.HTTPTimeout)
	connection, err := plex.New(strings.TrimSuffix(cmd.PlexBaseURL.String(), "/"), cmd.PlexToken)

	if err != nil {
		return err
	}

	connection.HTTPClient = *client.Client()

	metadbClient, err := newMetaDBClient(cmd.ApiURL, cmd.DnsResolver, cmd.HTTPTimeout)

	if err != nil {
		return err
	}

	return withThreads(func(ctx context.Context, queue chan plex.Metadata) error {
		return produceMoviesMetadata(ctx, connection, queue)
	}, func(ctx context.Context, queue chan plex.Metadata, s *ysmrr.Spinner) error {
		for job := range queue {
			imdbID := getImdbID(job)

			if imdbID == "" {
				continue
			}

			metadbUrl, err := getPosterByImdbId(ctx, metadbClient, imdbID, s)

			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err != nil {
				return err
			}

			if metadbUrl == "" {
				continue
			}

			metadbPath, err := downloadOrCache(client.Get, cli.CacheBasePath, metadbUrl)

			if err != nil {
				return fmt.Errorf("%s: %w", cli.Preview.ImdbID, err)
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
	}, cli.Threads)*/
}
