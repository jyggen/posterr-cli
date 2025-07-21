package metadb

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/jyggen/posterr-cli/internal/http"
	"github.com/miekg/dns"
	"iter"
	"math/rand"
	"slices"
	"strings"
)

const svcbName = "posters.metadb.info."

//go:embed public.pem
var publicKey []byte

func NewClientFromServiceDiscovery(ctx context.Context, dnsResolver string, client *http.Client) (*Client, error) {
	seq, err := findRemoteService(dnsResolver)

	if err != nil {
		return nil, err
	}

	for target := range seq {
		c := NewClient(fmt.Sprintf("https://%s", target), client)

		middleware, innerErr := withPublicKeyVerification(publicKey)

		if innerErr != nil {
			return nil, innerErr
		}

		c.client = c.client.WithOptions(middleware)

		if innerErr = c.CheckConnectivity(ctx); innerErr == nil {
			return c, nil
		}
	}

	return nil, errors.New("all remote services are unavailable")
}

func withPublicKeyVerification(publicKey []byte) (http.Option, error) {
	block, _ := pem.Decode(publicKey)

	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pk, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	publicKey = pk.(ed25519.PublicKey)

	return http.WithMiddleware(func(next http.Middleware) http.Middleware {
		return func(req *http.Request) (*http.Response, error) {
			res, innerErr := next(req)

			if innerErr != nil {
				return nil, innerErr
			}

			var b bytes.Buffer

			b.Write([]byte(fmt.Sprintf("%s %s\n%s\n", req.Method, req.URL.Path, res.Status)))

			includeHeaders := strings.Split(res.Header.Get("X-Signature-Headers"), ",")
			excludeHeaders := make(map[string]bool, len(res.Header))

			for h := range res.Header {
				if slices.Contains(includeHeaders, h) {
					excludeHeaders[h] = false
				} else {
					excludeHeaders[h] = true
				}
			}

			if innerErr = res.Header.WriteSubset(&b, excludeHeaders); innerErr != nil {
				return nil, innerErr
			}

			signature, innerErr := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(res.Header.Get("X-Signature"))

			if innerErr != nil {
				return nil, innerErr
			}

			if !ed25519.Verify(publicKey, b.Bytes(), signature) {
				return nil, fmt.Errorf("invalid signature")
			}

			return res, nil
		}
	}), nil
}

func findRemoteService(dnsResolver string) (iter.Seq[string], error) {
	msg := new(dns.Msg)
	msg.Id = dns.Id()
	msg.RecursionDesired = true
	msg.Question = make([]dns.Question, 1)
	msg.Question[0] = dns.Question{Name: svcbName, Qtype: dns.TypeSVCB, Qclass: dns.ClassINET}
	client := new(dns.Client)
	res, _, err := client.Exchange(msg, dnsResolver)

	if err != nil {
		return nil, err
	}

	priorities := make([]uint16, 0, len(res.Answer))
	targets := make(map[uint16][]string)

	for _, z := range res.Answer {
		svcb := z.(*dns.SVCB)

		if !slices.Contains(priorities, svcb.Priority) {
			priorities = append(priorities, svcb.Priority)
		}

		targets[svcb.Priority] = append(targets[svcb.Priority], svcb.Target)
	}

	slices.Sort(priorities)

	return func(yield func(string) bool) {
		for _, p := range priorities {
			for len(targets[p]) > 0 {
				i := rand.Intn(len(targets[p]))
				target := targets[p][i]
				targets[p] = append(targets[p][:i], targets[p][i+1:]...)

				if !yield(strings.TrimRight(target, ".")) {
					return
				}
			}
		}
	}, nil
}
