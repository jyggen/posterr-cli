package compare

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"sync"

	"github.com/chelnak/ysmrr"
	"github.com/jyggen/posterr-cli/internal/cmd"
	"github.com/jyggen/posterr-cli/internal/concurrency"
	"github.com/jyggen/posterr-cli/internal/http"
	"github.com/jyggen/posterr-cli/internal/metadb"
	"github.com/jyggen/posterr-cli/internal/plex"
	"github.com/vincent-petithory/dataurl"
)

//go:embed compare.gohtml
var tmplRaw string
var tmpl = template.Must(template.New("").Parse(tmplRaw))

type tmplData struct {
	ImdbId         string
	PlexDataUrl    template.URL
	PosterrDataUrl template.URL
}

type Command struct {
	cmd.CacheConfig       `embed:""`
	cmd.ConcurrencyConfig `embed:""`
	cmd.HTTPConfig        `embed:""`
	cmd.MetaDBConfig      `embed:""`
	cmd.PlexConfig        `embed:""`
	OutputFile            string `arg:"" default:"-" type:"path" help:""`
}

func (c *Command) Run(ctx context.Context, httpClient *http.Client, metadbClient *metadb.Client, plexClient *plex.Client) (err error) {
	var outputFile *os.File

	if c.OutputFile == "-" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.Create(c.OutputFile)

		if err != nil {
			return err
		}

		defer func() {
			err = errors.Join(outputFile.Close(), err)
		}()
	}

	b := bufio.NewWriter(outputFile)

	defer func() {
		err = errors.Join(b.Flush(), err)
	}()

	if err = tmpl.ExecuteTemplate(b, "prefix", nil); err != nil {
		return err
	}

	var mutex sync.Mutex

	err = concurrency.WithThreads(ctx, plex.NewMoviesProducer(plexClient), func(ctx context.Context, queue chan *plex.Metadata, s *ysmrr.Spinner) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case m, ok := <-queue:
				if !ok {
					return nil
				}

				s.UpdateMessage(m.Title)

				if err = compareMovie(ctx, m, b, &mutex, httpClient, metadbClient, plexClient); err != nil {
					return err
				}
			}
		}
	}, c.Threads)

	if err = tmpl.ExecuteTemplate(b, "suffix", nil); err != nil {
		return err
	}

	if err = b.Flush(); err != nil {
		return err
	}

	return nil
}

func compareMovie(ctx context.Context, m *plex.Metadata, b io.Writer, mutex *sync.Mutex, httpClient *http.Client, metadbClient *metadb.Client, plexClient *plex.Client) (err error) {
	imdbId := plex.ImdbID(m)

	if imdbId == "" {
		return nil
	}

	posterUrl, err := metadbClient.PosterByImdbId(ctx, imdbId)
	if err != nil {
		return err
	}

	if posterUrl == "" {
		return nil
	}

	posterrResponse, err := httpClient.Get(ctx, posterUrl)
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, posterrResponse.Body.Close())
	}()

	posterrData, err := io.ReadAll(posterrResponse.Body)
	if err != nil {
		return err
	}

	plexData, err := plexClient.Thumbnail(m.RatingKey, m.Thumb)
	if err != nil {
		return err
	}

	if bytes.Equal(posterrData, plexData) {
		return nil
	}

	mutex.Lock()

	err = tmpl.ExecuteTemplate(b, "loop", tmplData{
		ImdbId:         imdbId,
		PlexDataUrl:    template.URL(dataurl.New(plexData, "image/jpeg").String()),
		PosterrDataUrl: template.URL(dataurl.New(posterrData, posterrResponse.Header.Get("Content-Type")).String()),
	})

	mutex.Unlock()

	return err
}
