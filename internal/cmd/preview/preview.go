package preview

import (
	"context"
	"github.com/jyggen/posterr-cli/internal/cmd"
	"github.com/jyggen/posterr-cli/internal/metadb"
	"github.com/pkg/browser"
)

type Command struct {
	cmd.CacheConfig  `embed:""`
	cmd.HTTPConfig   `embed:""`
	cmd.MetaDBConfig `embed:""`
	ImdbID           string `arg:"" name:"imdb-id" help:""`
}

func (cmd *Command) Run(ctx context.Context, metadbClient *metadb.Client) error {
	posterUrl, err := metadbClient.GetPosterByImdbId(ctx, cmd.ImdbID)

	if err != nil {
		return err
	}

	return browser.OpenURL(posterUrl)
}
