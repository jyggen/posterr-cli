package main

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"slices"
	"strings"

	"github.com/jyggen/go-plex-client"
)

func getPosterByMetadata(connection *plex.Plex, cacheDir string, metadata plex.Metadata) (string, error) {
	fileName, err := downloadOrCache(func(_ string) (*http.Response, error) {
		return connection.GetThumbnail(metadata.RatingKey, path.Base(metadata.Thumb))
	}, cacheDir, metadata.Thumb)

	if err != nil {
		return fileName, fmt.Errorf("%s: %w", metadata.RatingKey, err)
	}

	return fileName, nil
}

func getImdbID(metadata plex.Metadata) string {
	idx := slices.IndexFunc(metadata.AltGUIDs, func(guid plex.AltGUID) bool {
		return strings.HasPrefix(guid.ID, "imdb://")
	})

	if idx == -1 {
		return ""
	}

	return metadata.AltGUIDs[idx].ID[7:]
}

func produceMoviesMetadata(ctx context.Context, connection *plex.Plex, queue chan plex.Metadata) error {
	if _, err := connection.Test(); err != nil {
		return err
	}

	libraries, err := connection.GetLibraries()

	if err != nil {
		return err
	}

	for _, l := range libraries.MediaContainer.Directory {
		if l.Type != "movie" {
			continue
		}

		content, err := connection.GetLibraryContent(l.Key, "?includeGuids=1")

		if err != nil {
			return err
		}

		for _, m := range content.MediaContainer.Metadata {
			language := l.Language

			if m.LanguageOverride != "" {
				language = m.LanguageOverride
			}

			languageParts := strings.SplitN(language, "-", 2)

			if languageParts[0] != "en" {
				continue
			}

			select {
			case <-ctx.Done():
				return nil
			case queue <- m:
				continue
			}
		}
	}

	return nil
}
