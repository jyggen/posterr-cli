package preview

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jyggen/posterr-cli/internal/cmd"
	internalhttp "github.com/jyggen/posterr-cli/internal/http"
	"github.com/jyggen/posterr-cli/internal/metadb"
	"github.com/pkg/browser"
)

type Command struct {
	Cache      *cmd.CacheConfig     `embed:"" prefix:"cache-"`
	HTTP       *cmd.HTTPConfig      `embed:"" prefix:"http-"`
	PostersApi cmd.PostersApiConfig `embed:""`
	ImdbID     string               `arg:"" help:"IMDb ID of the movie to preview."`
}

func (cmd *Command) Run(ctx context.Context, httpClient *internalhttp.Client, metadbClient *metadb.Client) (err error) {
	posterUrl, err := metadbClient.PosterByImdbId(ctx, cmd.ImdbID)
	if err != nil {
		return fmt.Errorf("unable to get poster url from metadb: %w", err)
	}

	if posterUrl == "" {
		return errors.New("unknown imdb id")
	}

	res, err := httpClient.Get(ctx, posterUrl)
	if err != nil {
		return fmt.Errorf("unable to download poster: %w", err)
	}

	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	f, err := internalhttp.DumpResponseBodyToDisk(res)
	if err != nil {
		return fmt.Errorf("unable to save poster to disk: %w", err)
	}

	err = browser.OpenFile(f)
	if err != nil {
		return fmt.Errorf("unable to open browser: %w", err)
	}

	return nil
}
