package preview

import (
	"context"
	"errors"
	"fmt"
	"github.com/jyggen/posterr-cli/internal/cmd"
	internalhttp "github.com/jyggen/posterr-cli/internal/http"
	"github.com/jyggen/posterr-cli/internal/metadb"
	"github.com/pkg/browser"
	"net/http"
)

type Command struct {
	cmd.CacheConfig  `embed:""`
	cmd.HTTPConfig   `embed:""`
	cmd.MetaDBConfig `embed:""`
	ImdbID           string `arg:"" name:"imdb-id" help:""`
}

func (cmd *Command) Run(ctx context.Context, httpClient *internalhttp.Client, metadbClient *metadb.Client) (err error) {
	posterUrl, err := metadbClient.PosterByImdbId(ctx, cmd.ImdbID)

	if err != nil {
		return err
	}

	if posterUrl == "" {
		return errors.New("unknown IMDb ID")
	}

	res, err := httpClient.Get(ctx, posterUrl)

	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, res.Body.Close())
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	f, err := internalhttp.DumpResponseBodyToDisk(res)

	if err != nil {
		return err
	}

	return browser.OpenFile(f)
}
