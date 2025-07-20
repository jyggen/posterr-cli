package cache

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	internalhttp "github.com/jyggen/posterr-cli/internal/http"
	"github.com/pquerna/cachecontrol"
	"net/http"
	"net/http/httputil"
)

func NewCachingMiddleware(cache *Cache) internalhttp.MiddlewareFunc {
	return func(next internalhttp.Middleware) internalhttp.Middleware {
		return func(req *http.Request) (*http.Response, error) {
			cacheKey, err := httputil.DumpRequestOut(req, true)

			if err != nil {
				return nil, err
			}

			h := sha256.New()
			_, err = h.Write(cacheKey)

			if err != nil {
				return nil, err
			}

			cacheKey = []byte(fmt.Sprintf("%x", h.Sum(nil)))
			v, err := cache.Get(cacheKey)

			if err == nil {
				return http.ReadResponse(bufio.NewReader(bytes.NewBuffer(v)), req)
			}

			if !errors.Is(badger.ErrKeyNotFound, err) {
				return nil, err
			}

			res, err := next(req)

			if err != nil {
				return nil, err
			}

			reasons, expires, err := cachecontrol.CachableResponse(req, res, cachecontrol.Options{})

			if err != nil {
				return nil, err
			}

			if len(reasons) == 0 {
				v, err = httputil.DumpResponse(res, true)

				if err != nil {
					return nil, err
				}

				if err = cache.SetWithExpiry(cacheKey, v, expires); err != nil {
					return nil, err
				}
			}

			return res, nil
		}
	}
}
