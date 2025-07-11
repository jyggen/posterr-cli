package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chelnak/ysmrr"
)

func updateMessagef(s *ysmrr.Spinner, format string, a ...interface{}) {
	if s == nil {
		return
	}

	s.UpdateMessagef(format, a...)
}

func getPosterByImdbId(ctx context.Context, client *http.Client, cacheDir string, imdbId string, s *ysmrr.Spinner) (string, error) {
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	for ctx.Err() == nil {
		updateMessagef(s, "%s: Checking MetaDB for the best poster available...", imdbId)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://posters.metadb.info/imdb/"+imdbId, nil)

		if err != nil {
			return "", fmt.Errorf("%s: %w", imdbId, err)
		}

		var res *http.Response

		res, err = client.Do(req)

		if err != nil {
			return "", fmt.Errorf("%s: %w", imdbId, err)
		}

		if err = res.Body.Close(); err != nil {
			return "", err
		}

		switch res.StatusCode {
		case http.StatusAccepted:
			var sleepTime time.Duration

			sleepTime, err = getRetryAfter(res)

			if err != nil {
				return "", fmt.Errorf("%s: %w", imdbId, err)
			}

			updateMessagef(s, "%s: Waiting for MetaDB to scour the internet for available posters...", imdbId)
			time.Sleep(sleepTime)
		case http.StatusServiceUnavailable:
			var sleepTime time.Duration

			sleepTime, err = getRetryAfter(res)

			if err != nil {
				return "", fmt.Errorf("%s: %w", imdbId, err)
			}

			updateMessagef(s, "%s: Waiting for MetaDB's servers to catch up...", imdbId)
			time.Sleep(sleepTime)
		case http.StatusNotFound:
			return "", nil
		case http.StatusSeeOther:
			updateMessagef(s, "%s: Writing poster to disk...", imdbId)
			return downloadOrCache(func(u string) (*http.Response, error) {
				req, err = http.NewRequestWithContext(ctx, http.MethodGet, u, nil)

				if err != nil {
					return nil, fmt.Errorf("%s: %w", imdbId, err)
				}

				return client.Do(req)
			}, cacheDir, res.Header.Get("Location"))
		default:
			return "", fmt.Errorf("%s: unknown error: %d", imdbId, res.StatusCode)
		}
	}

	return "", fmt.Errorf("%s: cancelled", imdbId)
}

func getRetryAfter(res *http.Response) (time.Duration, error) {
	sleepHeader := res.Header.Get("Retry-After")

	var sleepTime time.Duration

	if sleepHeader == "" {
		sleepTime = 1 * time.Second
	} else {
		sleepSeconds, err := strconv.Atoi(sleepHeader)

		if err != nil {
			return sleepTime, err
		}

		sleepTime = time.Duration(sleepSeconds) * time.Second
	}

	return sleepTime, nil
}
