package update

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/chelnak/ysmrr"
	"github.com/jyggen/posterr-cli/internal/cmd"
	"github.com/jyggen/posterr-cli/internal/concurrency"
	"github.com/jyggen/posterr-cli/internal/http"
	"github.com/jyggen/posterr-cli/internal/metadb"
	"github.com/jyggen/posterr-cli/internal/plex"
)

type Command struct {
	cmd.CacheConfig       `embed:""`
	cmd.ConcurrencyConfig `embed:""`
	cmd.HTTPConfig        `embed:""`
	cmd.MetaDBConfig      `embed:""`
	cmd.PlexConfig        `embed:""`
	Force                 bool `help:""`
	SinceDaysAgo          uint `help:""`
}

func (cmd *Command) Run(ctx context.Context, httpClient *http.Client, metadbClient *metadb.Client, plexClient *plex.Client) error {
	var filters []string

	if cmd.SinceDaysAgo > 0 {
		filters = append(filters, fmt.Sprintf("addedAt>>=%s", time.Now().Add(-time.Duration(cmd.SinceDaysAgo)*24*time.Hour).Format(time.DateOnly)))
	}

	return concurrency.WithThreads(ctx, plex.NewMoviesProducer(plexClient, filters...), func(ctx context.Context, queue chan *plex.Metadata, s *ysmrr.Spinner) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case m, ok := <-queue:
				if !ok {
					return nil
				}

				s.UpdateMessage(m.RatingKey)

				if err := updateMovie(ctx, m, httpClient, metadbClient, plexClient); err != nil {
					return fmt.Errorf("%s: %w", m.RatingKey, err)
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
		return fmt.Errorf("unable to get poster url from metadb: %w", err)
	}

	if posterUrl == "" {
		return nil
	}

	posterrResponse, err := httpClient.Get(ctx, posterUrl)
	if err != nil {
		return fmt.Errorf("unable to download poster: %w", err)
	}

	defer func() {
		err = errors.Join(err, posterrResponse.Body.Close())
	}()

	b := bytes.NewBuffer(nil)
	r := io.TeeReader(posterrResponse.Body, b)

	posterrData, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("unable to read poster data: %w", err)
	}

	plexData, err := plexClient.Thumbnail(m.RatingKey, m.Thumb)
	if err != nil {
		return fmt.Errorf("unable to download plex poster: %w", err)
	}

	if bytes.Equal(posterrData, plexData) {
		return nil
	}

	if err = plexClient.UploadPoster(m.RatingKey, b); err != nil {
		return fmt.Errorf("unable to upload poster to plex: %w", err)
	}

	return nil
}
