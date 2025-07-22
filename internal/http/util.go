package http

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

func DumpResponseBodyToDisk(res *http.Response) (path string, err error) {
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	var fileName string

	switch res.Header.Get("Content-Type") {
	case "image/bmp":
		fileName = "posterr_*.bmp"
	case "image/gif":
		fileName = "posterr_*.gif"
	case "image/jpeg":
		fileName = "posterr_*.jpg"
	case "image/png":
		fileName = "posterr_*.png"
	default:
		return "", fmt.Errorf("unsupported content type: %s", res.Header.Get("Content-Type"))
	}

	f, err := os.CreateTemp(os.TempDir(), fileName)

	if err != nil {
		return "", err
	}

	defer func() {
		err = errors.Join(err, f.Close())
	}()

	b := bufio.NewWriter(f)

	if _, err = b.Write(body); err != nil {
		return "", err
	}

	if err = b.Flush(); err != nil {
		return "", err
	}

	return f.Name(), nil
}
