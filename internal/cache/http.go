package cache

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/jyggen/posterr-cli/internal/http"
	"github.com/pquerna/cachecontrol"
	"net/http/httputil"
)

func NewCachingMiddleware(cache *Cache) http.MiddlewareFunc {
	return func(next http.Middleware) http.Middleware {
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

			cacheKey = []byte(fmt.Sprintf("%+x\n", h.Sum(nil)))

			fmt.Println(string(cacheKey))
			_, err = cache.Get(cacheKey)

			if err == nil {
				return next(req)
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
				v, innerErr := httputil.DumpResponse(res, true)

				if innerErr != nil {
					return nil, innerErr
				}

				if innerErr = cache.SetWithExpiry(cacheKey, v, expires); innerErr != nil {
					return nil, innerErr
				}
			}

			return res, nil
		}
	}
}
