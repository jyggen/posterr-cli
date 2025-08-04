package plex

import (
	"fmt"
	"github.com/jyggen/go-plex-client"
	"github.com/jyggen/posterr-cli/internal/http"
	"io"
	"iter"
	"path"
	"strings"
)

type Client struct {
	plex *plex.Plex
}

func NewClient(baseUrl string, token string, client *http.Client) (*Client, error) {
	connection, err := plex.New(strings.TrimSuffix(baseUrl, "/"), token)

	if err != nil {
		return nil, err
	}

	connection.HTTPClient = *client.Client()

	return &Client{
		plex: connection,
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

func (c *Client) Thumbnail(ratingKey string, thumb string) (*http.Response, error) {
	return c.plex.GetThumbnail(ratingKey, path.Base(thumb))
}

func (c *Client) UploadPoster(ratingKey string, r io.Reader) error {
	return c.plex.UploadPoster(ratingKey, r)
}
