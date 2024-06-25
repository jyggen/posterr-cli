package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/chelnak/ysmrr"
	"github.com/jyggen/go-plex-client"
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

	client := newClient(cli.Compare.HTTPTimeout)
	connection, err := plex.New(strings.TrimSuffix(cli.Compare.PlexBaseURL.String(), "/"), cli.Compare.PlexToken)

	if err != nil {
		return err
	}

	connection.HTTPClient = *client
	b := bufio.NewWriter(outputFile)

	defer b.Flush()

	if _, err = b.WriteString(`
<!doctype html>
<html lang="en">
	<head>
		<link
			rel="stylesheet"
			href="https://cdn.jsdelivr.net/npm/bootstrap@4.1.3/dist/css/bootstrap.min.css"
			integrity="sha384-MCw98/SFnGE8fJT3GXwEOngsV7Zt27NXFoaoApmYm81iuXoPkFOJwJ8ERdknLPMO"
			crossorigin="anonymous"
		>
	</head>
	<body>
		<main role="main" class="container">
			<table class="table table-striped">
				<thead class="thead-dark">
					<tr>
						<th>IMDb ID</th>
						<th>Plex</th>
						<th>MetaDB</th>
					</tr>
				</thead>
			<tbody>
`); err != nil {
		return err
	}

	var mutex sync.Mutex

	err = withThreads(func(ctx context.Context, queue chan plex.Metadata) error {
		return produceMoviesMetadata(ctx, connection, queue)
	}, func(ctx context.Context, queue chan plex.Metadata, s *ysmrr.Spinner) error {
		for job := range queue {
			imdbID := getImdbID(job)

			if imdbID == "" {
				continue
			}

			metadbPath, err := getPosterByImdbId(ctx, client, cli.CacheBasePath, imdbID, s)

			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err != nil {
				return err
			}

			s.UpdateMessagef("%s: Downloading current poster from Plex...", imdbID)
			plexPath, err := getPosterByMetadata(connection, cli.CacheBasePath, job)

			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err != nil {
				return err
			}

			s.UpdateMessagef("%s: Comparing poster checksums...", imdbID)
			metadbHash, err := hashFile(metadbPath)

			if err != nil {
				return err
			}

			plexHash, err := hashFile(plexPath)

			if err != nil {
				return err
			}

			if plexHash == metadbHash {
				continue
			}

			s.UpdateMessagef("%s: Waiting for other threads...", imdbID)
			mutex.Lock()
			s.UpdateMessagef("%s: Writing comparison to disk...", imdbID)
			if _, err = b.WriteString(fmt.Sprintf(`
<tr>
	<td>%s</td>
	<td><img width=300 src="file://%s"></td>
	<td><img width=300 src="file://%s"></td>
</tr>
`, imdbID, plexPath, metadbPath)); err != nil {
				return err
			}
			mutex.Unlock()
		}

		return nil
	}, cli.Threads)

	_, err2 := b.WriteString("</tbody></table></main></body></html>")

	if err != nil {
		return err
	}

	return err2
}
