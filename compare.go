package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/chelnak/ysmrr"
	"github.com/jyggen/go-plex-client"
	"io"
	"os"
	"strings"
	"sync"
)

type compareCmd struct {
	httpConfig
	plexConfig
	OutputFile string `help:"Defaults to stdout." type:"path" default:"-"`
}

func hashFile(file string) (string, error) {
	f, err := os.Open(file)

	if err != nil {
		return "", err
	}

	defer f.Close()

	data, err := io.ReadAll(f)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func compare(cli *posterrCli) error {
	var outputFile *os.File
	var err error

	if cli.Compare.OutputFile == "-" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.Create(cli.Compare.OutputFile)

		if err != nil {
			return err
		}

		defer outputFile.Close()
	}

	client := newClient(cli.Compare.HttpTimeout)
	connection, err := plex.New(strings.TrimSuffix(cli.Compare.PlexBaseUrl.String(), "/"), cli.Compare.PlexToken)

	if err != nil {
		return err
	}

	connection.HTTPClient = *client
	b := bufio.NewWriter(outputFile)

	defer b.Flush()

	if _, err = b.Write([]byte("<!doctype html><html lang=\"en\"><head><link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/bootstrap@4.1.3/dist/css/bootstrap.min.css\" integrity=\"sha384-MCw98/SFnGE8fJT3GXwEOngsV7Zt27NXFoaoApmYm81iuXoPkFOJwJ8ERdknLPMO\" crossorigin=\"anonymous\"></head><body><main role=\"main\" class=\"container\"><table class=\"table table-striped\"><thead class=\"thead-dark\"><tr><th>IMDb ID</th><th>Plex</th><th>MetaDB</th></tr></thead><tbody>")); err != nil {
		return err
	}

	var mutex sync.Mutex

	withThreads(func(ctx context.Context, queue chan plex.Metadata) {
		produceMoviesMetadata(ctx, connection, queue)
	}, func(ctx context.Context, queue chan plex.Metadata, s *ysmrr.Spinner) {
		for job := range queue {
			imdbId := getImdbId(job)

			if imdbId == "" {
				continue
			}

			metadbPath, err := getPosterByImdbId(ctx, client, cli.CacheBasePath, imdbId, s)

			if ctx.Err() != nil {
				s.ErrorWithMessage("Cancelled.")
				return
			}

			if err != nil {
				s.UpdateMessagef("Errored.")
				continue
			}

			s.UpdateMessagef("%s: Downloading current poster from Plex...", imdbId)
			plexPath, err := getPosterByMetadata(connection, cli.CacheBasePath, job)

			if err != nil {
				s.UpdateMessagef("Errored.")
				continue
			}

			s.UpdateMessagef("%s: Comparing poster checksums...", imdbId)
			metadbHash, err := hashFile(metadbPath)

			if err != nil {
				s.UpdateMessagef("Errored.")
				continue
			}

			plexHash, err := hashFile(metadbPath)

			if err != nil {
				s.UpdateMessagef("Errored.")
				continue
			}

			if plexHash == metadbHash {
				continue
			}

			s.UpdateMessagef("%s: Waiting for other threads...", imdbId)
			mutex.Lock()
			s.UpdateMessagef("%s: Writing comparison to disk...", imdbId)
			if _, err = b.Write([]byte(fmt.Sprintf("<tr><td>%s</td><td><img width=300 src=\"file://%s\"></td><td><img width=300 src=\"file://%s\"></td></tr>\n", imdbId, plexPath, metadbPath))); err != nil {
				s.UpdateMessagef("Errored.")
			}
			mutex.Unlock()
		}

		s.CompleteWithMessage("Done.")
	}, cli.Threads)

	_, err = b.Write([]byte("</tbody></table></main></body></html>"))

	return err
}
