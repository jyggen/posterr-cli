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
	Cache       *cmd.CacheConfig       `embed:"" prefix:"cache-"`
	Concurrency *cmd.ConcurrencyConfig `embed:""`
	HTTP        *cmd.HTTPConfig        `embed:"" prefix:"http-"`
	Plex        *cmd.PlexConfig        `embed:"" prefix:"plex-"`
	PostersApi  *cmd.PostersApiConfig  `embed:""`
	OutputFile  string                 `arg:"" default:"-" type:"path" help:"Where to output the resulting HTML. Defaults to stdout."`
}

func (cmd *Command) Run(ctx context.Context, httpClient *http.Client, metadbClient *metadb.Client, plexClient *plex.Client) (err error) {
	var outputFile *os.File

	if cmd.OutputFile == "-" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.Create(cmd.OutputFile)
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

				s.UpdateMessage(m.RatingKey)

				if err = compareMovie(ctx, m, b, &mutex, httpClient, metadbClient, plexClient); err != nil {
					return fmt.Errorf("%s: %w", m.RatingKey, err)
				}
			}
		}
	}, cmd.Concurrency.Workers)

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
		return fmt.Errorf("unable to get poster url from metadb: %w", err)
	}

	if posterUrl == "" {
		return nil
	}

	posterrResponse, err := httpClient.Get(ctx, posterUrl)
	if err != nil {
		return fmt.Errorf("unable to download poster: %w", err)
	}

	defer func() {
		err = errors.Join(err, posterrResponse.Body.Close())
	}()

	posterrData, err := io.ReadAll(posterrResponse.Body)
	if err != nil {
		return fmt.Errorf("unable to read poster data: %w", err)
	}

	plexData, err := plexClient.Thumbnail(m.RatingKey, m.Thumb)
	if err != nil {
		return fmt.Errorf("unable to download plex poster: %w", err)
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

	if err != nil {
		return fmt.Errorf("unable to render template: %w", err)
	}

	return nil
}
