package cache

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/dgraph-io/badger/v4"
	internalhttp "github.com/jyggen/posterr-cli/internal/http"
	"github.com/pquerna/cachecontrol/cacheobject"
)

func NewCachingMiddleware(cache *Cache) internalhttp.MiddlewareFunc {
	return func(next internalhttp.Middleware) internalhttp.Middleware {
		return func(req *http.Request) (*http.Response, error) {
			cacheKey, err := httputil.DumpRequestOut(req, true)
			if err != nil {
				return nil, err
			}

			cacheKey = []byte(fmt.Sprintf("%x", sha256.Sum256(cacheKey)))
			v, err := cache.Get(cacheKey)

			if err == nil {
				return http.ReadResponse(bufio.NewReader(bytes.NewBuffer(v)), req)
			}

			if !errors.Is(err, badger.ErrKeyNotFound) {
				return nil, err
			}

			res, err := next(req)
			if err != nil {
				return nil, err
			}

			reasons, expires, _, obj, err := cacheobject.UsingRequestResponseWithObject(req, res.StatusCode, res.Header, false)
			if err != nil {
				return nil, err
			}

			if len(reasons) == 0 && !obj.RespDirectives.NoCachePresent {
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
