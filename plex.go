package main

import (
	"context"
	"github.com/jyggen/go-plex-client"
	"log"
	"net/http"
	"path"
	"slices"
	"strings"
)

func getPosterByMetadata(connection *plex.Plex, cacheDir string, metadata plex.Metadata) (string, error) {
	return downloadOrCache(func(u string) (*http.Response, error) {
		return connection.GetThumbnail(metadata.RatingKey, path.Base(metadata.Thumb))
	}, cacheDir, metadata.Thumb)
}

func getImdbId(metadata plex.Metadata) string {
	idx := slices.IndexFunc(metadata.AltGUIDs, func(guid plex.AltGUID) bool {
		return strings.HasPrefix(guid.ID, "imdb://")
	})

	if idx == -1 {
		return ""
	}

	return metadata.AltGUIDs[idx].ID[7:]
}

func produceMoviesMetadata(ctx context.Context, connection *plex.Plex, queue chan plex.Metadata) {
	libraries, err := connection.GetLibraries()

	if err != nil {
		return
	}

OuterLoop:
	for _, l := range libraries.MediaContainer.Directory {
		if l.Type != "movie" {
			continue
		}

		content, err := connection.GetLibraryContent(l.Key, "?includeGuids=1")

		if err != nil {
			log.Println(err)
			continue
		}

		for _, m := range content.MediaContainer.Metadata {
			if ctx.Err() != nil {
				break OuterLoop
			}

			language := l.Language
			overrideIndex := slices.IndexFunc(m.Preferences.Setting, func(setting plex.Setting) bool {
				return setting.ID == "languageOverride"
			})

			if overrideIndex != -1 && m.Preferences.Setting[overrideIndex].Value != "" {
				language = m.Preferences.Setting[overrideIndex].Value
			}

			languageParts := strings.SplitN(language, "-", 2)

			if languageParts[0] != "en" {
				continue
			}

			queue <- m
		}
	}
}
