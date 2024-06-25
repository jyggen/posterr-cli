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
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	for {
		if ctx.Err() != nil {
			break
		}

		s.UpdateMessagef("%s: Checking MetaDB for the best poster available...", imdbId)

		res, err := client.Get("https://posters.metadb.info/imdb/" + imdbId)

		if err != nil {
			return "", err
		}

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
				return client.Get(u)
			}, cacheDir, res.Header.Get("Location"))
		default:
			return "", fmt.Errorf("unknown error: %v", res.StatusCode)
		}
	}

	return "", errors.New("cancelled")
}
