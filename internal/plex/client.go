package plex

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/jyggen/go-plex-client"
	"github.com/jyggen/posterr-cli/internal/cache"
	internalhttp "github.com/jyggen/posterr-cli/internal/http"
)

type Client struct {
	cache *cache.Cache
	plex  *plex.Plex
}

func NewClient(baseUrl string, token string, client *internalhttp.Client, cacheSvc *cache.Cache) (*Client, error) {
	connection, err := plex.New(strings.TrimSuffix(baseUrl, "/"), token)
	if err != nil {
		return nil, err
	}

	connection.HTTPClient = *client.Client()

	return &Client{
		cache: cacheSvc,
		plex:  connection,
	}, nil
}

func (c *Client) CheckConnectivity() error {
	_, err := c.plex.Test()

	return err
}

func (c *Client) Libraries() (iter.Seq[*plex.Directory], error) {
	libraries, err := c.plex.GetLibraries()
	if err != nil {
		return nil, err
	}

	return func(yield func(*plex.Directory) bool) {
		for _, l := range libraries.MediaContainer.Directory {
			if !yield(&l) {
				return
			}
		}
	}, nil
}

func (c *Client) LibraryContent(libraryKey string, filters ...string) (iter.Seq[*plex.Metadata], error) {
	filter := "?includeGuids=1"

	for _, f := range filters {
		filter += fmt.Sprintf("&%s", f)
	}

	content, err := c.plex.GetLibraryContent(libraryKey, filter)
	if err != nil {
		return nil, err
	}
	return func(yield func(*plex.Metadata) bool) {
		for _, m := range content.MediaContainer.Metadata {
			if !yield(&m) {
				return
			}
		}
	}, nil
}

const thumbnailTtl = 259200 * time.Second

func (c *Client) Thumbnail(ratingKey string, thumb string) (data []byte, err error) {
	cacheKey := []byte(fmt.Sprintf("%x", sha256.Sum256([]byte(thumb))))
	data, err = c.cache.Get(cacheKey)

	if err == nil {
		return data, nil
	}

	if !errors.Is(err, badger.ErrKeyNotFound) {
		return nil, err
	}

	res, err := c.plex.GetThumbnail(ratingKey, path.Base(thumb))
	if err != nil {
		return nil, err
	}

	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	data, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return data, c.cache.SetWithExpiry(cacheKey, data, time.Now().Add(thumbnailTtl))
}

func (c *Client) UploadPoster(ratingKey string, r io.Reader) error {
	return c.plex.UploadPoster(ratingKey, r)
}
