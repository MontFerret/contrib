package memory

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"unicode"

	"github.com/MontFerret/contrib/modules/web/html/drivers/internal/origin"
)

var errOriginDenied = errors.New("document request denied by origin allowlist")

type originCancelKey struct{}

type originTransport struct {
	origins *origin.Allowlist
	next    http.RoundTripper
}

func newOriginTransport(origins *origin.Allowlist, next http.RoundTripper) *originTransport {
	if next == nil {
		next = http.DefaultTransport
	}

	return &originTransport{origins: origins, next: next}
}

func (t *originTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	host, allowed := t.origins.Match(req.URL)
	if allowed {
		// net/http uses different IDNA conversions for dialing and for the Host
		// header/proxy request. Send the exact ASCII name matched by the policy.
		if strings.ContainsFunc(req.URL.Host, func(r rune) bool { return r > unicode.MaxASCII }) {
			req = req.Clone(req.Context())
			if port := req.URL.Port(); port != "" {
				host = net.JoinHostPort(host, port)
			}

			req.URL.Host = host
			req.Host = host
		}

		return t.next.RoundTrip(req)
	}

	if req.Body != nil {
		_ = req.Body.Close()
	}

	ctx := req.Context()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	denial := errOriginDenied
	if req.URL.Host != "" {
		// Identify the denied hop without exposing URL credentials, paths or queries.
		denial = fmt.Errorf("%w: %s://%s", errOriginDenied, req.URL.Scheme, req.URL.Host)
	}

	if cancel, ok := ctx.Value(originCancelKey{}).(context.CancelCauseFunc); ok {
		cancel(denial)
	}

	return nil, denial
}
