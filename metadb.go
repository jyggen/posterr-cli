package main

import (
	"context"
	"fmt"
	internalhttp "github.com/jyggen/posterr-cli/internal/http"
	"github.com/jyggen/posterr-cli/internal/metadb"
	"net/http"
	"strconv"
	"time"

	"github.com/chelnak/ysmrr"
)

func newMetaDBClient(apiUrl string, dnsResolver string, timeout time.Duration) (*metadb.Client, error) {
	options := []internalhttp.Option{
		internalhttp.WithTimeout(timeout),
		internalhttp.WithUserAgent(fmt.Sprintf("posterr/%s", version)),
	}

	if apiUrl == "" {
		return metadb.NewClientFromServiceDiscovery(dnsResolver, options...)
	}

	return metadb.NewClient(apiUrl, options...), nil
}

func getPosterByImdbId(ctx context.Context, client *metadb.Client, imdbId string, s *ysmrr.Spinner) (string, error) {
	for ctx.Err() == nil {
		updateMessagef(s, "%s: Checking MetaDB for the best poster available...", imdbId)

		res, err := client.GetPosterByImdbId(ctx, imdbId)

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
			return res.Header.Get("Location"), nil
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

func updateMessagef(s *ysmrr.Spinner, format string, a ...interface{}) {
	if s == nil {
		return
	}

	s.UpdateMessagef(format, a...)
}
