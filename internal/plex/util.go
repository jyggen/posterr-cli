package plex

import (
	"context"
	"github.com/jyggen/go-plex-client"
	"slices"
	"strings"
)

func NewMoviesProducer(plexClient *Client) func(context.Context, chan *plex.Metadata) error {
	return func(ctx context.Context, queue chan *plex.Metadata) error {
		libraries, err := plexClient.Libraries()
		if err != nil {
			return err
		}

		for l := range libraries {
			if l.Type != "movie" {
				continue
			}

			content, innerErr := plexClient.LibraryContent(l.Key)

			if innerErr != nil {
				return innerErr
			}

			for c := range content {
				language := l.Language

				if c.LanguageOverride != "" {
					language = c.LanguageOverride
				}

				languageParts := strings.SplitN(language, "-", 2)

				if languageParts[0] != "en" {
					continue
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				case queue <- c:
				}
			}
		}

		return nil
	}
}

func ImdbID(metadata *plex.Metadata) string {
	idx := slices.IndexFunc(metadata.AltGUIDs, func(guid plex.AltGUID) bool {
		return strings.HasPrefix(guid.ID, "imdb://")
	})

	if idx == -1 {
		return ""
	}

	return metadata.AltGUIDs[idx].ID[7:]
}
