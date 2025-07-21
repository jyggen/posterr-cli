package update

import (
	"bytes"
	"context"
	"errors"
	"github.com/chelnak/ysmrr"
	"github.com/jyggen/posterr-cli/internal/cmd"
	"github.com/jyggen/posterr-cli/internal/concurrency"
	"github.com/jyggen/posterr-cli/internal/http"
	"github.com/jyggen/posterr-cli/internal/metadb"
	"github.com/jyggen/posterr-cli/internal/plex"
	"io"
)

type Command struct {
	cmd.CacheConfig       `embed:""`
	cmd.ConcurrencyConfig `embed:""`
	cmd.HTTPConfig        `embed:""`
	cmd.MetaDBConfig      `embed:""`
	cmd.PlexConfig        `embed:""`
	Force                 bool `help:""`
}

func (cmd *Command) Run(ctx context.Context, httpClient *http.Client, metadbClient *metadb.Client, plexClient *plex.Client) error {
	return concurrency.WithThreads(ctx, plex.NewMoviesProducer(plexClient), func(ctx context.Context, queue chan *plex.Metadata, s *ysmrr.Spinner) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case m, ok := <-queue:
				if !ok {
					return nil
				}

				s.UpdateMessage(m.Title)

				if err := updateMovie(ctx, m, httpClient, metadbClient, plexClient); err != nil {
					return err
				}
			}
		}
	}, cmd.Threads)
}

func updateMovie(ctx context.Context, m *plex.Metadata, httpClient *http.Client, metadbClient *metadb.Client, plexClient *plex.Client) error {
	imdbId := plex.ImdbID(m)

	if imdbId == "" {
		return nil
	}

	posterUrl, err := metadbClient.PosterByImdbId(ctx, imdbId)

	if err != nil {
		return err
	}

	if posterUrl == "" {
		return nil
	}

	posterrResponse, err := httpClient.Get(ctx, posterUrl)

	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, posterrResponse.Body.Close())
	}()

	b := bytes.NewBuffer(nil)
	r := io.TeeReader(posterrResponse.Body, b)

	posterrData, err := io.ReadAll(r)

	if err != nil {
		return err
	}

	plexResponse, err := plexClient.Thumbnail(m.RatingKey, m.Thumb)

	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, plexResponse.Body.Close())
	}()

	plexData, err := io.ReadAll(plexResponse.Body)

	if err != nil {
		return err
	}

	if bytes.Equal(posterrData, plexData) {
		return nil
	}

	return plexClient.UploadPoster(m.RatingKey, b)
}
