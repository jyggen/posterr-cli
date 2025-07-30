package metadb

import (
	"context"
	"crypto/rand"
	"fmt"
	internalhttp "github.com/jyggen/posterr-cli/internal/http"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseUrl string
	client  *internalhttp.Client
}

func NewClient(baseUrl string, client *internalhttp.Client) *Client {
	return &Client{
		baseUrl: strings.TrimRight(baseUrl, "/"),
		client:  client.WithOptions(internalhttp.WithSkipRedirects()),
	}
}

func (c *Client) CheckConnectivity(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/_/%s", c.baseUrl, rand.Text()), nil)

	if err != nil {
		return err
	}

	res, err := c.client.Do(req)

	if err != nil {
		return err
	}

	if err = res.Body.Close(); err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return nil
}

func (c *Client) PosterByImdbId(ctx context.Context, imdbId string) (string, error) {
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/imdb/%s", c.baseUrl, imdbId), nil)

			if err != nil {
				return "", err
			}

			res, err := c.client.Do(req)

			if err != nil {
				return "", err
			}

			if err = res.Body.Close(); err != nil {
				return "", err
			}

			switch res.StatusCode {
			case http.StatusAccepted:
				sleep(ctx, getRetryAfter(res))
			case http.StatusServiceUnavailable:
				sleep(ctx, getRetryAfter(res))
			case http.StatusNotFound:
				return "", nil
			case http.StatusSeeOther:
				return res.Header.Get("Location"), nil
			default:
				return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
			}
		}
	}
}

const defaultSleepTime = 1 * time.Second

func getRetryAfter(res *http.Response) time.Duration {
	sleepHeader := res.Header.Get("Retry-After")

	if sleepHeader == "" {
		return defaultSleepTime
	}

	sleepSeconds, err := strconv.Atoi(sleepHeader)

	if err != nil {
		return defaultSleepTime
	}

	return time.Duration(sleepSeconds) * time.Second
}

func sleep(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(duration):
		return
	}
}
