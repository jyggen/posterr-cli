package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chelnak/ysmrr"
)

func getPosterByImdbId(ctx context.Context, client *http.Client, cacheDir string, imdbId string, s *ysmrr.Spinner) (string, error) {
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	for {
		if ctx.Err() != nil {
			break
		}

		s.UpdateMessagef("%s: Checking MetaDB for the best poster available...", imdbId)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://posters.metadb.info/imdb/"+imdbId, nil)

		if err != nil {
			return "", err
		}

		res, err := client.Do(req)

		if err != nil {
			return "", err
		}

		defer res.Body.Close()

		switch res.StatusCode {
		case http.StatusAccepted:
			sleepHeader := res.Header.Get("Retry-After")

			var sleepTime time.Duration

			if sleepHeader == "" {
				sleepTime = 1 * time.Second
			} else {
				sleepSeconds, err := strconv.Atoi(sleepHeader)

				if err != nil {
					return "", err
				}

				sleepTime = time.Duration(sleepSeconds) * time.Second
			}

			s.UpdateMessagef("%s: Waiting for MetaDB to scour the internet for available posters...", imdbId)
			time.Sleep(sleepTime)
		case http.StatusServiceUnavailable:
			sleepHeader := res.Header.Get("Retry-After")

			var sleepTime time.Duration

			if sleepHeader == "" {
				sleepTime = 1 * time.Second
			} else {
				sleepSeconds, err := strconv.Atoi(sleepHeader)

				if err != nil {
					return "", err
				}

				sleepTime = time.Duration(sleepSeconds) * time.Second
			}

			s.UpdateMessagef("%s: Waiting for MetaDB's servers to catch up...", imdbId)
			time.Sleep(sleepTime)
		case http.StatusNotFound:
			return "", errors.New("not found")
		case http.StatusSeeOther:
			s.UpdateMessagef("%s: Writing poster to disk...", imdbId)
			return downloadOrCache(func(u string) (*http.Response, error) {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)

				if err != nil {
					return nil, err
				}

				return client.Do(req)
			}, cacheDir, res.Header.Get("Location"))
		default:
			return "", fmt.Errorf("unknown error: %v", res.StatusCode)
		}
	}

	return "", errors.New("cancelled")
}
