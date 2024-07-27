package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

func downloadOrCache(download func(u string) (*http.Response, error), cacheDir string, u string) (string, error) {
	sum := fmt.Sprintf("%x", sha256.Sum256([]byte(u)))
	cacheDir = filepath.Join(cacheDir, sum[0:3], sum[3:6], sum[6:9], sum[9:12], sum[12:15], sum[15:18])
	parsed, err := url.Parse(u)

	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return "", err
	}
	fileName := filepath.Join(cacheDir, sum+path.Ext(parsed.Path))
	f, err := os.Open(fileName)

	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		resp, err := download(u)

		if err != nil {
			return "", err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("response error: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return "", err
		}

		if err = os.WriteFile(fileName, body, 0400); err != nil {
			return "", err
		}
	} else if err = f.Close(); err != nil {
		return "", err
	}

	fileName, err = filepath.Abs(fileName)

	if err != nil {
		return "", err
	}

	return fileName, nil
}
